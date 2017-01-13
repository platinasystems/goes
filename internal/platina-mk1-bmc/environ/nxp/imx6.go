// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package nxp provides access to the NXP iMX6 ARM CPU

package imx6

import (
	"fmt"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "imx6"
const everything = true
const onlyChanges = false

type cmd chan struct{}

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
	//update(everything)
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
			update(onlyChanges)
		}
	}
	return nil
}

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

func update(everything bool) error {
	var si syscall.Sysinfo_t
	if err := syscall.Sysinfo(&si); err != nil {
		return err
	}
	pub, err := redis.Publish(redis.DefaultHash)
	if err != nil {
		return err
	}

	if everything {
		pub <- fmt.Sprint("cpu.status: ", 1)
	} else {
		pub <- fmt.Sprint("cpu.status: ", 1)
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
