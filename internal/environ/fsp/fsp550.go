// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package fsp provides access to the power supply unit

package fsp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "fsp"

type I2cDev struct {
	Slot       int
	Installed  int
	Id         string
	Bus        int
	Addr       int
	MuxBus     int
	MuxAddr    int
	MuxValue   int
	GpioPwrok  string
	GpioPrsntL string
	GpioPwronL string
	GpioIntL   string
}

var Vdev [2]I2cDev

var VpageByKey map[string]uint8

type cmd struct {
	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]float64)
	cmd.lasts = make(map[string]string)
	cmd.lastu = make(map[string]uint16)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	//if err = cmd.update(); err != nil {
	//	close(cmd.stop)
	//	return err
	//}
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

	for k, i := range VpageByKey {

		pin, found := gpio.Pins[Vdev[i].GpioPrsntL]
		t, err := pin.Value()
		if !found || err != nil || t {
			//not present
			if strings.Contains(k, "status") {
				v := Vdev[i].PsuStatus()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
				}
			}
			if strings.Contains(k, "admin.state") {
				v := Vdev[i].GetAdminState()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
				}
			}
			k := "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_id"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_model"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_in"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_out"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temperature1"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temperature2"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_out"
			cmd.pub.Print("delete: ", k)
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_in"
			cmd.pub.Print("delete: ", k)

			if err != nil {
				log.Print("fsp550 gpio error: ", err)
			}
		} else {
			//present
			if strings.Contains(k, "status") {
				v := Vdev[i].PsuStatus()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
				}
			}
			if strings.Contains(k, "admin.state") {
				v := Vdev[i].GetAdminState()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
				}
			}
			if strings.Contains(k, "page") {
				v := Vdev[i].Page()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "status_word") {
				v := Vdev[i].StatusWord()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "status_vout") {
				v := Vdev[i].StatusVout()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "status_iout") {
				v := Vdev[i].StatusIout()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "status_input") {
				v := Vdev[i].StatusInput()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "v_in") {
				v := Vdev[i].Vin()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "i_in") {
				v := Vdev[i].Iin()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "v_out") {
				v := Vdev[i].Vout()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "i_out") {
				v := Vdev[i].Iout()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "status_temp") {
				v := Vdev[i].StatusTemp()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "p_out") {
				v := Vdev[i].Pout()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "p_in") {
				v := Vdev[i].Pin()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "p_out_raw") {
				v := Vdev[i].PoutRaw()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "p_in_raw") {
				v := Vdev[i].PinRaw()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "p_mode_raw") {
				v := Vdev[i].ModeRaw()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "pmbus_rev") {
				v := Vdev[i].PMBusRev()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "status_fans") {
				v := Vdev[i].StatusFans()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "temperature1") {
				v := Vdev[i].Temp1()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "temperature2") {
				v := Vdev[i].Temp2()
				if v != cmd.last[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.last[k] = v
				}
			}
			if strings.Contains(k, "fan_speed") {
				v := Vdev[i].FanSpeed()
				if v != cmd.lastu[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lastu[k] = v
				}
			}
			if strings.Contains(k, "mfg_id") {
				v := Vdev[i].MfgIdent()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
				}
			}
			if strings.Contains(k, "mfg_model") {
				v := Vdev[i].MfgModel()
				if v != cmd.lasts[k] {
					cmd.pub.Print(k, ": ", v)
					cmd.lasts[k] = v
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

func (h *I2cDev) convert(v uint16) float64 {
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
		return vv
	} else if strings.Contains(h.Id, "FSP") {
		r := getRegs()
		var nn float64
		r.VoutMode.get(h)
		closeMux(h)
		DoI2cRpc()
		n := (uint16(s[1].D[0])) & 0x1f
		if n > 0xf {
			n = ((n ^ 0x1f) + 1) & 0x1f
			nn = float64(n) * (-1)
		} else {
			nn = float64(n)
		}
		vv := (float64(v) * (math.Exp2(nn)))
		vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
		return vv
	} else {
		return 0
	}
}

func (h *I2cDev) Page() uint16 {
	r := getRegs()
	r.Page.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) PageWr(i uint16) {
	r := getRegs()
	r.Page.set(h, uint8(i))
	closeMux(h)
	DoI2cRpc()
	return
}

func (h *I2cDev) StatusWord() uint16 {
	r := getRegs()
	r.StatusWord.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return uint16(t)
}

func (h *I2cDev) StatusVout() uint16 {
	r := getRegs()
	r.StatusVout.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) StatusIout() uint16 {
	r := getRegs()
	r.StatusIout.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) StatusInput() uint16 {
	r := getRegs()
	r.StatusInput.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) StatusTemp() uint16 {
	r := getRegs()
	r.StatusTemp.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) StatusFans() uint16 {
	r := getRegs()
	r.StatusFans.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) Vin() float64 {
	r := getRegs()
	r.Vin.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v := h.convert(t)
	return v
}

func (h *I2cDev) Iin() float64 {
	r := getRegs()
	r.Iin.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v := h.convert(t)
	return v
}

func (h *I2cDev) Vout() float64 {
	r := getRegs()
	r.Vout.get(h)
	//var nn float64
	r.VoutMode.get(h)
	closeMux(h)
	DoI2cRpc()

	vout := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	voutMode := uint8(s[3].D[0])
	v := h.convertVoutMode(voutMode, vout)

	return v
}

func (h *I2cDev) Iout() float64 {
	r := getRegs()
	r.Iout.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v := h.convert(t)
	return v
}

func (h *I2cDev) Temp1() float64 {
	r := getRegs()
	r.Temp1.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v = h.convert(t)
	} else if strings.Contains(h.Id, "FSP") {
		v = float64(t)
	}
	return v
}

