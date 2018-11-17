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

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

type Command struct {
	Config func()
}

func (Command) String() string { return "eeprom" }

func (Command) Usage() string {
	return "eeprom [-n] [-FIELD | FIELD=VALUE]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "show, delete or modify eeprom fields",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Show, delete or modify system eeprom fields.

	-n     dry-run to show modifications

	-vendor-extension 
		set, modify, or delete vendor sub-fields

	Without any args, show current eeprom configuation.`,
	}
}

func (c Command) Complete(args ...string) []string {
	var a []string
	for _, v := range c.flags() {
		a = append(a, v.(string))
	}
	for _, v := range c.parms() {
		a = append(a, v.(string))
	}
	sort.Strings(a)
	return a
}

func (c Command) Main(args ...string) error {
	var eeprom Eeprom
	if c.Config != nil {
		c.Config()
	}
	nargs := len(args)
	flag, args := flags.New(args, c.flags()...)
	parm, args := parms.New(args, c.parms()...)
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
	for k, t := range flag.ByName {
		if k != "-n" && t {
			eeprom.Del(k[1:])
		}
	}
	for k, s := range parm.ByName {
		if len(s) == 0 {
			continue
		}
		if err = eeprom.Set(k, s); err != nil {
			return err
		}

	}
	os.Stdout.WriteString(eeprom.String())
	if nargs == 0 || flag.ByName["-n"] {
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

func (Command) flags() []interface{} {
	a := make([]interface{}, 1+len(Types))
	a[0] = "-n"
	for i, t := range Types {
		a[1+i] = fmt.Sprint("-", t)
	}
	return a
}

func (Command) parms() []interface{} {
	a := make([]interface{}, 2+len(Types))
	a[0] = "Onie.Data"
	a[1] = "Onie.Version"
	for i, t := range Types {
		a[2+i] = t.String()
	}
	return a
}
