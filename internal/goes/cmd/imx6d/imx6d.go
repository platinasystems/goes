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

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const (
	Name    = "imx6d"
	Apropos = "FIXME"
	Usage   = "imx6d"
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

var (
	Init = func() {}
	once sync.Once

	VpageByKey map[string]uint8
)

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]float64
}

func (*cmd) Apropos() lang.Alt { return apropos }
func (*cmd) Kind() goes.Kind   { return goes.Daemon }
func (*cmd) String() string    { return Name }
func (*cmd) Usage() string     { return Name }

func (cmd *cmd) Main(...string) error {
	once.Do(Init)

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
	for k, _ := range VpageByKey {
		v := ReadTemp()
		if v != cmd.last[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.last[k] = v
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

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
