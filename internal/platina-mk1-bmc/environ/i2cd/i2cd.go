// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package i2cd

import (
	"bufio"
	"os"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	//	"github.com/platinasystems/go/internal/redis"
	"github.com/golang/go/src/pkg/encoding/gob"
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

const (
	NONE  = 0
	DO    = 1
	DOMUX = 2
)
const (
	REG8   = 1
	REG16  = 2
	REG16R = 3
)
const MAXTRANS = 10

type I struct {
	Op        int
	RegType   int
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [4]byte
	ErrCode   error
	Bus       int
	Addr      int
	MuxBus    int
	MuxAddr   int
	MuxValue  int
}
type R struct {
	F float64
	I uint8
	J uint16
	S string
	E error
}

var b = [4]byte{0, 0, 0, 0}
var i = I{NONE, 0, 0, 0, 0, b, nil, 0, 0, 0, 0, 0}
var j [MAXTRANS]I
var x int
var r = R{float64(0), uint8(0), uint16(0), "string", nil}
var s [MAXTRANS]R

func clearJS() {
	x = 0
	for k := 0; k < MAXTRANS; k++ {
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
		if size == 319 {
			break
		}
		log.Print("i2cd SIZE: ", size)
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

	for x := 0; x < MAXTRANS; x++ { //loop over all i2c req
		//log.Print(x, j[x].Op)
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
