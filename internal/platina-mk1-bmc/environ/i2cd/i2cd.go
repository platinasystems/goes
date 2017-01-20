// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package i2cd

import (
	"bufio"
	"encoding/gob"
	"os"
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
	//i2cServer()
	t := time.NewTicker(20 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-cmd:
			return nil
		case <-t.C:
			i2cServer()
		}
	}
	return nil
}

func (cmd cmd) Close() error {
	close(cmd)
	return nil
}

const MAXOPS = 10

type I struct {
	InUse     bool
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [34]byte
	Bus       int
	Addr      int
	Count     int
	Delay     int
	Eeprom    int
}
type R struct {
	D [34]byte
	E error
}

var b = [34]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var i = I{false, i2c.RW(0), 0, 0, b, 0, 0, 0, 0, 0}
var j [MAXOPS]I
var r = R{b, nil}
var s [MAXOPS]R
var x int

func clearJS() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
		s[k] = r
	}
}

func i2cServer() {
	//LOOP OVER ALL FILES

	ServeI2cRpc()

	//NEXT FILE CHECK
}

//CONVERT TO RPC
func ServeI2cRpc() {
	_, err := os.Stat("/tmp/i2c_ti.dat") //file available?
	if os.IsNotExist(err) {
		return
	}
	for {
		fi, er := os.Stat("/tmp/i2c_ti.dat")
		if er != nil {
			log.Print("ti error: ", er)
		}
		size := fi.Size()
		if size == 584 {
			break
		}
		//log.Print("i2cd SIZE: ", size)
	}

	clearJS() //open, reader
	filename := "/tmp/i2c_ti.dat"
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	dec := gob.NewDecoder(reader) //read, decode
	err = dec.Decode(&j)
	if err != nil {
		log.Print("i2c decode error:", err)
	}

	var bus i2c.Bus
	var data i2c.SMBusData
	for x := 0; x < MAXOPS; x++ {
		//log.Print(x, j[x].InUse, j[x].Bus, j[x].Addr, j[x].RegOffset)
		if j[x].InUse == true {
			err = bus.Open(j[x].Bus)
			if err != nil {
				log.Print("ERR1")
				return
			}
			defer bus.Close()

			err = bus.ForceSlaveAddress(j[x].Addr)
			if err != nil {
				log.Print("ERR2")
				return
			}
			data[0] = j[x].Data[0]
			data[1] = j[x].Data[1]
			data[2] = j[x].Data[2]
			data[3] = j[x].Data[3]
			//log.Print("PRE: ", j[x].RW, j[x].RegOffset, j[x].BusSize, data[0])
			err = bus.Do(j[x].RW, j[x].RegOffset, j[x].BusSize, &data)
			if err != nil {
				log.Print("ERR3", err)
				return
			}
			s[x].D[0] = data[0]
			s[x].D[1] = data[1]
			//log.Print("RESULT: ", x, s[x].D[0], s[x].D[1])
		}
	}

	f.Close() //close, remove
	err = os.Remove("/tmp/i2c_ti.dat")
	if err != nil {
		log.Print("could not remove i2c_ti.dat file, error:", err)
		return
	}

	f, err = os.Create("/tmp/i2c_ti.res") //create, writer
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	enc := gob.NewEncoder(w) //encode, write
	err = enc.Encode(&s)
	if err != nil {
		log.Print("err", "i2c encode error: ", err)
	}
	w.Flush()
	f.Close()
}
