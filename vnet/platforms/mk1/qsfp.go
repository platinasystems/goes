// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mk1

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/machine"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/optics/sfp"
	fe1_platform "github.com/platinasystems/go/vnet/platforms/fe1"
)

const (
	mux0_addr           = 0x70
	mux1_addr           = 0x71
	qsfp_addr           = 0x50
	qsfp_gpio_base_addr = 0x20
	numPorts            = 32

	alphaParameter     = "/sys/module/platina_mk1/parameters/alpha"
	provisionParameter = "/sys/module/platina_mk1/parameters/provision"
)

var pub *publisher.Publisher
var firstPort bool
var maxTemp float64
var maxTempPort string
var bmcIpv6LinkLocalRedis string
var lasts = make(map[string]string)

var cached struct {
	base, provision struct {
		once sync.Once
		val  interface{}
	}
}

func PortBase() int {
	cached.base.once.Do(func() {
		base := 1
		if f, err := os.Open(alphaParameter); err == nil {
			fmt.Fscan(f, &base)
			f.Close()
			if base > 1 {
				base = 1
			} else if base < 0 {
				base = 0
			}
		} else {
			s, err := redis.Hget(machine.Name,
				"eeprom.DeviceVersion")
			if err == nil {
				var ver int
				_, err = fmt.Sscan(s, &ver)
				if err == nil && (ver == 0 || ver == 0xff) {
					base = 0
				}
			}
		}
		cached.base.val = base
	})
	return cached.base.val.(int)
}

func Provision(i int) int {
	cached.provision.once.Do(func() {
		provision := make([]int, numPorts)
		buf, err := ioutil.ReadFile(provisionParameter)
		if err == nil {
			for i, s := range strings.Split(string(buf), ",") {
				if i < numPorts {
					fmt.Sscan(s, &provision[i])
				}
			}
		}
		cached.provision.val = provision
	})
	if i < numPorts {
		return cached.provision.val.([]int)[i]
	}
	return 0
}

// set 0 vs 1-base port numbering based on HW version
// use xethPORT-SUBPORT instead of xethPORT if provision[PORT] > 0
func IfnameOf(port, subport int) string {
	if Provision(port) == 0 {
		return fmt.Sprint("xeth", port+PortBase())
	}
	return fmt.Sprint("xeth", port+PortBase(), "-", subport+PortBase())
}

func i2cMuxSelectPort(port uint) {
	// Select 2 level mux.
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (port / 8)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	i2c.Do(0, mux1_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (port % 8)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
}

func readWriteQsfp(addr uint8, b []byte, isWrite bool) (err error) {
	i, n_left := 0, len(b)

	for n_left >= 2 {
		err = i2c.Do(0, qsfp_addr, func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			if isWrite {
				d[0] = b[i+0]
				d[1] = b[i+1]
				err = bus.ReadWrite(i2c.Write, addr+uint8(i), i2c.WordData, &d)
			} else {
				err = bus.ReadWrite(i2c.Read, addr+uint8(i), i2c.WordData, &d)
				if err == nil {
					b[i+0] = d[0]
					b[i+1] = d[1]
				}
			}
			return
		})
		if err != nil {
			return
		}
		n_left -= 2
		i += 2
	}

	for n_left > 0 {
		err = i2c.Do(0, qsfp_addr, func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			if isWrite {
				d[0] = b[i+0]
				err = bus.ReadWrite(i2c.Write, addr+uint8(i), i2c.ByteData, &d)
			} else {
				err = bus.ReadWrite(i2c.Read, addr+uint8(i), i2c.ByteData, &d)
				if err == nil {
					b[i+0] = d[0]
				}
			}
			return
		})
		if err != nil {
			return
		}
		n_left -= 1
		i += 1
	}
	return
}

type qsfpStatus struct {
	// 1 => qsfp module is present
	is_present uint64
	// 1 => interrupt active
	interrupt_status uint64
}
type qsfpSignals struct {
	qsfpStatus

	// 1 => low power mode
	is_low_power_mode uint64
	// 1 => in reset; 0 not in reset
	is_reset_active uint64
}

