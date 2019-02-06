// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipcfg

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/platinasystems/goes/cmd/platina/mk1/bmc/upgrade"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/parms"
)

const (
	Name    = "ipcfg"
	Apropos = "set bmc ip address, via bootargs ip="
	Usage   = "ipcfg [-ip]"
	Man     = `
DESCRIPTION
        The ipcfg command sets bmc ip address, via bootargs ip="`
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
	parm, args := parms.New(args, "-ip")
	if len(parm.ByName["-ip"]) == 0 {
		if err = dispIP(false); err != nil {
			return err
		}
		if err = dispIP(true); err != nil {
			return err
		}
		return
	} else {
		s := parm.ByName["-ip"]
		n := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
		r := n + "\\." + n + "\\." + n + "\\." + n + "::eth0:on"
		x := regexp.MustCompile(r)
		if parm.ByName["-ip"] == "dhcp" || x.MatchString(s) {
			if err = updateIP(s, false); err != nil {
				fmt.Println(err)
			}
			if err = updateIP(s, true); err != nil {
				fmt.Println(err)
			}
		} else {
			err = fmt.Errorf("invalid ip string")
			return err
		}
	}
	return nil
}

func dispIP(q bool) error {
	e, bootargs, err := upgrade.GetEnv(q)
	if err != nil {
		return err
	}
	n := strings.SplitAfter(e[bootargs], "ip=")
	if !q {
		fmt.Println("QSPI0:  ip=" + n[1])
	} else {
		fmt.Println("QSPI1:  ip=" + n[1])
	}
	return nil
}

func updateIP(ip string, q bool) (err error) {
	if err = updatePer(ip, q); err != nil {
		return err
	}
	if err = upgrade.UpdateEnv(q); err != nil {
		return err
	}
	return nil
}

func updatePer(ip string, q bool) (err error) {
	if !q {
		fmt.Println("Updating QSPI0 persistent block")
	} else {
		fmt.Println("Updating QSPI1 persistent block")
	}
	s := ip + "\x00"
	upgrade.PutPer([]byte(s), q)
	if err != nil {
		return err
	}
	return nil
}
