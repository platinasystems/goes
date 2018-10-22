// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package fspd provides access to the power supply unit

package fspd

import (
	"encoding/hex"
	"fmt"
	"math"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/atsock"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
)

var (
	Vdev [2]I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)

	command *Command
)

const (
	nFanTrays = 4
	nPSUs     = 2
)

type Command struct {
	Info
	Init func()
	init sync.Once
	Gpio func()
	gpio sync.Once
}

type Info struct {
	mutex sync.Mutex
	rpc   *atsock.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

type I2cDev struct {
	Slot       int
	Installed  int
	Id         string
	Model      string
	Bus        int
	Addr       int
	AddrProm   int
	MuxBus     int
	MuxAddr    int
	MuxValue   int
	GpioPwrok  string
	GpioPrsntL string
	GpioPwronL string
	GpioIntL   string
	Update     [3]bool
	Delete     bool
}

func (*Command) String() string { return "fspd" }

func (*Command) Usage() string { return "fspd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "fsp power supply daemon, publishes to redis",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var si syscall.Sysinfo_t

	command = c
	c.init.Do(c.Init)

	err := redis.IsReady()
	if err != nil {
		return err
	}

	c.stop = make(chan struct{})
	c.last = make(map[string]float64)
	c.lasts = make(map[string]string)
	c.lastu = make(map[string]uint16)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if c.rpc, err = atsock.NewRpcServer("fspd"); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", "fspd", "Info")
		if err != nil {
			return err
		}
	}

	holdOff := 3
	t := time.NewTicker(1 * time.Second)
	tm := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if holdOff == 0 {
				if err = c.update(); err != nil {
					holdOff = 5
				}
			}
		case <-tm.C:
			if holdOff > 0 {
				holdOff--
			}
			if holdOff == 0 {
				if err = c.updateMon(); err != nil {
					holdOff = 5
				}
			}
		}
	}
	return nil
}

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (c *Command) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}
	if err := writeRegs(); err != nil {
		return err
	}

	for k, i := range VpageByKey {

		pin, found := gpio.Pins[Vdev[i].GpioPrsntL]
		t, err := pin.Value()
		if !found || err != nil || t {
			//not present
			if strings.Contains(k, "status") {
				v := Vdev[i].PsuStatus()
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
			}
			if strings.Contains(k, "admin.state") {
				v := Vdev[i].GetAdminState()
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
			}
			if Vdev[i].Delete {
				k := "psu" + strconv.Itoa(Vdev[i].Slot) + ".eeprom"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".sn"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".fan_speed.units.rpm"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".i_out.units.A"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_id"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_model"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_in.units.W"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_out.units.W"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temp1.units.C"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temp2.units.C"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".fan_direction"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_out.units.V"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_in.units.V"
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
				Vdev[i].Delete = false
			}

			if err != nil {
				log.Print("fspd gpio error: ", err)
			}
		} else {
			//present
			if strings.Contains(k, "status") {
				v := Vdev[i].PsuStatus()
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
			}
			if strings.Contains(k, "admin.state") {
				v := Vdev[i].GetAdminState()
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
			}
			if Vdev[i].Update[0] {
				if strings.Contains(k, "mfg_id") {
					v, err := Vdev[i].MfgIdent()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
					Vdev[i].Update[0] = false
				}
			}
			if Vdev[i].Update[1] {
				if strings.Contains(k, "mfg_model") {
					v, err := Vdev[i].MfgModel()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
					Vdev[i].Update[1] = false
				}
			}
			if Vdev[i].Update[2] {
				if strings.Contains(k, "eeprom") {
					v, err := Vdev[i].Eeprom()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
					Vdev[i].Update[2] = false
					var d string
					if v[0x38:0x3a] == "52" {
						d = "back->front"
					} else {
						d = "front->back"
					}
					k = strings.Replace(k, "eeprom", "fan_direction", 1)
					if d != c.lasts[k] {
						c.pub.Print(k, ": ", d)
						c.lasts[k] = d
					}
					sb, err := hex.DecodeString(v[0x5a:0x76])
					sn := string(sb)
					if err == nil {
						k = strings.Replace(k, "fan_direction", "sn", 1)
						if sn != c.lasts[k] {
							c.pub.Print(k, ": ", sn)
							c.lasts[k] = sn
						}
					}
				}
			}
		}
	}
	return nil
}

