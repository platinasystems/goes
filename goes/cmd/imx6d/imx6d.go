// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package imx6d

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const (
	Name    = "imx6d"
	Apropos = "ARM CPU temperature daemon, publishes to redis"
	Usage   = "imx6d"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

var (
	Init = func() {}
	once sync.Once

	VpageByKey map[string]uint8
)

type Command struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]float64
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Name }

func (c *Command) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t
	var err error

	c.stop = make(chan struct{})
	c.last = make(map[string]float64)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err = c.update(); err != nil {
				close(c.stop)
				return err
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
	for k, _ := range VpageByKey {
		v := ReadTemp()
		if v != c.last[k] {
			c.pub.Print(k, ": ", v)
			c.last[k] = v
		}
	}
	return nil
}

func ReadTemp() float64 {
	tmp, _ := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	tmp2 := fmt.Sprintf("%.4s", string(tmp[:]))
	tmp3, _ := strconv.Atoi(tmp2)
	tmp4 := float64(tmp3)
	return float64(tmp4 / 100.0)
}
