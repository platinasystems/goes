// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package eeprom

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const Name = "eeprom"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return "eeprom [-n] [-FIELD | FIELD=VALUE]..." }

func (cmd cmd) Complete(args ...string) []string {
	a := append(cmd.flags(), cmd.parms()...)
	sort.Strings(a)
	return a
}

func (cmd) flags() []string {
	a := make([]string, 1+len(Types))
	a[0] = "-n"
	for i, t := range Types {
		a[1+i] = fmt.Sprint("-", t)
	}
	return a
}

func (cmd) parms() []string {
	a := make([]string, 2+len(Types))
	a[0] = "Onie.Data"
	a[1] = "Onie.Version"
	for i, t := range Types {
		a[2+i] = t.String()
	}
	return a
}

func (cmd cmd) Main(args ...string) error {
	var eeprom Eeprom
	nargs := len(args)
	flag, args := flags.New(args, cmd.flags()...)
	parm, args := parms.New(args, cmd.parms()...)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	buf, err := Vendor.ReadBytes()
	if err != nil {
		return err
	}
	_, err = eeprom.Write(buf)
	if err != nil {
		return err
	}
	for k, t := range flag {
		if k != "-n" && t {
			eeprom.Del(k[1:])
		}
	}
	for k, s := range parm {
		if len(s) == 0 {
			continue
		}
		if err = eeprom.Set(k, s); err != nil {
			return err
		}

	}
	os.Stdout.WriteString(eeprom.String())
	if nargs == 0 || flag["-n"] {
		return nil
	}
	if !WriteEnable {
		return fmt.Errorf("write disabled")
	}
	clone, err := eeprom.Clone()
	if err != nil {
		return err
	}
	err = eeprom.Equal(clone)
	if err != nil {
		return err
	}
	fmt.Print("Write in  ")
	for i := 5; i > 0; i-- {
		fmt.Print("\b", i)
		time.Sleep(time.Second)
	}
	fmt.Print("\rWriting...")
	signal.Ignore(os.Interrupt)
	_, err = Vendor.Write(eeprom.Bytes())
	fmt.Print("\r          \r")
	return err
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "show, delete or modify eeprom fields",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	eeprom - show, delete or modify eeprom fields

SYNOPSIS
	eeprom [-n] [-vendor-extension] [-FIELD | FIELD=VALUE]...

DESCRIPTION
	Show, delete or modify system eeprom fields.

	-n     dry-run to show modifications

	-vendor-extension 
		set, modify, or delete vendor sub-fields

	Without any args, show current eeprom configuation.`,
	}
}