func (c *Command) updateMon() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}
	if err := writeRegs(); err != nil {
		return err
	}

	for k, i := range VpageByKey {

		pin, found := gpio.Pins[Vdev[i].GpioPrsntL]
		t, err := pin.Value()
		if !found || err != nil || t {
			// PSU not present
			return nil
		} else {
			// PSU present
			if Vdev[i].Id != "" {
				if strings.Contains(k, "page") {
					v, err := Vdev[i].Page()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_word") {
					v, err := Vdev[i].StatusWord()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_vout") {
					v, err := Vdev[i].StatusVout()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_iout") {
					v, err := Vdev[i].StatusIout()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_input") {
					v, err := Vdev[i].StatusInput()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "v_in") {
					v, err := Vdev[i].Vin()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "i_in") {
					v, err := Vdev[i].Iin()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "v_out") {
					v, err := Vdev[i].Vout()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "i_out") {
					v, err := Vdev[i].Iout()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "status_temp") {
					v, err := Vdev[i].StatusTemp()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_out") {
					v, err := Vdev[i].Pout()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "p_in") {
					v, err := Vdev[i].Pin()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "p_out_raw") {
					v, err := Vdev[i].PoutRaw()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_in_raw") {
					v, err := Vdev[i].PinRaw()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_mode_raw") {
					v, err := Vdev[i].ModeRaw()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "pmbus_rev") {
					v, err := Vdev[i].PMBusRev()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_fans") {
					v, err := Vdev[i].StatusFans()
					if err != nil {
						return err
					}
					if v != c.lastu[k] {
						c.pub.Print(k, ": ", v)
						c.lastu[k] = v
					}
				}
				if strings.Contains(k, "temp1") {
					v, err := Vdev[i].Temp1()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "temp2") {
					v, err := Vdev[i].Temp2()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
				if strings.Contains(k, "fan_speed.units.rpm") {
					v, err := Vdev[i].FanSpeed()
					if err != nil {
						return err
					}
					if v != c.lasts[k] {
						c.pub.Print(k, ": ", v)
						c.lasts[k] = v
					}
				}
			}

		}
	}
	return nil
}

func (h *I2cDev) convertVoutMode(voutMode uint8, vout uint16) float64 {
	var nn float64
	n := voutMode & 0x1f
	if n > 0xf {
		n = ((n ^ 0x1f) + 1) & 0x1f
		nn = float64(n) * (-1)
	} else {
		nn = float64(n)
	}
	vv := (float64(vout) * (math.Exp2(nn)))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return vv
}

func (h *I2cDev) convertLinear(v uint16) (float64, error) {
	var nn int
	var y int
	if (v >> 11) > 0xf {
		nn = int(((v>>11)^0x1f)+1) * (-1)
	} else {
		nn = int(v >> 11)
	}
	v = v & 0x7ff
	if v > 0x3ff {
		y = int(v^0x7ff+1) * (-1)
	} else {
		y = int(v)
	}
	vv := float64(y) * (math.Exp2(float64(nn)))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return vv, nil
}

func (h *I2cDev) convert(v uint16) (float64, error) {
	if strings.Contains(h.Id, "Great Wall") {
		var nn int
		var y int
		if (v >> 11) > 0xf {
			nn = int(((v>>11)^0x1f)+1) * (-1)
		} else {
			nn = int(v >> 11)
		}
		v = v & 0x7ff
		if v > 0x3ff {
			y = int(v^0x7ff+1) * (-1)
		} else {
			y = int(v)
		}
		vv := float64(y) * (math.Exp2(float64(nn)))
		vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
		return vv, nil
	} else if strings.Contains(h.Id, "FSP") {
		r := getRegs()
		var nn float64
		r.VoutMode.get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return 0, err
		}
		n := (uint16(s[1].D[0])) & 0x1f
		if n > 0xf {
			n = ((n ^ 0x1f) + 1) & 0x1f
			nn = float64(n) * (-1)
		} else {
			nn = float64(n)
		}
		vv := (float64(v) * (math.Exp2(nn)))
		vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
		return vv, nil
	} else {
		return 0, nil
	}
}

func (h *I2cDev) Page() (uint16, error) {
	r := getRegs()
	r.Page.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) PageWr(i uint16) error {
	r := getRegs()
	r.Page.set(h, uint8(i))
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) StatusWord() (uint16, error) {
	r := getRegs()
	r.StatusWord.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return uint16(t), nil
}