func (h *I2cDev) Temp2() float64 {
	r := getRegs()
	r.Temp2.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	var v float64
	if strings.Contains(h.Id, "Great Wall") {
		v = h.convert(t)
	} else if strings.Contains(h.Id, "FSP") {
		v = float64(t)
	}
	return v
}

func (h *I2cDev) FanSpeed() uint16 {
	r := getRegs()
	r.FanSpeed.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[1]) + (uint16(s[1].D[0]) << 8)
	return t
}

func (h *I2cDev) Pout() float64 {
	r := getRegs()
	r.Pout.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v := h.convert(t)
	return v
}

func (h *I2cDev) Pin() float64 {
	r := getRegs()
	r.Pin.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	v := h.convert(t)
	return v
}

func (h *I2cDev) PoutRaw() uint16 {
	r := getRegs()
	r.Pout.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return t
}

func (h *I2cDev) PinRaw() uint16 {
	r := getRegs()
	r.Pin.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
	return t
}

func (h *I2cDev) ModeRaw() uint16 {
	if h.Id == "Great Wall" {
		r := getRegs()
		r.Pin.get(h)
		closeMux(h)
		DoI2cRpc()
		t := uint16(s[1].D[0]) + (uint16(s[1].D[1]) << 8)
		return t
	} else {
		return 0
	}
}

func (h *I2cDev) PMBusRev() uint16 {
	r := getRegs()
	r.PMBusRev.get(h)
	closeMux(h)
	DoI2cRpc()
	t := uint16(s[1].D[0])
	return uint16(t)
}

func (h *I2cDev) MfgIdent() string {
	var l byte = 15
	r := getRegs()
	r.MfgId.get(h, l)
	closeMux(h)
	DoI2cRpc()
	n := s[1].D[1] + 2
	t := string(s[1].D[2:n])
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	h.Id = t
	return t
}

func (h *I2cDev) MfgModel() string {
	var l byte = 15
	r := getRegs()
	r.MfgMod.get(h, l)
	closeMux(h)
	DoI2cRpc()
	n := s[1].D[1] + 2
	t := string(s[1].D[2:n])
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	return t
}

func (h *I2cDev) PsuStatus() string {
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
			h.Installed = 0
			return "not_installed"
		}
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