// j == 0 => abs_l + int_l
// j == 1 => lpmode + rst_l
func readSignals(j uint) (v [2]uint32, err error) {
	err = i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (4 + j)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	if err != nil {
		return
	}
	// Read 0x20 21 22 23 to get 32 bits of status.
	for i := 0; i < 4; i++ {
		err = i2c.Do(0, qsfp_gpio_base_addr+i,
			func(bus *i2c.Bus) (err error) {
				var d i2c.SMBusData
				err = bus.Read(0, i2c.WordData, &d)
				v[i/2] |= (uint32(d[0]) | uint32(d[1])<<8) << (16 * uint(i%2))
				return
			})
	}
	return
}

const m32 = 1<<32 - 1

func (s *qsfpStatus) read() {
	v, err := readSignals(0)
	if err != nil {
		return
	}
	s.is_present = m32 &^ uint64(v[0])
	s.interrupt_status = m32 &^ uint64(v[1])
}

func (s *qsfpSignals) read() {
	s.qsfpStatus.read()
	v, err := readSignals(1)
	if err != nil {
		return
	}
	s.is_low_power_mode = uint64(v[0])
	s.is_reset_active = m32 &^ uint64(v[1])
}

func initSignals() {
	//set all ports in reset and low power mode
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (4 + 1)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	for i := 0; i < 2; i++ {
		rstBase := qsfp_gpio_base_addr + 2 + i
		for j := 0; j < 2; j++ {
			i2c.Do(0, rstBase,
				func(bus *i2c.Bus) (err error) {
					var d i2c.SMBusData
					d[0] = 0x0
					reg := uint8(2 + j)
					err = bus.Write(reg, i2c.ByteData, &d)
					return
				})
			i2c.Do(0, rstBase,
				func(bus *i2c.Bus) (err error) {
					var d i2c.SMBusData
					d[0] = 0x0
					reg := uint8(6 + j)
					err = bus.Write(reg, i2c.ByteData, &d)
					return
				})
		}
		lpBase := qsfp_gpio_base_addr + i
		for j := 0; j < 2; j++ {
			i2c.Do(0, lpBase,
				func(bus *i2c.Bus) (err error) {
					var d i2c.SMBusData
					d[0] = 0xff
					reg := uint8(2 + j)
					err = bus.Write(reg, i2c.ByteData, &d)
					return
				})
			i2c.Do(0, lpBase,
				func(bus *i2c.Bus) (err error) {
					var d i2c.SMBusData
					d[0] = 0x0
					reg := uint8(6 + j)
					err = bus.Write(reg, i2c.ByteData, &d)
					return
				})
		}
	}
}

func writeSignal(port uint, is_rst_l bool, high bool) {
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (4 + 1)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	// 0x20 0x21 for lpmode
	slave := qsfp_gpio_base_addr + int(port/16)
	if is_rst_l {
		slave += 2 // 0x22 0x23 for rst_l
	}

	var rv uint8
	i2c.Do(0, slave,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			reg := uint8(2)
			if (port/8)%2 == 1 {
				reg = 3
			}
			err = bus.Read(reg, i2c.WordData, &d)
			rv = d[0]
			return
		})
	i2c.Do(0, slave,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			if !high {
				d[0] = rv & (0xff ^ uint8(1<<(port%8)))
			} else {
				d[0] = rv | uint8(1<<(port%8))
			}
			reg := uint8(2)
			if (port/8)%2 == 1 {
				reg = 3
			}
			err = bus.Write(reg, i2c.ByteData, &d)
			return
		})
	i2c.Do(0, slave,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			reg := uint8(6)
			if (port/8)%2 == 1 {
				reg = 7
			}
			err = bus.Read(reg, i2c.WordData, &d)
			rv = d[0]
			return
		})
	i2c.Do(0, slave,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			d[0] = rv & (0xff ^ uint8(1<<(port%8)))
			reg := uint8(6)
			if (port/8)%2 == 1 {
				reg = 7
			}
			err = bus.Write(reg, i2c.ByteData, &d)
			return
		})
	return
}

type qsfpState struct {
}

