// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package i2cd

import (
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
)

const Name = "i2cd"

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

	i2cReq := new(I2cReq)
	rpc.Register(i2cReq)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1233")
	if e != nil {
		log.Print("listen error:", e)
	}
	log.Print("listen OKAY")
	go http.Serve(l, nil)

	t := time.NewTicker(20 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
		}
	}
	return nil
}

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

const MAXOPS = 30

type I struct {
	InUse     bool
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [34]byte
	Bus       int
	Addr      int
	Delay     int
}
type R struct {
	D [34]byte
	E error
}

type I2cReq int

var b = [34]byte{0}
var i = I{false, i2c.RW(0), 0, 0, b, 0, 0, 0}
var j [MAXOPS]I
var r = R{b, nil}
var s [MAXOPS]R
var x int

func (t *I2cReq) ReadWrite(g *[MAXOPS]I, f *[MAXOPS]R) error {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	var bus i2c.Bus
	var data i2c.SMBusData
	for x := 0; x < MAXOPS; x++ {
		if g[x].InUse == true {
			err := bus.Open(g[x].Bus)
			if err != nil {
				log.Print("Error opening I2C bus")
				return err
			}
			defer bus.Close()

			err = bus.ForceSlaveAddress(g[x].Addr)
			if err != nil {
				log.Print("ERR2")
				log.Print("Error setting I2C slave address")
				return err
			}
			data[0] = g[x].Data[0]
			data[1] = g[x].Data[1]
			data[2] = g[x].Data[2]
			data[3] = g[x].Data[3]
			err = bus.Do(g[x].RW, g[x].RegOffset, g[x].BusSize, &data)
			if err != nil {
				log.Print("Error doing I2C R/W")
				return err
			}
			f[x].D[0] = data[0]
			f[x].D[1] = data[1]
			if g[x].BusSize == i2c.I2CBlockData {
				for y := 2; y < 34; y++ {
					f[x].D[y] = data[y]
				}
			}
			bus.Close()
		}
	}
	return nil
}

func clearJS() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
		s[k] = r
	}
}
