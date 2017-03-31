// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qsfp

import (
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "qsfp"

type I2cDev struct {
	Bus       int
	Addr      int
	MuxBus    int
	MuxAddr   int
	MuxValue  int
	MuxBus2   int
	MuxAddr2  int
	MuxValue2 int
}

var Init = func() {}
var once sync.Once

var Vdev [32]I2cDev

var portLpage0 [32]lpage0
var portUpage3 [32]upage3
var portIsCopper [32]bool

var VpageByKey map[string]uint8

var latestPresent = [2]uint16{0xffff, 0xffff}
var present = [2]uint16{0xffff, 0xffff}

type cmd struct {
	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint8
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t
	var err error

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]float64)
	cmd.lasts = make(map[string]string)
	cmd.lastu = make(map[string]uint8)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}
	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	go qsfpioTicker(cmd)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if err = cmd.update(); err != nil {
				close(cmd.stop)
				return err
			}
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	close(cmd.stop)
	return nil
}

func (cmd *cmd) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}

	for j := 0; j < 2; j++ {
		//when qsfp is installed or removed from a port
		if present[j] != latestPresent[j] {
			for i := 0; i < 16; i++ {
				if (1<<uint(i))&(latestPresent[j]^present[j]) != 0 {
					//physical to logical port translation
					lp := i + j*16
					if (lp % 2) == 0 {
						lp += 2
					}
					var typeString string
					if ((1 << uint(i)) & (latestPresent[j] ^ 0xffff)) != 0 {
						//when qsfp is installed publish static data
						k := "port-" + strconv.Itoa(lp) + ".qsfp.compliance"
						v := Vdev[i+j*16].Compliance()
						var portConfig string

						//identify copper vs optic and set media and speed
						ready, err := redis.Hget(redis.DefaultHash, "vnet.ready")
						if err == nil && ready == "true" {
							media, err := redis.Hget(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.media")
							if err != nil {
								log.Print("qsfp hget error:", err)
							}
							speed, err := redis.Hget(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.speed")
							if err != nil {
								log.Print("qsfp hget error:", err)
							}
							if strings.Contains(v, "-CR") {
								portIsCopper[i+j*16] = true
								if media != "copper" {
									ret, err := redis.Hset(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.media", "copper")
									if err != nil || ret != 1 {
										log.Print("qsfp hset error:", err, " ", ret)
									} else {
										portConfig += "copper "
									}
								}
							} else if strings.Contains(v, "40G") {
								portIsCopper[i+j*16] = false
								if media != "fiber" {
									ret, err := redis.Hset(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.media", "fiber")
									if err != nil || ret != 1 {
										log.Print("qsfp hset error:", err, " ", ret)
									} else {
										portConfig += "fiber "
									}
								}
								if speed != "40g" {
									ret, err := redis.Hset(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.speed", "40g")
									if err != nil || ret != 1 {
										log.Print("qsfp hset error:", err, " ", ret)
									} else {
										portConfig += "40g fixed speed"
									}
								}
							} else {
								portIsCopper[i+j*16] = false
								if media != "fiber" {
									ret, err := redis.Hset(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.media", "fiber")
									if err != nil || ret != 1 {
										log.Print("qsfp hset error:", err, " ", ret)
									} else {
										portConfig += "fiber "
									}
								}
								if speed != "100g" {
									ret, err := redis.Hset(redis.DefaultHash, "vnet.eth-"+strconv.Itoa(lp)+"-1.speed", "100g")
									if err != nil || ret != 1 {
										log.Print("qsfp hset error:", err, " ", ret)
									} else {
										portConfig += "100g fixed speed"
									}
								}
							}
						}
						typeString += strings.Trim(v, " ") + ", "
						if v != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", v)
							cmd.lasts[k] = v
						}
						k = "port-" + strconv.Itoa(lp) + ".qsfp.vendor"
						v = Vdev[i+j*16].Vendor()
						typeString += strings.Trim(v, " ") + ", "
						if v != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", v)
							cmd.lasts[k] = v
						}
						k = "port-" + strconv.Itoa(lp) + ".qsfp.partnumber"
						v = Vdev[i+j*16].PN()
						typeString += strings.Trim(v, " ") + ", "
						if v != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", v)
							cmd.lasts[k] = v
						}
						k = "port-" + strconv.Itoa(lp) + ".qsfp.serialnumber"
						v = Vdev[i+j*16].SN()
						typeString += strings.Trim(v, " ")
						if v != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", v)
						}
						if !portIsCopper[i+j*16] {
							// get monitoring thresholds if qsfp is not a cable
							Vdev[i+j*16].StaticBlocks(i + j*16)
							v = Temp(portUpage3[i+j*16].tempHighAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.temperature.highAlarmThreshold.units.C"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Temp(portUpage3[i+j*16].tempLowAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.temperature.lowAlarmThreshold.units.C"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Temp(portUpage3[i+j*16].tempHighWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.temperature.highWarnThreshold.units.C"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Temp(portUpage3[i+j*16].tempLowWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.temperature.lowWarnThreshold.units.C"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Voltage(portUpage3[i+j*16].vccHighAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.vcc.highAlarmThreshold.units.V"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Voltage(portUpage3[i+j*16].vccLowAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.vcc.lowAlarmThreshold.units.V"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Voltage(portUpage3[i+j*16].vccHighWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.vcc.highWarnThreshold.units.V"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Voltage(portUpage3[i+j*16].vccLowWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.vcc.lowWarnThreshold.units.V"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].rxPowerHighAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.rx.power.highAlarmThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].rxPowerLowAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.rx.power.lowAlarmThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].rxPowerHighWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.rx.power.highWarnThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].rxPowerLowWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.rx.power.lowWarnThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].txPowerHighAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.power.highAlarmThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].txPowerLowAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.power.lowAlarmThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].txPowerHighWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.power.highWarnThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = Power(portUpage3[i+j*16].txPowerLowWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.power.lowWarnThreshold.units.mW"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = TxBias(portUpage3[i+j*16].txBiasHighAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.biasHighAlarmThreshold.units.mA"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = TxBias(portUpage3[i+j*16].txBiasLowAlarm)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.biasLowAlarmThreshold.units.mA"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = TxBias(portUpage3[i+j*16].txBiasHighWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.biasHighWarnThreshold.units.mA"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
							v = TxBias(portUpage3[i+j*16].txBiasLowWarning)
							k = "port-" + strconv.Itoa(lp) + ".qsfp.tx.biasLowWarnThreshold.units.mA"
							if v != cmd.lasts[k] {
								cmd.pub.Print(k, ": ", v)
								cmd.lasts[k] = v
							}
						}
						log.Print("QSFP detected in port ", lp, ": ", typeString)
						if portConfig != "" {
							log.Print("Port ", lp, " setting changed to ", portConfig)
						}
					} else {
						//when qsfp is removed, delete associated fields
						for _, v := range redisFields {
							k := "port-" + strconv.Itoa(lp) + "." + v
							cmd.pub.Print("delete: ", k)
							cmd.lasts[k] = ""
						}
						log.Print("QSFP removed from port ", lp)
						portIsCopper[lp-1] = true
					}
				}
			}
		}
		present[j] = latestPresent[j]
	}
	for i := 0; i < 32; i++ {
		//publish dynamic monitoring data
		var port int
		if (i % 2) == 0 {
			port = i + 2
		} else {
			port = i
		}
		if present[i/16]&(1<<uint(i%16)) == 0 {
			if !portIsCopper[i] {
				// get monitoring data only if qsfp is present and not a cable
				if Vdev[i].DataReady() {
					Vdev[i].DynamicBlocks(i)
					k := "port-" + strconv.Itoa(port) + ".qsfp.temperature.units.C"
					v := Temp(portLpage0[i].freeMonTemp)
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
					k = "port-" + strconv.Itoa(port) + ".qsfp.vcc.units.V"
					v = Voltage(portLpage0[i].freeMonVoltage)
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
					va := LanePower(portLpage0[i].rxPower)
					for x := 0; x < 4; x++ {
						k = "port-" + strconv.Itoa(port) + ".qsfp.rx" + strconv.Itoa(x+1) + ".power.units.mW"
						if va[x] != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", va[x])
							cmd.lasts[k] = va[x]
						}
					}
					va = LanePower(portLpage0[i].txPower)
					for x := 0; x < 4; x++ {
						k = "port-" + strconv.Itoa(port) + ".qsfp.tx" + strconv.Itoa(x+1) + ".power.units.mW"
						if va[x] != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", va[x])
							cmd.lasts[k] = va[x]
						}
					}
					va = LanesTxBias(portLpage0[i].txBias)
					for x := 0; x < 4; x++ {
						k = "port-" + strconv.Itoa(port) + ".qsfp.tx" + strconv.Itoa(x+1) + ".bias.units.mA"
						if va[x] != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", va[x])
							cmd.lasts[k] = va[x]
						}
					}
					vs := ChannelAlarms(portLpage0[i].channelStatusInterrupt, portLpage0[i].channelMonitorInterruptFlags)
					for x := 0; x < 4; x++ {
						k = "port-" + strconv.Itoa(port) + ".qsfp.rx" + strconv.Itoa(x+1) + ".alarms"
						if vs[x] != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", vs[x])
							cmd.lasts[k] = vs[x]
						}
					}
					for x := 4; x < 8; x++ {
						k = "port-" + strconv.Itoa(port) + ".qsfp.tx" + strconv.Itoa(x-3) + ".alarms"
						if vs[x] != cmd.lasts[k] {
							cmd.pub.Print(k, ": ", vs[x])
							cmd.lasts[k] = vs[x]
						}
					}
					vs[0] = FreeSideAlarms(portLpage0[i].freeMonitorInterruptFlags)
					k = "port-" + strconv.Itoa(port) + ".qsfp.alarms"
					if vs[0] != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", vs[0])
						cmd.lasts[k] = vs[0]
					}
				}
			}
		}

	}

	return nil
}
func (h *I2cDev) DataReady() bool {
	var t bool
	r := getRegsLpage0()

	r.status.get(h)
	DoI2cRpc()

	if (s[2].D[1] & 0x1) == 1 {
		t = false
	} else {
		t = true
	}

	return t
}
func (h *I2cDev) Compliance() string {
	r := getRegsUpage0()

	r.SpecCompliance.get(h)
	DoI2cRpc()
	cp := s[2].D[0]

	r.ExtSpecCompliance.get(h)
	DoI2cRpc()
	ecp := s[2].D[0]

	var t string
	if (cp & 0x80) != 0x80 {
		t = specComplianceValues[cp]
	} else {
		t = extSpecComplianceValues[ecp]
	}
	return t
}

