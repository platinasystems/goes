// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qsfpio

import (
	//	"strconv"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	//	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "qsfpio"

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

type qsfpI2cGpio struct {
	present [2]uint16
}

var Vdev [2]I2cDev
var qsfpIG qsfpI2cGpio

var VpageByKey map[string]uint8

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]string
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error

	qsfpIG.present[0] = 0xffff
	qsfpIG.present[1] = 0xffff

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]string)

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

	port := uint8(0)
	for k, i := range VpageByKey {
		for j := 1; j < 33; j++ {
			if strings.Contains(k, "port."+strconv.Itoa(int(j))+".") {
				port = uint8(j) - 1
				//log.Print("qsfp port: ", port)
				break
			}
		}
		v := Vdev[i].QsfpStatus(port)
		if v != cmd.last[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.last[k] = v
		}
	}
	return nil
}

func (h *I2cDev) QsfpStatus(port uint8) string {
	r := getRegs()
	var present uint16

	if port == 0 || port == 16 {

		r.Input[0].get(h)
		DoI2cRpc()
		p := uint16(s[1].D[0])

		r.Input[1].get(h)
		DoI2cRpc()
		p += uint16(s[1].D[0]) << 8
		if port == 0 {
			qsfpIG.present[0] = p
		} else {
			qsfpIG.present[1] = p
		}
	}

	if port < 16 {
		present = qsfpIG.present[0]
	} else {
		present = qsfpIG.present[1]
	}

	pmask := uint16(1) << (port % 16)
	log.Print("p: ", present, " port: ", port, " pmask: ", pmask)
	if (present&pmask)>>(port%16) == 1 {
		return "not_installed"
	}
	return "installed"
}
