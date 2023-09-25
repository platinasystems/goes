// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux,amd64

package biosupdate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "biosupdate" }

func (Command) Usage() string {
	return "biosupdate [-h|-V|[-s <slot>][-E|(-r|-w|-v) <file>]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "read/write the BIOS image in SPI0 or SPI1",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	This command reads and writes the BIOS image in SPI0 or SPI1.
	  -r | --read <file>                 Read BIOS image from flash and save to <file>
	  -w | --write <file>                Write BIOS image <file> to flas
	  -v | --verify <file>               Verify BIOS image against <file>
	  -E | --erase                       Erase BIOS image from flash memory
	  -s | --spi <num>                   Select SPI 0 or 1

	You can specify one of -h, -V, -E, -r, -w, -v or no operation.
	If no operation is specified, then the programmer will be tested.`,
	}
}

const (
	op_none = iota
	op_read
	op_write
	op_verify
	op_erase
)

func (Command) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-E")
	parm, args := parms.New(args, "-r", "-w", "-v", "-s")

	var spinum uint64
	op := op_none

	if len(parm.ByName["-s"]) != 0 {
		spinum, err = strconv.ParseUint(parm.ByName["-s"], 0, 1)
		if err != nil {
			return fmt.Errorf("%s: %v", parm.ByName["-s"], err)
		}
	}

	if flag.ByName["-E"] {
		op = op_erase
	}

	if len(parm.ByName["-r"]) != 0 {
		if op != op_none {
			return fmt.Errorf("Multiple operations specified.")
		}
		op = op_read
	}

	if len(parm.ByName["-w"]) != 0 {
		if op != op_none {
			return fmt.Errorf("Multiple operations specified.")
		}
		op = op_write
	}

	if len(parm.ByName["-v"]) != 0 {
		if op != op_none {
			return fmt.Errorf("Multiple operations specified.")
		}
		op = op_verify
	}

	switch op {
	case op_read:
		err = doRead(parm.ByName["-r"], uint(spinum))
	case op_write:
		err = doWrite(parm.ByName["-w"], uint(spinum))
	case op_erase:
		err = doErase(uint(spinum))
	case op_verify:
		err = doVerify(parm.ByName["-v"], uint(spinum))
	default:
		err = doFindProgrammer()
	}

	return
}

func doFindProgrammer() (err error) {
	var programmer Programmer

	if err = programmer.Open(0); err != nil {
		return
	}

	defer programmer.Close()

	fmt.Printf("BIOS Programmer found.\n")
	return
}

func showAddressProgress(opName string, c chan float32) {
	for a := range c {
		fmt.Printf("%s......%3.0f%%\r", opName, a)
	}

	fmt.Printf("%s......done.\n", opName)
}

func doRead(filename string, spinum uint) (err error) {
	fmt.Printf("Reading BIOS from SPI%d to file %s.\n", spinum, filename)

	var programmer Programmer

	if err = programmer.Open(spinum); err != nil {
		return
	}

	defer programmer.Close()

	d := make([]byte, programmer.TotalFlashSize())
	base := (programmer.BiosBase())
	limit := (programmer.BiosLimit())

	c := make(chan float32)
	var n int
	go func() {
		n, err = programmer.ReadAt(d[base:(limit+1)], int64(base), c)
		close(c)
		return
	}()

	showAddressProgress("Reading", c)

	if err != nil {
		return
	}
	if uint32(n) != programmer.BiosImageSize() {
		return fmt.Errorf("%d bytes read, expecting %d bytes", n, programmer.BiosImageSize())
	}

	err = ioutil.WriteFile(filename, d, 0644)

	return
}

func doVerify(filename string, spinum uint) (err error) {
	fmt.Printf("Verifying BIOS from SPI%d with file %s.\n", spinum, filename)

	var programmer Programmer

	if err = programmer.Open(spinum); err != nil {
		return
	}

	defer programmer.Close()

	d1 := make([]byte, programmer.TotalFlashSize())
	base := (programmer.BiosBase())
	limit := (programmer.BiosLimit())

	c := make(chan float32)
	var n int
	go func() {
		n, err = programmer.ReadAt(d1[base:(limit+1)], int64(base), c)
		close(c)
		return
	}()

	showAddressProgress("Reading", c)

	if err != nil {
		return
	}
	if uint32(n) != programmer.BiosImageSize() {
		return fmt.Errorf("%d bytes read, expecting %d bytes", n, programmer.BiosImageSize())
	}

	fmt.Printf("Verifying......")
	d0, err := ioutil.ReadFile(filename)
	if len(d0) != len(d1) {
		return fmt.Errorf("Read %d bytes, expecting %d bytes", len(d0), len(d1))
	}

	if bytes.Equal(d0, d1) {
		fmt.Printf("Valid.\n")
	} else {
		fmt.Printf("Invalid!\n")
	}

	return
}

func doErase(spinum uint) (err error) {
	fmt.Printf("Erasing BIOS from SPI%d.\n", spinum)

	var programmer Programmer

	if err = programmer.Open(spinum); err != nil {
		return
	}

	defer programmer.Close()

	base := programmer.BiosBase()
	n := programmer.BiosImageSize()

	err = doEraseAndVerify(&programmer, base, n)

	return
}

func doWrite(filename string, spinum uint) (err error) {
	var programmer Programmer

	if err = programmer.Open(spinum); err != nil {
		return
	}

	defer programmer.Close()

	d, err := ioutil.ReadFile(filename)
	if uint32(len(d)) != programmer.TotalFlashSize() {
		err = fmt.Errorf("File is %d bytes, expecting %d bytes", len(d), programmer.TotalFlashSize())
		return
	}

	base := programmer.BiosBase()
	limit := programmer.BiosLimit()
	size := programmer.BiosImageSize()

	err = doEraseAndVerify(&programmer, base, size)
	if err != nil {
		return
	}

	c := make(chan float32)
	var n int
	go func() {
		n, err = programmer.WriteAt(d[base:(limit+1)], int64(base), c)
		close(c)
		return
	}()

	showAddressProgress(fmt.Sprintf("Writing BIOS in SPI%d with file %s", spinum, filename), c)

	if err != nil {
		return
	}
	if uint32(n) != programmer.BiosImageSize() {
		return fmt.Errorf("%d bytes written, expecting %d bytes", n, size)
	}

	return
}

func doEraseAndVerify(programmer *Programmer, base uint32, size uint32) (err error) {
	spinum := programmer.CurrentSPINum()

	erased, err := doVerifyErase(programmer, base, size)
	if err != nil || erased {
		return
	}

	c := make(chan float32)
	go func() {
		err = programmer.EraseFlash(base, size, c)
		close(c)
		return
	}()

	showAddressProgress(fmt.Sprintf("Erasing BIOS in SPI%d", spinum), c)

	if err != nil {
		fmt.Printf("Error erasing flash.\n")
		return
	}

	erased, err = doVerifyErase(programmer, base, size)
	if err != nil {
		return
	}

	if !erased {
		return fmt.Errorf("Unable to erase flash!")
	}

	return
}

func doVerifyErase(programmer *Programmer, base uint32, size uint32) (erased bool, err error) {
	spinum := programmer.CurrentSPINum()

	c := make(chan float32)
	go func() {
		erased, err = programmer.CheckErased(base, size, c)
		close(c)
		return
	}()

	showAddressProgress(fmt.Sprintf("Checking BIOS in SPI%d", spinum), c)

	if err != nil {
		fmt.Printf("Error Verifying Erase!\n")
		return
	}

	if erased {
		fmt.Printf("Erased.\n")
	} else {
		fmt.Printf("BIOS not erased.\n")
	}

	return
}
