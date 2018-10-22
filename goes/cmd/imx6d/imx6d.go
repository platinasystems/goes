// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package imx6d

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
)

const ERRORMAX = 5

var readError int

type Command struct {
	VpageByKey map[string]uint8

	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]float64
}

func (*Command) String() string { return "imx6d" }

func (*Command) Usage() string { return "imx6d" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "ARM CPU temperature daemon, publishes to redis",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var si syscall.Sysinfo_t

	err := redis.IsReady()
	if err != nil {
		return err
	}

	c.stop = make(chan struct{})
	c.last = make(map[string]float64)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err = c.update(); err != nil {
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
	for k, _ := range c.VpageByKey {
		v, err := ReadTemp()
		if err != nil {
			if readError < ERRORMAX {
				readError++
				return nil
			}
			return err
		}
		readError = 0
		if v != c.last[k] {
			c.pub.Print(k, ": ", v)
			c.last[k] = v
		}
	}
	return nil
}

func ReadTemp() (float64, error) {
	tmp, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}
	tmp2 := fmt.Sprintf("%.4s", string(tmp[:]))
	tmp3, _ := strconv.Atoi(tmp2)
	tmp4 := float64(tmp3)
	return float64(tmp4 / 100.0), nil
}