func (h *I2cDev) StatusVout() (uint16, error) {
	r := getRegs()
	r.StatusVout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) StatusIout() (uint16, error) {
	r := getRegs()
	r.StatusIout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) StatusInput() (uint16, error) {
	r := getRegs()
	r.StatusInput.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) StatusTemp() (uint16, error) {
	r := getRegs()
	r.StatusTemp.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) StatusFans() (uint16, error) {
	r := getRegs()
	r.StatusFans.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) Vin() (string, error) {
	r := getRegs()
	r.Vin.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v, errs := h.convert(t)
	if errs != nil {
		return "", errs
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Iin() (string, error) {
	r := getRegs()
	r.Iin.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v, errs := h.convert(t)
	if errs != nil {
		return "", errs
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Vout() (string, error) {
	r := getRegs()
	r.Vout.get(h)
	r.VoutMode.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	vout := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	voutMode := uint8(s[3].D[0])
	var v float64
	var errs error
	if !strings.Contains(h.Model, "CRPS800") {
		v = h.convertVoutMode(voutMode, vout)
	} else {
		v, errs = h.convert(vout)
		if errs != nil {
			return "", errs
		}
	}

	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Iout() (string, error) {
	r := getRegs()
	r.Iout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v, err = h.convert(t)
		if err != nil {
			return "", err
		}
	} else if strings.Contains(h.Id, "FSP") {
		v, err = h.convertLinear(t)
		if err != nil {
			return "", err
		}
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Temp1() (string, error) {
	r := getRegs()
	r.Temp1.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v, err = h.convert(t)
		if err != nil {
			return "", err
		}
	} else if strings.Contains(h.Id, "FSP") {
		v = float64(t)
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Temp2() (string, error) {
	r := getRegs()
	r.Temp2.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v, err = h.convert(t)
		if err != nil {
			return "", err
		}
	} else if strings.Contains(h.Id, "FSP") {
		v = float64(t)
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) FanSpeed() (string, error) {
	r := getRegs()
	r.FanSpeed.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v, err = h.convert(t)
		if err != nil {
			return "", err
		}
	} else if strings.Contains(h.Id, "FSP") {
		v = float64(t)
	}
	return strconv.FormatFloat(v, 'f', 0, 64), nil
}

func (h *I2cDev) Pout() (string, error) {
	r := getRegs()
	r.Pout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v, errs := h.convert(t)
	if errs != nil {
		return "", errs
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) Pin() (string, error) {
	r := getRegs()
	r.Pin.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v, errs := h.convert(t)
	if errs != nil {
		return "", errs
	}
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) PoutRaw() (uint16, error) {
	r := getRegs()
	r.Pout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return t, nil
}

func (h *I2cDev) PinRaw() (uint16, error) {
	r := getRegs()
	r.Pin.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return t, nil
}

func (h *I2cDev) ModeRaw() (uint16, error) {
	if h.Id == "Great Wall" {
		r := getRegs()
		r.Pin.get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return 0, err
		}
		t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
		return t, nil
	} else {
		return 0, nil
	}
}

func (h *I2cDev) PMBusRev() (uint16, error) {
	r := getRegs()
	r.PMBusRev.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	t := uint16(s[1].D[0])
	return uint16(t), nil
}

func (h *I2cDev) MfgIdent() (string, error) {
	var l byte = 15
	r := getRegs()
	r.MfgId.get(h, l)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "error", err
	}
	if s[1].D[1] == 0xff {
		h.Id = "FSP"
		return "FSP", nil
	}
	n := s[1].D[1] + 2
	t := string(s[1].D[2:n])
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	h.Id = t
	return t, nil
}

func (h *I2cDev) MfgModel() (string, error) {
	var l byte = 15
	r := getRegs()
	r.MfgMod.get(h, l)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "error", err
	}
	if s[1].D[1] == 0xff {
		return "FSP", nil
	}
	n := s[1].D[1] + 2
	t := string(s[1].D[2:n])
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	h.Model = t
	return t, nil
}

func (h *I2cDev) Eeprom() (string, error) {
	var v string
	r := getRegsE()
	for n := 0; n < 8; n++ {
		r.block[n].get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return "", err
		}
		for k := 1; k < 33; k++ {
			v += fmt.Sprintf("%02x", s[1].D[k])
		}
	}
	return v, nil
}

func (h *I2cDev) PsuStatus() string {
	command.gpio.Do(command.Gpio)
	pin, found := gpio.Pins[h.GpioPrsntL]
	if !found {
		h.Installed = 0
		return "not_found"
	} else {
		t, err := pin.Value()
		if err != nil {
			h.Installed = 0
			return err.Error()
		} else if t {
			if h.Installed == 1 {
				h.Delete = true
			}
			h.Installed = 0
			return "not_installed"
		}
	}
	if h.Installed == 0 {
		h.Update = [3]bool{true, true, true}
	}
	h.Installed = 1
	pin, found = gpio.Pins[h.GpioPwrok]
	if !found {
		return "undetermined"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if !t {
		return "powered_off"
	}
	return "powered_on"
}

func (h *I2cDev) SetAdminState(s string) {
	pin, found := gpio.Pins[h.GpioPwronL]
	if found {
		switch s {
		case "disable":
			pin.SetValue(true)
			log.Print("notice: psu", h.Slot, " ", s)
		case "enable":
			pin.SetValue(false)
			log.Print("notice: psu", h.Slot, " ", s)
		}
	}
}

func (h *I2cDev) GetAdminState() string {
	pin, found := gpio.Pins[h.GpioPwronL]
	if !found {
		return "not found"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if t {
		return "disabled"
	}
	return "enabled"
}

func powerCycle() error {

	stopI2c()
	time.Sleep(500 * time.Millisecond)

	log.Print("initiate manual power cycle")
	pin, found := gpio.Pins["PSU0_PWRON_L"]
	if found {
		pin.SetValue(true)
	}
	pin, found = gpio.Pins["PSU1_PWRON_L"]
	if found {
		pin.SetValue(true)
	}
	time.Sleep(1 * time.Second)
	pin, found = gpio.Pins["PSU0_PWRON_L"]
	if found {
		pin.SetValue(false)
	}
	pin, found = gpio.Pins["PSU1_PWRON_L"]
	if found {
		pin.SetValue(false)
	}
	time.Sleep(1 * time.Second)
	pin, found = gpio.Pins["ETHX_RST_L"]
	if found {
		pin.SetValue(false)
		time.Sleep(50 * time.Millisecond)
		pin.SetValue(true)
	}
	startI2c()
	return nil
}

func writeRegs() error {
	for k, v := range WrRegVal {
		switch WrRegFn[k] {
		case "example":
			//check for psuX when writing
			if false {
				log.Print("test", k, v)
			}
		case "powercycle":
			if v == "true" {
				powerCycle()
			}
		case "admin.state":
			if strings.Contains(k, "psu1") {
				Vdev[1].SetAdminState(v)
			} else if strings.Contains(k, "psu2") {
				Vdev[0].SetAdminState(v)
			}
		}
		delete(WrRegVal, k)
	}
	return nil
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	_, p := WrRegFn[args.Field]
	if !p {
		return fmt.Errorf("cannot hset: %s", args.Field)
	}
	_, q := WrRegRng[args.Field]
	if !q {
		err := i.set(args.Field, string(args.Value), false)
		if err == nil {
			*reply = 1
			WrRegVal[args.Field] = string(args.Value)
		}
		return err
	}
	var a [2]int
	var e [2]error
	if len(WrRegRng[args.Field]) == 2 {
		for i, v := range WrRegRng[args.Field] {
			a[i], e[i] = strconv.Atoi(v)
		}
		if e[0] == nil && e[1] == nil {
			val, err := strconv.Atoi(string(args.Value))
			if err != nil {
				return err
			}
			if val >= a[0] && val <= a[1] {
				err := i.set(args.Field,
					string(args.Value), false)
				if err == nil {
					*reply = 1
					WrRegVal[args.Field] =
						string(args.Value)
				}
				return err
			}
			return fmt.Errorf("Cannot hset.  Valid range is: %s",
				WrRegRng[args.Field])
		}
	}
	for _, v := range WrRegRng[args.Field] {
		if v == string(args.Value) {
			err := i.set(args.Field, string(args.Value), false)
			if err == nil {
				*reply = 1
				WrRegVal[args.Field] = string(args.Value)
			}
			return err
		}
	}
	return fmt.Errorf("Cannot hset.  Valid values are: %s",
		WrRegRng[args.Field])
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}
