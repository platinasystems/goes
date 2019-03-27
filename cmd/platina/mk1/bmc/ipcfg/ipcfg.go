// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipcfg

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/parms"
)

const (
	Machine = "platina-mk1-bmc"
)

var command *Command

type Command struct {
	Gpio func()
	gpio sync.Once
}

func (Command) String() string { return "ipcfg" }

func (Command) Usage() string { return "ipcfg [-ip]" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "set the bmc ip address, via bootargs ip=",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
        The ipcfg command sets the bmc ip address.

		The ip address will be updated in both QSPI0 and QSPI1.

		The value is stored in the "ip" string in u-boot "bootargs"
		environment variable.

		Example:
		ipcfg -ip 192.168.101.241::192.168.101.2:255.255.255.0::eth0:on`,
	}
}

func (c *Command) Main(args ...string) (err error) {
	command = c
	initQfmt()
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
	e, bootargs, err := GetEnv(q)
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
	if err = UpdateEnv(q); err != nil {
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
	PutPer([]byte(s), q)
	if err != nil {
		return err
	}
	return nil
}