func (h *I2cDev) Vendor() string {
	r := getRegsUpage0()
	r.VendorName.get(h, 16)
	DoI2cRpc()
	t := string(s[2].D[1:16])

	return t
}

func (h *I2cDev) PN() string {
	r := getRegsUpage0()
	r.VendorPN.get(h, 16)
	DoI2cRpc()
	t := string(s[2].D[1:16])

	return t
}

func (h *I2cDev) SN() string {
	r := getRegsUpage0()
	r.VendorSN.get(h, 16)
	DoI2cRpc()
	t := string(s[2].D[1:16])

	return t
}

func LanePower(t [8]byte) [4]string {
	var v [4]string
	var u uint16
	for i := 0; i < 4; i++ {
		u = uint16(t[i*2])<<8 + uint16(t[i*2+1])
		v[i] = strconv.FormatFloat(float64(u)*0.0001, 'f', 3, 64)
	}
	return v
}

func Power(t uint16) string {
	v := strconv.FormatFloat(float64(t)*0.0001, 'f', 3, 64)
	return v
}

func LanesTxBias(t [8]byte) [4]string {
	var v [4]string
	var u uint16
	for i := 0; i < 4; i++ {
		u = uint16(t[i*2])<<8 + uint16(t[i*2+1])
		v[i] = strconv.FormatFloat(float64(u)*0.002, 'f', 3, 64)
	}
	return v
}

