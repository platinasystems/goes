// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qsfpio

import (
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
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

type QsfpI2cGpio struct {
	init    int
	Present [2]uint16
}

var Vdev [8]I2cDev
var qsfpIG QsfpI2cGpio

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

	qsfpIG.Present[0] = 0xffff
	qsfpIG.Present[1] = 0xffff
	qsfpIG.init = 1

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
	t := time.NewTicker(1 * time.Second)
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
	var Present uint16

	if port == 0 || port == 16 {
		//initialize reset I2C GPIO
		if qsfpIG.init == 1 {
			Vdev[0].QsfpInit(0xff, 0xff, 0xff, 0xff)
			Vdev[1].QsfpInit(0xff, 0xff, 0xff, 0xff)
			Vdev[2].QsfpInit(0xff, 0xff, 0xff, 0xff)
			Vdev[3].QsfpInit(0xff, 0xff, 0xff, 0xff)
			Vdev[4].QsfpInit(0xff, 0xff, 0x00, 0x00)
			Vdev[5].QsfpInit(0xff, 0xff, 0x00, 0x00)
			Vdev[6].QsfpInit(0x0, 0x0, 0x0, 0x0)
			Vdev[7].QsfpInit(0x0, 0x0, 0x0, 0x0)
			qsfpIG.init = 0
		}

		r.Input[0].get(h)
		DoI2cRpc()
		p := uint16(s[1].D[0])

		r.Input[1].get(h)
		DoI2cRpc()
		p += uint16(s[1].D[0]) << 8
		if port == 0 && qsfpIG.Present[0] != p {
			//Take port out of reset and LP mode if qsfp is installed
			Vdev[6].QsfpReset((p ^ qsfpIG.Present[0]), p^0xffff)
			Vdev[4].QsfpLpMode((p ^ qsfpIG.Present[0]), p)
			qsfpIG.Present[0] = p

			//send to qspi.go
			err := SendPresRpc()
			if err != nil {
				log.Print("SendPresRpc error:", err)
			}
		} else if port == 16 && qsfpIG.Present[1] != p {
			//Take port out of reset and LP mode if qsfp is installed
			Vdev[7].QsfpReset((p ^ qsfpIG.Present[1]), p^0xffff)
			Vdev[5].QsfpLpMode((p ^ qsfpIG.Present[1]), p)
			qsfpIG.Present[1] = p

			//send to qspi.go
			err := SendPresRpc()
			if err != nil {
				log.Print("SendPresRpc error:", err)
			}
		}
	}

	if port < 16 {
		Present = qsfpIG.Present[0]
	} else {
		Present = qsfpIG.Present[1]
	}

	//swap upper/lower ports to match front panel numbering
	if (port % 2) == 0 {
		port++
	} else {
		port--
	}

	pmask := uint16(1) << (port % 16)
	if (Present&pmask)>>(port%16) == 1 {
		return "empty"
	}
	return "installed"
}

func (h *I2cDev) QsfpReset(ports uint16, reset uint16) {

	//if module was removed or inserted into a port, set reset line accordingly
	r := getRegs()
	if (ports & 0xff) != 0 {
		r.Output[0].get(h)
		DoI2cRpc()
		v := uint8((s[1].D[0] & uint8((ports&0xff)^0xff)) | uint8((ports&reset)&0xff))
		r.Output[0].set(h, v)
	}
	if (ports & 0xff00) != 0 {
		r.Output[1].get(h)
		DoI2cRpc()
		v := uint8((s[1].D[0] & uint8(((ports&0xff00)>>8)^0xff)) | uint8(((ports&reset)&0xff00)>>8))
		r.Output[1].set(h, v)
	}
	DoI2cRpc()
	return
}

func (h *I2cDev) QsfpLpMode(ports uint16, reset uint16) {

	//if module was removed or inserted into a port, set reset line accordingly
	r := getRegs()
	if (ports & 0xff) != 0 {
		r.Output[0].get(h)
		DoI2cRpc()
		v := uint8((s[1].D[0] & uint8((ports&0xff)^0xff)) | uint8((ports&reset)&0xff))
		r.Output[0].set(h, v)
	}
	if (ports & 0xff00) != 0 {
		r.Output[1].get(h)
		DoI2cRpc()
		v := uint8((s[1].D[0] & uint8(((ports&0xff00)>>8)^0xff)) | uint8(((ports&reset)&0xff00)>>8))
		r.Output[1].set(h, v)
	}
	DoI2cRpc()
	return
}

func (h *I2cDev) QsfpInit(out0 byte, out1 byte, conf0 byte, conf1 byte) {
	//all ports default in reset
	r := getRegs()
	r.Output[0].set(h, out0)
	r.Output[1].set(h, out1)
	DoI2cRpc()
	r.Config[0].set(h, conf0)
	r.Config[1].set(h, conf1)
	DoI2cRpc()
	return
}
