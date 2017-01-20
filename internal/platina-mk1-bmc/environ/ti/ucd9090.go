// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ucd9090

import (
	"fmt"
	"math"
	"strconv"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name        = "ucd9090"
	everything  = true
	onlyChanges = false
)
const (
	ucd9090Bus    = 0
	ucd9090Adr    = 0x7e
	ucd9090MuxBus = 0
	ucd9090MuxAdr = 0x76
	ucd9090MuxVal = 0x01
)

type cmd chan struct{}

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var dev = I2cDev{ucd9090Bus, ucd9090Adr, ucd9090MuxBus, ucd9090MuxAdr, ucd9090MuxVal}

func New() cmd { return cmd(make(chan struct{})) }

func (cmd) Kind() goes.Kind { return goes.Daemon }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return Name }

func (cmd cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	err := syscall.Sysinfo(&si)
	if err != nil {
		return err
	}
	//update()
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
			update()
		}
	}
	return nil
}

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

func update() error {
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err != nil {
		return err
	}
	pub, err := redis.Publish(redis.DefaultHash)
	if err != nil {
		return err
	}
	pub <- fmt.Sprint("vmon.5v.sb: ", dev.Vout(1))
	pub <- fmt.Sprint("vmon.3v8.bmc: ", dev.Vout(2))
	pub <- fmt.Sprint("vmon.3v3.sys: ", dev.Vout(3))
	pub <- fmt.Sprint("vmon.3v3.bmc: ", dev.Vout(4))
	pub <- fmt.Sprint("vmon.3v3.sb: ", dev.Vout(5))
	pub <- fmt.Sprint("vmon.1v0.thc: ", dev.Vout(6))
	pub <- fmt.Sprint("vmon.1v8.sys: ", dev.Vout(7))
	pub <- fmt.Sprint("vmon.1v25.sys: ", dev.Vout(8))
	pub <- fmt.Sprint("vmon.1v2.ethx: ", dev.Vout(9))
	pub <- fmt.Sprint("vmon.1v0.tha: ", dev.Vout(10))
	pub <- fmt.Sprint("vmon.5v.sb: ", dev.Vout(1))
	pub <- fmt.Sprint("vmon.3v8.bmc: ", dev.Vout(2))
	pub <- fmt.Sprint("vmon.3v3.sys: ", dev.Vout(3))
	pub <- fmt.Sprint("vmon.3v3.bmc: ", dev.Vout(4))
	pub <- fmt.Sprint("vmon.3v3.sb: ", dev.Vout(5))
	pub <- fmt.Sprint("vmon.1v0.thc: ", dev.Vout(6))
	pub <- fmt.Sprint("vmon.1v8.sys: ", dev.Vout(7))
	pub <- fmt.Sprint("vmon.1v25.sys: ", dev.Vout(8))
	pub <- fmt.Sprint("vmon.1v2.ethx: ", dev.Vout(9))
	pub <- fmt.Sprint("vmon.1v0.tha: ", dev.Vout(10))
	return nil
}

func (h *I2cDev) Vout(i uint8) float64 {
	if i > 10 {
		panic("Voltage rail subscript out of range\n")
	}
	i--

	clearJS()
	r := getPwmRegs()
	r.Page.set(h, i)
	r.VoutMode.get(h)
	r.ReadVout.get(h)
	DoI2cRpc()
	//log.Print("return vals: ", s[0].D[0], s[1].D[0], s[2].D[0], s[3].D[0], s[4].D[0], s[5].D[0], s[5].D[1])
	n := s[3].D[0] & 0xf
	n--
	n = (n ^ 0xf) & 0xf
	v := uint16(s[5].D[1])<<8 | uint16(s[5].D[0])

	nn := float64(n) * (-1)
	vv := float64(v) * (math.Exp2(nn))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return float64(vv)
}