func TxBias(t uint16) string {
	v := strconv.FormatFloat(float64(t)*0.002, 'f', 3, 64)
	return v
}

func FreeSideAlarms(t [3]byte) string {
	var v string
	if ((1 << 4) & t[0]) != 0 {
		v += "TempLowWarn,"
	}
	if ((1 << 5) & t[0]) != 0 {
		v += "TempHighWarn,"
	}
	if ((1 << 6) & t[0]) != 0 {
		v += "TempLowAlarm,"
	}
	if ((1 << 7) & t[0]) != 0 {
		v += "TempHighAlarm,"
	}
	if ((1 << 4) & t[1]) != 0 {
		v += "VccLowWarn,"
	}
	if ((1 << 5) & t[1]) != 0 {
		v += "VccHighWarn,"
	}
	if ((1 << 6) & t[1]) != 0 {
		v += "VccLowAlarm,"
	}
	if ((1 << 7) & t[1]) != 0 {
		v += "VccHighAlarm,"
	}
	if v == "" {
		v = "none"
	}
	v = strings.Trim(v, ",")
	return v
}

func ChannelAlarms(t [3]byte, w [6]byte) [8]string {
	var v [8]string
	for i := uint(0); i < 4; i++ {
		if ((1 << i) & t[0]) != 0 {
			v[i] += "RxLos,"
		}
		if ((1 << (i + 4)) & t[0]) != 0 {
			v[i+4] += "TxLos,"
		}
		if ((1 << i) & t[1]) != 0 {
			v[i+4] += "TxFault,"
		}
		if ((1 << (i + 4)) & t[1]) != 0 {
			v[i+4] += "TxEqFault,"
		}
		if ((1 << i) & t[2]) != 0 {
			v[i] += "RxCdrLol,"
		}
		if ((1 << (i + 4)) & t[2]) != 0 {
			v[i+4] += "TxCdrLol,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerHighAlarm,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasHighAlarm,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerHighAlarm,"
		}
		if v[i] == "" {
			v[i] = "none"
		}
		if v[i+4] == "" {
			v[i+4] = "none"
		}
		v[i] = strings.Trim(v[i], ",")
		v[i+4] = strings.Trim(v[i+4], ",")
	}
	return v
}

