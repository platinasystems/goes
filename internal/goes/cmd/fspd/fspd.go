// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package fspd provides access to the power supply unit

package fspd

import (
	"fmt"
	"math"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/gpio"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const Name = "fspd"

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
}

var (
	Init = func() {}
	once sync.Once

	Vdev [2]I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)
)

type cmd struct {
	Info
}

type Info struct {
	mutex sync.Mutex
	rpc   *sockfile.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
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
	cmd.lastu = make(map[string]uint16)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if cmd.rpc, err = sockfile.NewRpcServer(Name); err != nil {
		return err
	}

	rpc.Register(&cmd.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", Name, "Info")
		if err != nil {
			return err
		}
	}

	holdOff := 3
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if holdOff > 0 {
				holdOff--
			}
			if holdOff == 0 {
				if err = cmd.update(); err != nil {
					close(cmd.stop)
					return err
				}
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
			k := "psu" + strconv.Itoa(Vdev[i].Slot) + ".eeprom"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".fan_speed.units.rpm"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".i_out.units.A"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_id"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".mfg_model"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_in.units.W"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".p_out.units.W"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temp1.units.C"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".temp2.units.C"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_out.units.V"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""
			k = "psu" + strconv.Itoa(Vdev[i].Slot) + ".v_in.units.V"
			cmd.pub.Print("delete: ", k)
			cmd.lasts[k] = ""

			if err != nil {
				log.Print("fspd gpio error: ", err)
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
			if Vdev[i].Update[0] {
				if strings.Contains(k, "mfg_id") {
					v, err := Vdev[i].MfgIdent()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
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
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
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
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
					Vdev[i].Update[2] = false
				}
			}

			if Vdev[i].Id != "" {
				if strings.Contains(k, "page") {
					v, err := Vdev[i].Page()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_word") {
					v, err := Vdev[i].StatusWord()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_vout") {
					v, err := Vdev[i].StatusVout()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_iout") {
					v, err := Vdev[i].StatusIout()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_input") {
					v, err := Vdev[i].StatusInput()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "v_in") {
					v, err := Vdev[i].Vin()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "i_in") {
					v, err := Vdev[i].Iin()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "v_out") {
					v, err := Vdev[i].Vout()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "i_out") {
					v, err := Vdev[i].Iout()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "status_temp") {
					v, err := Vdev[i].StatusTemp()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_out") {
					v, err := Vdev[i].Pout()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "p_in") {
					v, err := Vdev[i].Pin()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "p_out_raw") {
					v, err := Vdev[i].PoutRaw()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_in_raw") {
					v, err := Vdev[i].PinRaw()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "p_mode_raw") {
					v, err := Vdev[i].ModeRaw()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "pmbus_rev") {
					v, err := Vdev[i].PMBusRev()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "status_fans") {
					v, err := Vdev[i].StatusFans()
					if err != nil {
						return err
					}
					if v != cmd.lastu[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lastu[k] = v
					}
				}
				if strings.Contains(k, "temp1") {
					v, err := Vdev[i].Temp1()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "temp2") {
					v, err := Vdev[i].Temp2()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "fan_speed.units.rpm") {
					v, err := Vdev[i].FanSpeed()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				/*if strings.Contains(k, "mfg_id") {
					v, err := Vdev[i].MfgIdent()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}
				if strings.Contains(k, "mfg_model") {
					v, err := Vdev[i].MfgModel()
					if err != nil {
						return err
					}
					if v != cmd.lasts[k] {
						cmd.pub.Print(k, ": ", v)
						cmd.lasts[k] = v
					}
				}*/
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
	v, errs := h.convert(t)
	if errs != nil {
		log.Print("convert err")
		return "", errs
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
	v, errs := h.convert(t)
	if errs != nil {
		return "", errs
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
	var data = [34]byte{0, 0, 0, 0}
	data[0] = byte(h.MuxValue)

	for n := 0; n < 16; n++ {
		j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 5}
		x++
		for i := n * 16; i < n*16+16; i++ {
			j[x] = I{true, i2c.Read, uint8(i), i2c.ByteData, data, h.Bus, h.AddrProm, 0}
			x++
		}
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return "", err
		}

		for k := 1; k < 17; k++ {
			v += fmt.Sprintf("%02x ", s[k].D[0])
		}
	}
	return v, nil
}

func (h *I2cDev) PsuStatus() string {
	if len(gpio.Pins) == 0 {
		gpio.Init()
	}
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
	time.Sleep(100 * time.Millisecond)

	startI2c()
	time.Sleep(1 * time.Second)

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
	return fmt.Errorf("Cannot hset.  Valid values are: %s", WrRegRng[args.Field])
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}
