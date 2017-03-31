// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package gpio provides utilities to query and dink with general purpose i/o
// pins.
package gpio

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// Lower 16 bits gives index.
	PinIndexMask Pin = 0xffff
	// High bits are flags: direction input/output
	IsInput    Pin = 1 << 31
	IsOutputLo Pin = 1 << 30
	IsOutputHi Pin = 1 << 29
)

type Pin uint32

func (p Pin) Index() int { return int(p & PinIndexMask) }

type GpioAliasMap map[string]string
type PinMap map[string]Pin

type Chip struct {
	// Chip has GPIOs base through base + count.
	Base, Count Pin
	// Value of compatible=XXX node in DTS file for this GPIO chip.
	Compatible map[string]bool
}

var Init = func() {}
var File = "/boot/linux.dtb"
var Aliases GpioAliasMap
var Pins PinMap

// File prefix for testing w/o proper sysfs.
var prefix string

func SetDebugPrefix(p string) { prefix = p }

var GpioBankToBase = map[string]Pin{
	"gpio0": 0,
	"gpio1": 32,
	"gpio2": 64,
	"gpio3": 96,
	"gpio4": 128,
	"gpio5": 160,
	"gpio6": 192,
}

var GpioPinMode = map[string]Pin{
	"output-high": IsOutputHi,
	"output-low":  IsOutputLo,
	"input":       IsInput,
}

func exportPin(p Pin) (err error) {
	fn := prefix + "/sys/class/gpio/export"
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%d\n", p&PinIndexMask)
	return
}

func pinOpen(p Pin, name string) (f *os.File, fn string, err error) {
	fn = fmt.Sprintf(prefix+"/sys/class/gpio/gpio%d/%s", p&PinIndexMask, name)
	f, err = os.OpenFile(fn, os.O_RDWR, 0)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		err = exportPin(p)
		if err != nil {
			return
		}
		// To be safe wait a bit for kernel to create /sys/class/gpio/PIN directory.
		// need this?
		time.Sleep(10 * time.Millisecond)
		f, err = os.OpenFile(fn, os.O_RDWR, 0)
		if err != nil {
			return
		}
	}
	return
}

// "direction" ... reads as either "in" or "out". This value may
// 	normally be written. Writing as "out" defaults to
// 	initializing the value as low. To ensure glitch free
// 	operation, values "low" and "high" may be written to
// 	configure the GPIO as an output with that initial value.
func setGetDir(p Pin, isSet bool) (isOut bool, err error) {
	var (
		f  *os.File
		fn string
	)
	f, fn, err = pinOpen(p, "direction")
	if err != nil {
		return
	}
	defer f.Close()

	var v string
	if isSet {
		v = "in"
		if p&IsOutputLo != 0 {
			v = "low"
		}
		if p&IsOutputHi != 0 {
			v = "high"
		}
		fmt.Fprintf(f, "%s\n", v)
	} else {
		fmt.Fscanf(f, "%s\n", &v)
		switch v {
		case "out":
			isOut = true
		case "in":
		default:
			err = fmt.Errorf("%s: read unexpected value `%s'", fn, v)
			return
		}
	}
	return
}

func (p Pin) SetDirection() (err error) {
	_, err = setGetDir(p, true)
	return
}

func (p Pin) Direction() (isOut bool, err error) {
	isOut, err = setGetDir(p, false)
	return
}

// "value" ... reads as either 0 (low) or 1 (high). If the GPIO
// 	is configured as an output, this value may be written;
// 	any nonzero value is treated as high.
func (p Pin) setGetValue(setValue ...bool) (value bool, err error) {
	var (
		f  *os.File
		fn string
	)
	f, fn, err = pinOpen(p, "value")
	if err != nil {
		return
	}
	defer f.Close()

	var v int
	if len(setValue) > 0 {
		if setValue[0] {
			v = 1
		}
		fmt.Fprintf(f, "%d\n", v)
	} else {
		_, err = fmt.Fscanf(f, "%d\n", &v)
		if err != nil {
			err = fmt.Errorf("%s: parse error %s", fn, err)
			return
		}
	}
	value = v != 0
	return
}

func (p Pin) SetValue(v bool) (err error) {
	_, err = p.setGetValue(v)
	return
}

func (p Pin) Value() (value bool, err error) {
	value, err = p.setGetValue()
	return
}

func pathScanf(path, format string, args ...interface{}) (n int, err error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()
	return fmt.Fscanf(f, format, args...)
}

var (
	chips []Chip
	once  sync.Once
)

func ChipsMatching(pattern string) ([]Chip, error) {
	var (
		err error
		cs  []Chip
	)

	once.Do(func() { chips, err = getChips() })

	if len(pattern) == 0 {
		cs = chips
	} else {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return cs, err
		}
		for i := range chips {
			var ok bool
			if _, ok = chips[i].Compatible[pattern]; ok {
				cs = append(cs, chips[i])
			}
			for k := range chips[i].Compatible {
				if re.MatchString(k) {
					cs = append(cs, chips[i])
					break
				}
			}
		}
	}
	return cs, err
}

func getChips() (chips []Chip, err error) {
	dirPath := prefix + "/sys/class/gpio"
	fis, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}
	for _, fi := range fis {
		if fi.IsDir() {
			var c Chip
			n := fi.Name()
			if _, e := fmt.Sscanf(n, "gpiochip%d", &c.Base); e != nil {
				continue
			}
			if _, err = pathScanf(filepath.Join(dirPath, n, "ngpio"), "%d", &c.Count); err != nil {
				return
			}
			var b []byte
			if b, err = ioutil.ReadFile(filepath.Join(dirPath, n, "device/of_node/compatible")); err != nil {
				return
			}
			cs := strings.Split(string(b), "\x00")
			c.Compatible = make(map[string]bool)
			for i := range cs {
				c.Compatible[cs[i]] = true
			}
			chips = append(chips, c)
		}
	}
	return
}