func Temp(t uint16) string {
	var u float64
	var v string
	if (t & 0x8000) != 0 {
		u = float64((t^0xffff)+1) / 256 * (-1)
	} else {
		u = float64(t) / 256
	}
	v = strconv.FormatFloat(u, 'f', 1, 64)
	return v
}

func Voltage(t uint16) string {
	var u float64
	var v string
	u = float64(t) * 0.0001
	v = strconv.FormatFloat(u, 'f', 3, 64)
	return v
}
func (h *I2cDev) DynamicBlocks(port int) {
	r := getRegsLpage0()
	r.pageSelect.set(h, 0)
	DoI2cRpc()
	rb := getBlocks()
	rb.lpage0b.get(h, 32)
	DoI2cRpc()

	portLpage0[port].id = s[2].D[1]
	portLpage0[port].status = uint16(s[2].D[2]) + uint16(s[2].D[3])<<8
	copy(portLpage0[port].channelStatusInterrupt[:], s[2].D[4:7])
	copy(portLpage0[port].freeMonitorInterruptFlags[:], s[2].D[7:10])
	copy(portLpage0[port].channelMonitorInterruptFlags[:], s[2].D[10:16])
	portLpage0[port].freeMonTemp = uint16(s[2].D[24]) + uint16(s[2].D[23])<<8
	portLpage0[port].freeMonVoltage = uint16(s[2].D[28]) + uint16(s[2].D[27])<<8

	rb.lpage1b.get(h, 32)
	DoI2cRpc()

	copy(portLpage0[port].rxPower[:], s[2].D[3:11])
	copy(portLpage0[port].txBias[:], s[2].D[11:19])
	copy(portLpage0[port].txPower[:], s[2].D[19:27])
}

func (h *I2cDev) StaticBlocks(port int) {
	if !portIsCopper[port] {
		r := getRegsLpage0()
		r.pageSelect.set(h, 3)
		DoI2cRpc()

		rb := getBlocks()
		rb.upage0b.get(h, 32)
		DoI2cRpc()
		portUpage3[port].tempHighAlarm = (uint16(s[2].D[1]) << 8) + uint16(s[2].D[2])
		portUpage3[port].tempLowAlarm = (uint16(s[2].D[3]) << 8) + uint16(s[2].D[4])
		portUpage3[port].tempHighWarning = (uint16(s[2].D[5]) << 8) + uint16(s[2].D[6])
		portUpage3[port].tempLowWarning = (uint16(s[2].D[7]) << 8) + uint16(s[2].D[8])
		portUpage3[port].vccHighAlarm = (uint16(s[2].D[17]) << 8) + uint16(s[2].D[18])
		portUpage3[port].vccLowAlarm = (uint16(s[2].D[19]) << 8) + uint16(s[2].D[20])
		portUpage3[port].vccHighWarning = (uint16(s[2].D[21]) << 8) + uint16(s[2].D[22])
		portUpage3[port].vccLowWarning = (uint16(s[2].D[23]) << 8) + uint16(s[2].D[24])

		rb.upage1b.get(h, 32)
		DoI2cRpc()
		portUpage3[port].rxPowerHighAlarm = (uint16(s[2].D[17]) << 8) + uint16(s[2].D[18])
		portUpage3[port].rxPowerLowAlarm = (uint16(s[2].D[19]) << 8) + uint16(s[2].D[20])
		portUpage3[port].rxPowerHighWarning = (uint16(s[2].D[21]) << 8) + uint16(s[2].D[22])
		portUpage3[port].rxPowerLowWarning = (uint16(s[2].D[23]) << 8) + uint16(s[2].D[24])
		portUpage3[port].txBiasHighAlarm = (uint16(s[2].D[25]) << 8) + uint16(s[2].D[26])
		portUpage3[port].txBiasLowAlarm = (uint16(s[2].D[27]) << 8) + uint16(s[2].D[28])
		portUpage3[port].txBiasHighWarning = (uint16(s[2].D[29]) << 8) + uint16(s[2].D[30])
		portUpage3[port].txBiasLowWarning = (uint16(s[2].D[31]) << 8) + uint16(s[2].D[32])

		rb.upage2b.get(h, 32)
		DoI2cRpc()
		portUpage3[port].txPowerHighAlarm = (uint16(s[2].D[1]) << 8) + uint16(s[2].D[2])
		portUpage3[port].txPowerLowAlarm = (uint16(s[2].D[3]) << 8) + uint16(s[2].D[4])
		portUpage3[port].txPowerHighWarning = (uint16(s[2].D[5]) << 8) + uint16(s[2].D[6])
		portUpage3[port].txPowerLowWarning = (uint16(s[2].D[7]) << 8) + uint16(s[2].D[8])

		r = getRegsLpage0()
		r.pageSelect.set(h, 0)
		DoI2cRpc()
	}
	return
}
