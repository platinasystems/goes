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

// cache
var flags_, parms_, vendorFlags_, vendorParms_ []string

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return "eeprom [-n] [-FIELD | FIELD=VALUE]..." }

func (cmd cmd) Complete(args ...string) []string {
	a := append(cmd.flags(), cmd.parms()...)
	a = append(a, cmd.vendorFlags()...)
	a = append(a, cmd.vendorParms()...)
	sort.Strings(a)
	return a
}

func (cmd) flags() []string {
	if len(flags_) == 0 {
		a := make([]string, 1+len(NamesByType))
		a[0] = "-n"
		i := 1
		for _, name := range NamesByType {
			a[i] = name
			i++
		}
		flags_ = a
	}
	return flags_
}

func (cmd) parms() []string {
	if len(parms_) == 0 {
		a := make([]string, 2+len(Vendor.Extension.NamesByType))
		a[0] = "Onie.Data"
		a[1] = "Onie.Version"
		i := 2
		for _, name := range Vendor.Extension.NamesByType {
			a[i] = name
			i++
		}
		parms_ = a
	}
	return parms_
}

func (cmd) vendorFlags() []string {
	if len(vendorFlags_) == 0 {
		a := make([]string, len(Vendor.Extension.NamesByType))
		i := 0
		for _, name := range Vendor.Extension.NamesByType {
			a[i] = name
			i++
		}
		vendorFlags_ = a
	}
	return vendorFlags_
}

func (cmd) vendorParms() []string {
	if len(vendorParms_) == 0 {
		a := make([]string, len(Vendor.Extension.NamesByType))
		i := 0
		for _, name := range Vendor.Extension.NamesByType {
			a[i] = name
			i++
		}
		vendorParms_ = a
	}
	return vendorParms_
}

func (cmd cmd) Main(args ...string) error {
	var eeprom Eeprom
	flag, args := flags.New(args, cmd.flags()...)
	parm, args := parms.New(args, cmd.parms()...)
	vendorFlag, args := flags.New(args, cmd.vendorFlags()...)
	vendorParm, args := parms.New(args, cmd.vendorParms()...)
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
	if len(args) == 0 {
		os.Stdout.WriteString(eeprom.String())
		return nil
	}
	for _, k := range cmd.flags() {
		if flag[k] {
			eeprom.Del(k[1:])
		}
	}
	for _, k := range cmd.parms() {
		s := parm[k]
		if len(s) == 0 {
			continue
		}
		eeprom.Set(k, s)
		if err != nil {
			return err
		}

	}
	ve := eeprom.Tlv[VendorExtensionType]
	for _, k := range cmd.vendorFlags() {
		if vendorFlag[k] {
			k = k[1:]
			method, found := ve.(Deler)
			if !found {
				return fmt.Errorf("%s: missing deleter", k)
			}
			method.Del(k)
		}
	}
	for _, k := range cmd.vendorParms() {
		s := vendorParm[k]
		if len(s) == 0 {
			continue
		}
		method, found := ve.(Setter)
		if !found {
			return fmt.Errorf("%s: missing setter", k)
		}
		err = method.Set(k, s)
		if err != nil {
			return err
		}
	}
	os.Stdout.WriteString(eeprom.String())
	if flag["-n"] {
		return nil
	}
	if !WriteEnable {
		return fmt.Errorf("write disabled")
	}
	clone, err := eeprom.Clone()
	if err != nil {
		return nil
	}
	err = eeprom.Equal(clone)
	if err != nil {
		return nil
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