func (m *qsfpMain) signalChange(signal sfp.QsfpSignal, changedPorts, newValues uint64) {

	// if qsfps are installed, take them out of reset and low power mode
	if signal == sfp.QsfpModuleIsPresent && ((changedPorts & newValues) != 0) {
		elib.Word(changedPorts).ForeachSetBit(func(i uint) {
			port := i ^ 1 // mk1 port swapping
			mod := &m.module_by_port[port]
			v := newValues&(1<<i) != 0
			if v {
				mod.SfpReset(false)
				mod.SfpSetLowPowerMode(false)
			}
		})
	}

	elib.Word(changedPorts).ForeachSetBit(func(i uint) {
		port := int(i ^ 1) // mk1 port swapping
		mod := &m.module_by_port[port]
		v := newValues&(1<<i) != 0
		q := &mod.q

		q.SetSignal(signal, v)
		if signal == sfp.QsfpModuleIsPresent {
			f := "port-" + strconv.Itoa(int(port)+PortBase()) + ".qsfp.installed"
			s := strconv.FormatBool(v)
			if s != lasts[f] {
				pub.Print(f, ": ", s)
				lasts[f] = s
			}

			// if qsfps are installed, set interface per compliance to bring dataplane up
			if v {
				// ~800ms delay is needed for hget when Goes first starts
				if firstPort {
					ifname := IfnameOf(port, 0)
					start := time.Now()
					_, err := redis.Hget(machine.Name, "vnet."+ifname+".speed")
					for err != nil {
						if time.Since(start) >= 2*time.Second {
							log.Print("hget timeout: ", err)
							break
						}
						_, err = redis.Hget(machine.Name, "vnet."+ifname+".speed")
						//ignore this value, just used to check if redis and vnet are up
					}
				}
				firstPort = false

				for i := 0; i < 4; i++ {
					ifname := IfnameOf(port, i)
					if ifname == "" {
						continue
					}

					speed, err := redis.Hget(machine.Name, "vnet."+ifname+".speed")
					if err != nil {
						continue
					} else {
						speed = strings.ToLower(speed)
						// if qsfp is copper, set media to copper and fec according to speed
						if strings.Contains(q.Ident.Compliance, "CR") {
							if speed == "100g" {
								//100g, cl91 needs to be enabled per ieee spec
								//100g should take up all 4 lanes so setting apply to first lane only
								redis.Hset(machine.Name, "vnet."+ifname+".fec", "cl91")
							} else {
								fec, err := redis.Hget(machine.Name, "vnet."+ifname+".fec")
								if err != nil {
									fmt.Printf("qsfp.go signalChange error getting fec %v\n", err)
									continue
								}
								fec = strings.ToLower(fec)
								if (speed == "40g") || (speed == "20g") || (speed == "10g") { //none or cl74 are valid, default to none if neither
									if (fec != "cl74") && (fec != "none") {
										redis.Hset(machine.Name, "vnet."+ifname+".fec", "none")
									}
								}
								if speed == "1g" { //only none is valid
									if fec != "none" {
										redis.Hset(machine.Name, "vnet."+ifname+".fec", "none")
									}
								}
								//50g, 25g can accept none, cl74, or cl94(gen2 fe1 only)
							}
							// if do not match any above, leave fec alone to what was manually configured and do not change

							//set media to copper triggers link training and should be done after fec setting
							//training will cause remote side to align phase even if tx FIR setting do not change
							{
								media, err := redis.Hget(machine.Name, "vnet."+ifname+".media")
								if err != nil {
									fmt.Printf("qsfp.go signalChange error getting media %v\n", err)
									continue
								}
								media = strings.ToLower(media)
								if media != "copper" {
									redis.Hset(machine.Name, "vnet."+ifname+".media", "copper")
								}
							}
						} else if i == 0 {
							// not copper (i.e. no "CR" in the compliance string)
							// these optics detection are for 4-lane optical module; therefore setting apply to first lane only
							if strings.Contains(q.Ident.Compliance, "40G") {
								redis.Hset(machine.Name, "vnet."+ifname+".speed", "40g")
								redis.Hset(machine.Name, "vnet."+ifname+".media", "fiber")
								redis.Hset(machine.Name, "vnet."+ifname+".fec", "none")
							} else if strings.Contains(q.Ident.Compliance, "100GBASE-SR4") {
								redis.Hset(machine.Name, "vnet."+ifname+".speed", "100g")
								redis.Hset(machine.Name, "vnet."+ifname+".media", "fiber")
								redis.Hset(machine.Name, "vnet."+ifname+".fec", "cl91")
							} else if strings.Contains(q.Ident.Compliance, "100G") {
								redis.Hset(machine.Name, "vnet."+ifname+".speed", "100g")
								redis.Hset(machine.Name, "vnet."+ifname+".media", "fiber")
								redis.Hset(machine.Name, "vnet."+ifname+".fec", "none")
							}
							// if not above optics, leave speed/media/fec config alone to what was manually configured and do not change
						}
						// if not above then not recognized module or speed, leave speed/media/fec config alone to what was manually configured and do not change
					}
				}
			}
		}
	})

	elib.Word(changedPorts).ForeachSetBit(func(i uint) {
		port := i ^ 1 // mk1 port swapping
		mod := &m.module_by_port[port]
		v := newValues&(1<<i) != 0
		q := &mod.q

		// publish or delete qsfp fields to redis on installation or removal
		if signal == sfp.QsfpModuleIsPresent {
			if v {
				// fetch and publish static identification fields
				s := q.String()
				log.Print("port ", port+uint(PortBase()), " installed: ", s)

				for _, k := range sfp.StaticRedisFields {
					f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
					if strings.Contains(k, "vendor") {
						s := q.Ident.Vendor
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
					if strings.Contains(k, "compliance") {
						s := q.Ident.Compliance
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
					if strings.Contains(k, "partnumber") {
						s := q.Ident.PartNumber
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
					if strings.Contains(k, "serialnumber") {
						s := q.Ident.SerialNumber
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
					if strings.Contains(k, "qsfp.id") {
						s := q.Ident.Id
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
					if strings.Contains(k, "qsfp.connectortype") {
						s := q.Ident.ConnectorType
						if s != lasts[f] {
							pub.Print(f, ": ", s)
							lasts[f] = s
						}
						continue
					}
				}
				// if qsfp is an optic publish static monitoring thresholds
				if !strings.Contains(q.Ident.Compliance, "CR") && q.Ident.Compliance != "" {
					// enable laser
					q.TxEnable(0xf, 0xf)

					q.Monitoring()
					for _, k := range sfp.StaticMonitoringRedisFields {
						f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
						if strings.Contains(k, "temperature") {
							if strings.Contains(k, "highAlarm") {
								s := strconv.FormatFloat(q.Config.TemperatureInCelsius.Alarm.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowAlarm") {
								s := strconv.FormatFloat(q.Config.TemperatureInCelsius.Alarm.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "highWarn") {
								s := strconv.FormatFloat(q.Config.TemperatureInCelsius.Warning.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowWarn") {
								s := strconv.FormatFloat(q.Config.TemperatureInCelsius.Warning.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}
						if strings.Contains(k, "rx.power") {
							if strings.Contains(k, "highAlarm") {
								s := strconv.FormatFloat(q.Config.RxPowerInWatts.Alarm.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowAlarm") {
								s := strconv.FormatFloat(q.Config.RxPowerInWatts.Alarm.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "highWarn") {
								s := strconv.FormatFloat(q.Config.RxPowerInWatts.Warning.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowWarn") {
								s := strconv.FormatFloat(q.Config.RxPowerInWatts.Warning.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}
						if strings.Contains(k, "tx.bias") {
							if strings.Contains(k, "highAlarm") {
								s := strconv.FormatFloat(q.Config.TxBiasCurrentInAmps.Alarm.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowAlarm") {
								s := strconv.FormatFloat(q.Config.TxBiasCurrentInAmps.Alarm.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "highWarn") {
								s := strconv.FormatFloat(q.Config.TxBiasCurrentInAmps.Warning.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowWarn") {
								s := strconv.FormatFloat(q.Config.TxBiasCurrentInAmps.Warning.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}
						if strings.Contains(k, "tx.power") {
							if strings.Contains(k, "highAlarm") {
								s := strconv.FormatFloat(q.Config.TxPowerInWatts.Alarm.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowAlarm") {
								s := strconv.FormatFloat(q.Config.TxPowerInWatts.Alarm.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "highWarn") {
								s := strconv.FormatFloat(q.Config.TxPowerInWatts.Warning.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowWarn") {
								s := strconv.FormatFloat(q.Config.TxPowerInWatts.Warning.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}
						if strings.Contains(k, "vcc") {
							if strings.Contains(k, "highAlarm") {
								s := strconv.FormatFloat(q.Config.SupplyVoltageInVolts.Alarm.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowAlarm") {
								s := strconv.FormatFloat(q.Config.SupplyVoltageInVolts.Alarm.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "highWarn") {
								s := strconv.FormatFloat(q.Config.SupplyVoltageInVolts.Warning.Hi, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "lowWarn") {
								s := strconv.FormatFloat(q.Config.SupplyVoltageInVolts.Warning.Lo, 'f', 3, 64)
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}
					}
				}
			} else {
				//qsfp has been removed
				log.Print("port ", port+uint(PortBase()), " QSFP removed")
				if maxTempPort == strconv.Itoa(int(port)) {
					maxTempPort = "-1"
				}
				// delete redis fields
				for _, k := range sfp.StaticRedisFields {
					f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
					pub.Print("delete: ", f)
					lasts[f] = ""
				}
				if !strings.Contains(q.Ident.Compliance, "CR") && q.Ident.Compliance != "" {
					for _, k := range sfp.StaticMonitoringRedisFields {
						f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
						pub.Print("delete: ", f)
						lasts[f] = ""
					}
					for _, k := range sfp.DynamicMonitoringRedisFields {
						f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
						pub.Print("delete: ", f)
						lasts[f] = ""
					}
				}
				//enable reset and low power mode
				mod.SfpReset(true)
				mod.SfpSetLowPowerMode(true)
			}
		}
	})
}

func (m *qsfpMain) poll() {
	var err error
	firstPort = true
	pub, err = publisher.New()
	if err != nil {
		log.Print("publisher.New() error: ", err)
	}

	// publish all ports empty
	for i := 0; i < numPorts; i++ {
		k := "port-" + strconv.Itoa(i+PortBase()) + ".qsfp.installed"
		s := "false"
		pub.Print(k, ": ", s)
		lasts[k] = s
	}

	sequence := 0
	for {
		old := m.current
		// Read initial state only first time; else read just status (presence + interrupt status).
		if sequence == 0 {
			m.current.read()
		} else {
			m.current.qsfpStatus.read()
		}
		new := m.current
		// Do lpmode/reset first; presence next; interrupt status last.
		// Presence change will have correct reset state when sequence == 0.
		/* LPMode and Reset are output signals
		if sequence == 0 {
			if d := new.is_low_power_mode ^ old.is_low_power_mode; d != 0 {
				m.signalChange(sfp.QsfpLowPowerMode, d, new.is_low_power_mode)
			}
			if d := new.is_reset_active ^ old.is_reset_active; d != 0 {
				m.signalChange(sfp.QsfpResetIsActive, d, new.is_reset_active)
			}
		}
		*/
		if d := new.is_present ^ old.is_present; d != 0 {
			m.signalChange(sfp.QsfpModuleIsPresent, d, new.is_present)
		}
		if d := new.interrupt_status ^ old.interrupt_status; d != 0 {
			m.signalChange(sfp.QsfpInterruptStatus, d, new.interrupt_status)
		}
		// if qsfp is present and is optic poll monitoring fields every 5 seconds
		if sequence%5 == 0 {
			for i := 0; i < numPorts; i++ {
				port := i ^ 1
				mod := &m.module_by_port[port]
				q := &mod.q
				// if qsfp is present and is optic poll monitoring fields
				if q.AllEepromValid {
					if !strings.Contains(q.Ident.Compliance, "CR") && q.Ident.Compliance != "" {
						q.Monitoring()
						for _, k := range sfp.DynamicMonitoringRedisFields {
							f := "port-" + strconv.Itoa(int(port)+PortBase()) + "." + k
							if strings.Contains(k, "qsfp.temperature.units.C") {
								s := q.Mon.Temperature
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
								update := false
								u, err := strconv.ParseFloat(q.Mon.Temperature, 64)

								if err == nil {
									if maxTempPort == "-1" {
										maxTemp = u
										maxTempPort = strconv.Itoa(int(port))
										update = true
									} else if maxTempPort == strconv.Itoa(int(port)) {
										maxTemp = u
										update = true
									} else if u > maxTemp || bmcIpv6LinkLocalRedis == "" {
										maxTemp = u
										maxTempPort = strconv.Itoa(int(port))
										update = true
									}

									if update {
										if bmcIpv6LinkLocalRedis == "" {
											m, err := redis.Hget(machine.Name, "eeprom.BaseEthernetAddress")
											if err == nil {
												o := strings.Split(m, ":")
												b, _ := hex.DecodeString(o[0])
												b[0] = b[0] ^ byte(2)
												o[0] = hex.EncodeToString(b)
												bmcIpv6LinkLocalRedis = "[fe80::" + o[0] + o[1] + ":" + o[2] + "ff:fe" + o[3] + ":" + o[4] + o[5] + "%eth0]:6379"
											}
										}
										if bmcIpv6LinkLocalRedis != "" {
											d, err := redigo.Dial("tcp", bmcIpv6LinkLocalRedis)
											if err == nil {
												d.Do("HSET", machine.Name, "qsfp.temp.units.C", q.Mon.Temperature)
												d.Close()
											}
										}

									}
								}

							}
							if strings.Contains(k, "qsfp.vcc.units.V") {
								s := q.Mon.Voltage
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.rx1.power.units.mW") {
								s := q.Mon.RxPower[0]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.rx2.power.units.mW") {
								s := q.Mon.RxPower[1]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.rx3.power.units.mW") {
								s := q.Mon.RxPower[2]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.rx4.power.units.mW") {
								s := q.Mon.RxPower[3]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx1.power.units.mW") {
								s := q.Mon.TxPower[0]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx2.power.units.mW") {
								s := q.Mon.TxPower[1]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx3.power.units.mW") {
								s := q.Mon.TxPower[2]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx4.power.units.mW") {
								s := q.Mon.TxPower[3]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx1.bias.units.mA") {
								s := q.Mon.TxBias[0]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx2.bias.units.mA") {
								s := q.Mon.TxBias[1]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx3.bias.units.mA") {
								s := q.Mon.TxBias[2]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.tx4.bias.units.mA") {
								s := q.Mon.TxBias[3]
								if s != lasts[f] {
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.alarms.module") {
								s := q.Alarms.Module
								if s == "" {
									s = "none"
								}
								if s != lasts[f] {
									if lasts[f] != "" && s != "none" {
										log.Print("warning: port-" + strconv.Itoa(int(port)+PortBase()) + " qsfp module alarm: " + s)
									}
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
							if strings.Contains(k, "qsfp.alarms.channels") {
								s := q.Alarms.Channels
								if s == "" {
									s = "none"
								}
								if s != lasts[f] {
									if lasts[f] != "" && s != "none" {
										log.Print("warning: port-" + strconv.Itoa(int(port)+PortBase()) + " qsfp channel alarm: " + s)
									}
									pub.Print(f, ": ", s)
									lasts[f] = s
								}
							}
						}

					}
				}
			}
		}
		sequence++
		time.Sleep(1 * time.Second)
	}
}

type qsfpModule struct {
	// Index into m.current.* bitmaps.
	port_index uint
	m          *qsfpMain
	q          sfp.QsfpModule
}

type qsfpMain struct {
	current        qsfpSignals
	module_by_port []qsfpModule
}

func (q *qsfpModule) SfpReset(is_active bool) {
	if is_active {
		writeSignal(q.port_index, true, false)
	} else {
		writeSignal(q.port_index, true, true)
	}
}
func (q *qsfpModule) SfpSetLowPowerMode(is_active bool) {
	if is_active {
		writeSignal(q.port_index, false, true)
	} else {
		writeSignal(q.port_index, false, false)
	}
}
func (q *qsfpModule) SfpReadWrite(offset uint, p []uint8, isWrite bool) (write_ok bool) {
	i2cMuxSelectPort(q.port_index)
	err := readWriteQsfp(uint8(offset), p, isWrite)
	if write_ok = err == nil; !write_ok {
		if errno, ok := err.(syscall.Errno); !ok || errno != syscall.ENXIO {
			panic(err)
		}
	}
	return
}

func qsfpInit(v *vnet.Vnet, p *fe1_platform.Platform) {
	maxTemp = 0
	maxTempPort = "-1"

	m := &qsfpMain{}

	p.QsfpModules = make(map[fe1_platform.SwitchPort]*sfp.QsfpModule)
	m.module_by_port = make([]qsfpModule, 32)
	for port := range m.module_by_port {
		q := &m.module_by_port[port]
		q.port_index = uint(port ^ 1)
		q.m = m
		q.q.Init(q)
		sp := fe1_platform.SwitchPort{Switch: 0, Port: uint8(port)}
		p.QsfpModules[sp] = &q.q
	}
	initSignals()
	go m.poll()
}
