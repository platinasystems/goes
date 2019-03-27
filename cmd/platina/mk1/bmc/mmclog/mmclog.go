// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mmclog

import (
	"fmt"
	"os"
	"strconv"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/flags"
	"github.com/platinasystems/parms"
)

const (
	LOGA      = "/mnt/dmesg.txt"
	LOGB      = "/mnt/dmesg2.txt"
	DfltByte  = "0"
	DfltCount = "25"
	linesize  = 160
)

type Command struct{}

func (Command) String() string { return "mmclog" }

func (Command) Usage() string { return "mmclog [-b BYTE] [-c COUNT] -2" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "display persistant MMC dmesg log",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
        The mmclog command displays MMC dmesg log

	The -b parameter specifies starting byte number.
	The -c parameter specifies number of lines to display.
	The -2 flag displays the secondary(older) dmesg log, if available.

	The default is to display the last 25 lines of the primary log.`,
	}
}

func (Command) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-2")
	parm, args := parms.New(args, "-b", "-c")
	if len(parm.ByName["-b"]) == 0 {
		parm.ByName["-b"] = DfltByte
	}
	if len(parm.ByName["-c"]) == 0 {
		parm.ByName["-c"] = DfltCount
	}
	log := LOGA
	if flag.ByName["-2"] {
		log = LOGB
	}
	if _, err := os.Stat(log); os.IsNotExist(err) {
		fmt.Println("log file: ", log, "does not exist")
		return nil
	}
	if err = dspSiz(log); err != nil {
		return err
	}

	displ, err := strconv.Atoi(parm.ByName["-b"])
	if err != nil {
		return err
	}
	count, err := strconv.Atoi(parm.ByName["-c"])
	if err != nil {
		return err
	}
	tail := false
	if displ == 0 {
		tail = true
	}
	if err = dspLog(log, displ, count, tail); err != nil {
		return err
	}
	return nil
}

func dspLog(log string, displ int, count int, tail bool) (err error) {
	nomsize := count * linesize
	f, err := os.Open(log)
	if err != nil {
		return err
	}
	defer f.Close()

	if tail {
		fi, err := f.Stat()
		if err != nil {
			panic(err)
		}
		l := int(fi.Size()) - (count * linesize)
		if l > 0 {
			displ = l
		} else {
			displ = 0
		}
	}

	_, err = f.Seek(int64(displ), 0)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, nomsize)
	n, err := f.Read(buf)
	if err != nil {
		panic(err)
	}

	i := 0
	if tail {
		c := 0
		for j := len(buf) - 1; j > 0; j-- {
			if string(buf[j]) == "\n" {
				c++
				if c == (count + 1) {
					i = j + 1
				}
			}
		}
	}

	l := 0
	for j := i; j < n; j++ {
		if string(buf[j]) == "\n" {
			if i != 0 {
				fmt.Print("byte=", displ+i, " seq#=")
				for _, c := range buf[i:j] {
					fmt.Printf("%c", c)
				}
				fmt.Println()
			}
			i = j + 1
			l++
		}
		if !tail && l > count {
			break
		}
	}
	f.Close()
	return nil
}

func dspSiz(log string) (err error) {
	f, err := os.Open(log)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	logsz := fi.Size()
	f.Close()

	fmt.Println("\nlog: ", log, "  size: ", logsz)
	return nil
}
