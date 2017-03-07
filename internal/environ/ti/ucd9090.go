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

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var Vdev I2cDev

var VpageByKey map[string]uint8

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]float64
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

	//if err = cmd.update(); err != nil {
	//	close(cmd.stop)
	//	return err
	//}
	t := time.NewTicker(10 * time.Second)
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
		v, err := Vdev.Vout(i)
		if err != nil {
			return err
		}
		if v != cmd.last[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.last[k] = v
		}
	}
	return nil
}

func (h *I2cDev) Vout(i uint8) (float64, error) {
	if i > 10 {
		panic("Voltage rail subscript out of range\n")
	}
	i--

	r := getRegs()
	r.Page.set(h, i)
	r.VoutMode.get(h)
	r.ReadVout.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	n := s[3].D[0] & 0xf
	n--
	n = (n ^ 0xf) & 0xf
	v := uint16(s[5].D[1])<<8 | uint16(s[5].D[0])

	nn := float64(n) * (-1)
	vv := float64(v) * (math.Exp2(nn))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return float64(vv), nil
}
