// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package nxp provides access to the NXP iMX6 ARM CPU

package imx6

import (
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "imx6"
const everything = true
const onlyChanges = false

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error

	cmd.stop = make(chan struct{})

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if err = cmd.update(everything); err != nil {
		return err
	}

	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if err = cmd.update(onlyChanges); err != nil {
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

func (cmd *cmd) update(everything bool) error {
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err != nil {
		return err
	}

	if everything {
		cmd.pub.Print("cpu.status: ", 1)
	} else {
		cmd.pub.Print("cpu.status: ", 1)
	}
	return nil
}

/*
package imx6

import (
	"fmt"
	"io/ioutil"
	"strconv"
)

type Cpu struct {
}

func (h *Cpu) ReadTemp() float64 {
	tmp, _ := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	tmp2 := fmt.Sprintf("%.4s", string(tmp[:]))
	tmp3, _ := strconv.Atoi(tmp2)
	tmp4 := float64(tmp3)
	return float64(tmp4 / 100.0)
}
*/
