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
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "ucd9090"

var ( // FIXME these are is machine specific
	VpageByKey = map[string]uint8{
		"vmon.5v.sb":    1,
		"vmon.3v8.bmc":  2,
		"vmon.3v3.sys":  3,
		"vmon.3v3.bmc":  4,
		"vmon.3v3.sb":   5,
		"vmon.1v0.thc":  6,
		"vmon.1v8.sys":  7,
		"vmon.1v25.sys": 8,
		"vmon.1v2.ethx": 9,
		"vmon.1v0.tha":  10,
	}
	Vdev = I2cDev{
		Bus:      0,
		Addr:     0x7e,
		MuxBus:   0,
		MuxAddr:  0x76,
		MuxValue: 0x01,
	}
)

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]float64
}

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
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

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	cmd.update()

	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			cmd.update()
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	close(cmd.stop)
	return nil
}

func (cmd *cmd) update() {
	for k, i := range VpageByKey {
		v := Vdev.Vout(i)
		if v != cmd.last[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.last[k] = v
		}
	}
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
