// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ucd9090

import (
	"fmt"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "ucd9090"
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
		pub <- fmt.Sprint("vmon.5v.sb: ", pm.Vout(1))
		pub <- fmt.Sprint("vmon.3v8.bmc: ", pm.Vout(2))
		pub <- fmt.Sprint("vmon.3v3.sys: ", pm.Vout(3))
		pub <- fmt.Sprint("vmon.3v3.bmc: ", pm.Vout(4))
		pub <- fmt.Sprint("vmon.3v3.sb: ", pm.Vout(5))
		pub <- fmt.Sprint("vmon.1v0.thc: ", pm.Vout(6))
		pub <- fmt.Sprint("vmon.1v8.sys: ", pm.Vout(7))
		pub <- fmt.Sprint("vmon.1v25.sys: ", pm.Vout(8))
		pub <- fmt.Sprint("vmon.1v2.ethx: ", pm.Vout(9))
		pub <- fmt.Sprint("vmon.1v0.tha: ", pm.Vout(10))
	} else {
		pub <- fmt.Sprint("vmon.5v.sb: ", pm.Vout(1))
		pub <- fmt.Sprint("vmon.3v8.bmc: ", pm.Vout(2))
		pub <- fmt.Sprint("vmon.3v3.sys: ", pm.Vout(3))
		pub <- fmt.Sprint("vmon.3v3.bmc: ", pm.Vout(4))
		pub <- fmt.Sprint("vmon.3v3.sb: ", pm.Vout(5))
		pub <- fmt.Sprint("vmon.1v0.thc: ", pm.Vout(6))
		pub <- fmt.Sprint("vmon.1v8.sys: ", pm.Vout(7))
		pub <- fmt.Sprint("vmon.1v25.sys: ", pm.Vout(8))
		pub <- fmt.Sprint("vmon.1v2.ethx: ", pm.Vout(9))
		pub <- fmt.Sprint("vmon.1v0.tha: ", pm.Vout(10))
	}
	return nil
}
