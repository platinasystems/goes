// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mmclog

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "mmclog"
	Apropos = "display persistant MMC dmesg log"
	Usage   = "mmclog [-l LINE] [-c COUNT] -2"
	Man     = `
DESCRIPTION
        The mmclog command displays MMC dmesg log

	The -l parameter specifies starting line number.
	The -c parameter specifies number of lines to display.
	The -2 flag displays the secondary(older) dmesg log, if available.

	The default is to display the last 25 lines of the primary log.`

	LOGA      = "/mnt/dmesg.txt"
	LOGB      = "/mnt/dmesg2.txt"
	DfltLine  = "0"
	DfltCount = "25"
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

type cmd struct{}

func New() Interface { return cmd{} }

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func (cmd) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-2")
	parm, args := parms.New(args, "-l", "-c")
	if len(parm.ByName["-l"]) == 0 {
		parm.ByName["-l"] = DfltLine
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

	max := 0
	line := 0
	count := 0
	if max, err = dspSiz(log); err != nil {
		return err
	}
	if line, err = strconv.Atoi(parm.ByName["-l"]); err != nil {
		return err
	}
	if count, err = strconv.Atoi(parm.ByName["-c"]); err != nil {
		return err
	}
	if line == 0 {
		line = max - count + 1
	}
	if err = dspLog(log, line, count); err != nil {
		return err
	}
	return nil
}

func dspLog(log string, line int, count int) (err error) {
	f, err := os.Open(log)
	if err != nil {
		return err
	}
	defer f.Close()
	thisLine := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		thisLine++
		if thisLine >= line && thisLine < (line+count) {
			fmt.Println(thisLine, sc.Text())
		}
	}
	f.Close()
	return nil
}

func dspSiz(logname string) (max int, err error) {
	f, err := os.Open(logname)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return 0, err
	}
	siz := fi.Size()
	sc := bufio.NewScanner(f)
	cnt := 0
	for sc.Scan() {
		cnt++
	}
	f.Close()
	fmt.Println("\nlog: ", logname, "  size: ", siz, "  lines: ", cnt)
	return cnt, nil
}
