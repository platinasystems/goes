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

type PMon struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var pm = PMon{ucd9090Bus, ucd9090Adr, ucd9090MuxBus, ucd9090MuxAdr, ucd9090MuxVal}

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
	pub <- fmt.Sprint("vmon.5v.sb: ", pm.Vout(1))
	pub <- fmt.Sprint("vmon.3v8.bmc: ", pm.Vout(2))
	pub <- fmt.Sprint("vmon.3v3.sys: ", pm.Vout(3))
	pub <- fmt.Sprint("vmon.3v3.bmc: ", pm.Vout(4))
	pub <- fmt.Sprint("vmon.3v3.sb: ", pm.Vout(5))
	pub <- fmt.Sprint("vmon.1v0.thc: ", pm.Vout(6))
	pub <- fmt.Sprint("vmon.1v8.sys: ", pm.Vout(7))
	pub <- fmt.Sprint("vmon.1v25.sys: ", pm.Vout(8))
	pub <- fmt.Sprint("vmon.1v2.ethx: ", pm.Vout(9))
	pub <- fmt.Sprint("vmon.1v0.tha: ", pm.Vout(10))
	pub <- fmt.Sprint("vmon.5v.sb: ", pm.Vout(1))
	pub <- fmt.Sprint("vmon.3v8.bmc: ", pm.Vout(2))
	pub <- fmt.Sprint("vmon.3v3.sys: ", pm.Vout(3))
	pub <- fmt.Sprint("vmon.3v3.bmc: ", pm.Vout(4))
	pub <- fmt.Sprint("vmon.3v3.sb: ", pm.Vout(5))
	pub <- fmt.Sprint("vmon.1v0.thc: ", pm.Vout(6))
	pub <- fmt.Sprint("vmon.1v8.sys: ", pm.Vout(7))
	pub <- fmt.Sprint("vmon.1v25.sys: ", pm.Vout(8))
	pub <- fmt.Sprint("vmon.1v2.ethx: ", pm.Vout(9))
	pub <- fmt.Sprint("vmon.1v0.tha: ", pm.Vout(10))
	return nil
}

func (h *PMon) Vout(i uint8) float64 {
	if i > 10 {
		panic("Voltage rail subscript out of range\n")
	}
	i--

	clearJS()
	r := getPwmRegs()
	r.Page.set(h, i)
	r.ReadVout.get(h)
	DoI2cRpc()
	n := s[0].I & 0xf //Response #0, uint8
	n--
	n = (n ^ 0xf) & 0xf
	v := s[1].J //Response #1, uint16

	nn := float64(n) * (-1)
	vv := float64(v) * (math.Exp2(nn))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return float64(vv)
}
