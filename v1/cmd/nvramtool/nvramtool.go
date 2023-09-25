// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nvramtool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/nvram"
)

const (
	Name    = "nvramtool"
	Apropos = "nvramtool"
	Usage   = "nvramtool [-y LAYOUT_FILE | -t] PARAMETER ..."
	Man     = `
DESCRIPTION
       Read/write coreboot parameters or show info from coreboot table.

       -y LAYOUT_FILE: Use CMOS layout specified by LAYOUT_FILE.
       -t:             Use CMOS layout specified by CMOS option table.
       -D CMOS_FILE:   Use CMOS file for CMOS data.
       [-n] -r NAME:   Show parameter NAME.  If -n is given, show value only.
       -e NAME:        Show all possible values for parameter NAME.
       -a:             Show names and values for all parameters.
       -w NAME=VALUE:  Set parameter NAME to VALUE.
       -c:     		   Show CMOS checksum.
       -C [VALUE]:     Set checksum to VALUE.
       -Y:             Show CMOS layout info.
       -b OUTPUT_FILE: Dump CMOS memory contents to file.
       -B INPUT_FILE:  Write file contents to CMOS memory.
       -h:             Show this message.
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

var ErrMultipleOperations = errors.New("Multiple operations specified.")
var nv nvram.NVRAM

type opArguments []interface{}
type opFunction func(args ...interface{}) error

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) ShowHelp(args ...interface{}) (err error) {
	fmt.Println(Usage)
	fmt.Println(Man)
	return
}

func (Command) ShowParameter(args ...interface{}) (err error) {
	name := args[0].(string)
	showname := args[1].(bool)

	paramValue, err := nv.ReadCMOSParameter(name)
	if err != nil {
		return
	}

	if showname {
		fmt.Printf("%s = ", name)
	}

	switch paramValue.(type) {
	case string:
		fmt.Printf("%s\n", paramValue.(string))
	case uint64:
		fmt.Printf("0x%X\n", paramValue.(uint64))
	}

	return
}

func (Command) ShowParamterValues(args ...interface{}) (err error) {
	name := args[0].(string)
	e, ok := nv.FindCMOSEntry(name)
	if !ok {
		err = fmt.Errorf("CMOS parameter %s not found.", name)
		return
	}

	switch e.Config() {
	case nvram.CMOSEntryString:
		fmt.Printf("Parameter %s requires a %d-byte string.\n", e.Name(),
			e.Length()/8)
	case nvram.CMOSEntryEnum:
		items, ok := nv.GetCMOSEnumItemsById(e.ConfigId())
		if !ok {
			err = fmt.Errorf("CMOS entry %s has an invalid enum id %d.",
				e.Name(), e.ConfigId())
			return
		}
		if len(items) == 0 {
			err = fmt.Errorf("CMOS entry %s enum id %d has no values.",
				e.Name(), e.ConfigId())
			return
		}
		for _, item := range items {
			fmt.Println(item.Text())
		}
	case nvram.CMOSEntryHex:
		fmt.Printf("Parameter %s requires a %d-bit unsigned integer.\n",
			e.Name(), e.Length())
	case nvram.CMOSEntryReserved:
		fmt.Printf("Parameter %s is reserved.\n", e.Name())
	default:
		return fmt.Errorf("CMOS entry %s has invalid config type.", e.Name())
	}

	return
}

func (c *Command) ShowAllParamtersAndValues(args ...interface{}) (err error) {

	for _, e := range nv.GetCMOSEntriesList() {
		if e.Config() == nvram.CMOSEntryReserved || e.Name() == "check_sum" {
			continue
		}
		err = c.ShowParameter(e.Name(), true)
		if err != nil {
			return
		}
	}

	return
}

func (Command) SetParameter(args ...interface{}) (err error) {
	name := args[0].(string)
	value := args[1].(string)

	paramValue, err := nv.NewParameterType(name)
	if err != nil {
		return
	}

	switch paramValue.(type) {
	case uint64:
		paramValue, err = strconv.ParseUint(value, 0, 64)
		if err != nil {
			return fmt.Errorf("%s is not a valid value.", value)
		}
	case string:
		paramValue = value
	default:
		err = fmt.Errorf("Unknown parameter type.")
		return
	}

	return nv.WriteCMOSParameter(name, paramValue)
}

func (Command) ShowChecksum(args ...interface{}) (err error) {
	var sum uint16
	sum, err = nv.CMOS.ReadChecksum()
	if err != nil {
		return err
	}
	fmt.Printf("0x%X\n", sum)
	return
}

func (Command) WriteChecksum(args ...interface{}) (err error) {
	value := args[0].(string)

	sum, err := strconv.ParseUint(value, 0, 16)
	if err != nil {
		return fmt.Errorf("%s is not a valid 16-bit value.", value)
	}
	err = nv.CMOS.WriteChecksum(uint16(sum))
	return
}

func (Command) ShowLayoutInfo(args ...interface{}) (err error) {
	fmt.Println("entries")
	for _, e := range nv.GetCMOSEntriesList() {
		fmt.Println(e)
	}
	fmt.Println("\nenumerations")
	items := nv.GetCMOSEnumItems()
	for _, item := range items {
		fmt.Println(item)
	}

	fmt.Println("\nchecksums")
	fmt.Println("checksum", nv.GetCheckChecksum())
	return
}

func (Command) DumpCMOSMemory(args ...interface{}) (err error) {
	filename := args[0].(string)
	d, err := nv.CMOS.ReadAllMemory()
	return ioutil.WriteFile(filename, d, 0644)
}

func (Command) WriteCMOSMemory(args ...interface{}) (err error) {
	filename := args[0].(string)
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	return nv.CMOS.WriteAllMemory(d)
}

func (c Command) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-h", "-t", "-n", "-a", "-c", "-Y")
	parm, args := parms.New(args, "-y", "-D", "-r", "-e", "-w", "-C", "-b", "-B")

	var opfunc opFunction
	var opArgs opArguments
	var validateChecksum bool = false

	defer nv.Close()

	// Show help and usage
	if flag.ByName["-h"] {
		c.ShowHelp()
		return nil
	}

	// Initialize NVRAM access
	var layoutFileName = parm.ByName["-y"]
	var cmosMemFileName = parm.ByName["-D"]
	if flag.ByName["-t"] {
		cmosMemFileName = ""
	}

	err = nv.Open(layoutFileName, cmosMemFileName)
	if err != nil {
		return err
	}

	/* Find Operation to run */
	if len(parm.ByName["-r"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}

		opfunc = c.ShowParameter
		validateChecksum = true
		opArgs = opArguments{parm.ByName["-r"], !flag.ByName["-n"]}
	}

	if len(parm.ByName["-e"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}

		opfunc = c.ShowParamterValues
		opArgs = opArguments{parm.ByName["-e"]}
	}

	if flag.ByName["-a"] {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		validateChecksum = true
		opfunc = c.ShowAllParamtersAndValues
	}

	if len(parm.ByName["-w"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}

		fields := strings.FieldsFunc(parm.ByName["-w"],
			func(c rune) bool { return c == '=' })
		if len(fields) != 2 {
			return fmt.Errorf("Invalid input parameter.")
		}

		opArgs = opArguments{fields[0], fields[1]}
		opfunc = c.SetParameter
	}

	if flag.ByName["-c"] {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		opfunc = c.ShowChecksum
	}

	if len(parm.ByName["-C"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		opArgs = opArguments{parm.ByName["-C"]}
		opfunc = c.WriteChecksum
	}

	if flag.ByName["-Y"] {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		opfunc = c.ShowLayoutInfo
	}

	if len(parm.ByName["-b"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		opArgs = opArguments{parm.ByName["-b"]}
		opfunc = c.DumpCMOSMemory
	}

	if len(parm.ByName["-B"]) != 0 {
		if opfunc != nil {
			return ErrMultipleOperations
		}
		opArgs = opArguments{parm.ByName["-B"]}
		opfunc = c.WriteCMOSMemory
	}

	if opfunc == nil || len(args) > 1 {
		c.ShowHelp()
		return
	}

	if err = opfunc(opArgs...); err != nil {
		return
	}

	if validateChecksum {
		err = nv.ValidateChecksum()
	}
	return
}
