// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipcfg

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bmc/upgrade"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "ipcfg"
	Apropos = "set bmc ip address, via bootargs ip="
	Usage   = "ipcfg [-ip]"
	Man     = `
DESCRIPTION
        The ipcfg command sets bmc ip address, via bootargs ip="`
	ENVSIZE = 8192
	ENVCRC  = 4
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
		if err = updatePer(parm.ByName["-ip"], false); err != nil {
			return err
		}
		if err = UpdateEnv(false); err != nil {
			return err
		}
		if err = updatePer(parm.ByName["-ip"], true); err != nil {
			return err
		}
		if err = UpdateEnv(true); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func dispIP(q bool) error {
	m, bootargs, err := getEnv(q)
	if err != nil {
		return err
	}
	n := strings.SplitAfter(m[bootargs], "ip=")
	if !q {
		fmt.Println("QSPI0:  ip=" + n[1])
	} else {
		fmt.Println("QSPI1:  ip=" + n[1])
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
	putPer([]byte(s), q)
	if err != nil {
		return err
	}
	return nil
}

func UpdateEnv(q bool) (err error) {
	b, err := getPer(q)
	s := strings.Split(string(b), "\x00")
	ip := s[0]
	if len(string(ip)) > 500 {
		err = fmt.Errorf("no 'ip=' in per, skipping env update")
		return err
	}
	m, bootargs, err := getEnv(q)
	if err != nil {
		return err
	}
	if !strings.Contains(m[bootargs], "ip=") {
		err = fmt.Errorf("no 'ip=' in env, skipping env update")
		return err
	}
	n := strings.SplitAfter(m[bootargs], "ip=")
	if n[1] == string(ip) {
		err = fmt.Errorf("ip-env == ip-per, skipping env update")
		return err
	}
	m[bootargs] = n[0] + string(ip)
	err = putEnv(m, q)
	if err != nil {
		return err
	}
	return nil
}

func getEnv(q bool) (env []string, bootargs int, err error) {
	b, err := upgrade.ReadBlk("env", q)
	if err != nil {
		return nil, 0, err
	}
	m := strings.Split(string(b[ENVCRC:ENVSIZE]), "\x00")
	var end int
	for j, n := range m {
		if strings.Contains(n, "bootargs") {
			bootargs = j
		}
		if len(n) == 0 {
			end = j
			break
		}
	}
	return m[:end], bootargs, nil
}

func putEnv(s []string, q bool) (err error) {
	fmt.Println("SKIP PUT ENV FOR NOW")
	return nil
	//calc new crc, write env //FIXME
	//err = upgrade.WriteBlk("env", b, q)
	//if err != nil {
	//	return err
	//}
	return nil
}

func getPer(q bool) (b []byte, err error) {
	b, err = upgrade.ReadBlk("per", q)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func putPer(b []byte, q bool) (err error) {
	err = upgrade.WriteBlk("per", b, q)
	if err != nil {
		return err
	}
	return nil
}
