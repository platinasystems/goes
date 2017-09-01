// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO AUTHENTICATE,  CERT OR EQUIV
//TODO UPGRADE AUTOMATICALLY IF ENABLED, contact boot server

package upgrade

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "upgrade"
	Apropos = "upgrade images"
	Usage   = "upgrade [-v VER] [-s SRVR[/dir]] [-r] [-l] [-c] [-t] [-u -d -e -k -i -a] [-f]"
	Man     = `
DESCRIPTION
	The upgrade command updates firmware images.

	The default upgrade version is "LATEST". 
	Or specify a version using "-v", in the form v0.0[.0][.0]

	The -l flag lists available upgrade versions.

	The -r flag prints a report on version numbers.

	By default, images are downloaded from "downloads.platina.com".
	Or specify a server using "-s" followed by a URL or IPv4 address.

	Upgrades only happen if the version numbers differ,
	unless overridden with the "-f" force flag.

OPTIONS
	-v [VER]          version number or hash, default is LATEST
	-s [SERVER[/dir]] IP4 or URL, default is downloads.platina.com
	-t                use TFTP instead of HTTP
	-l                shows list of available versions for upgrade
	-r                report current versions, booted from qspi0/1
	-c                check checksums of flash
	-u                upgrade u-boot bootloader
	-d                upgrade DTB device tree
	-e                upgrade environment for u-boot
	-k                upgrade kernel
	-i                upgrade initrd/goes
	-f                force upgrade (ignore version check)`

	DfltMod = 0755
	DfltSrv = "downloads.platinasystems.com"
	DfltVer = "LATEST"
	Machine = "platina-mk1-bmc"

	//names of server files
	ArchiveName = "platina-mk1-bmc.zip"
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-t", "-l", "-f", "-r",
		"-c", "-u", "-d", "-e", "k", "i", "-a")
	parm, args := parms.New(args, "-v", "-s")
	if len(parm.ByName["-v"]) == 0 {
		parm.ByName["-v"] = DfltVer
	}
	if len(parm.ByName["-s"]) == 0 {
		parm.ByName["-s"] = DfltSrv
	}

	if flag.ByName["-a"] {
		flag.ByName["-u"] = true
		flag.ByName["-d"] = true
		flag.ByName["-e"] = true
		flag.ByName["-k"] = true
		flag.ByName["-i"] = true
	}
	if flag.ByName["-l"] == false &&
		flag.ByName["-r"] == false && flag.ByName["-u"] == false &&
		flag.ByName["-d"] == false && flag.ByName["-e"] == false &&
		flag.ByName["-k"] == false && flag.ByName["-i"] == false {
		flag.ByName["-a"] = true
	}

	if flag.ByName["-l"] {
		if err := showList(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
			return err
		}
		return nil
	}
	if flag.ByName["-r"] {
		if err := reportVersions(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
			return err
		}
		return nil
	}
	if flag.ByName["-c"] {
		if err := checkChecksums(); err != nil {
			return err
		}
		return nil
	}

	//FIXME Remove on next commit
	flag.ByName["-f"] = true
	flag.ByName["-a"] = true

	if err := doUpgrade(parm.ByName["-s"], parm.ByName["-v"],
		flag.ByName["-t"], flag.ByName["-u"], flag.ByName["-d"],
		flag.ByName["-e"], flag.ByName["-k"], flag.ByName["-i"],
		flag.ByName["-a"], flag.ByName["-f"]); err != nil {
		return err
	}
	return nil
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
var Reboot_flag bool = false

func showList(s string, v string, t bool) error {
	fn := "LIST"
	_, err := getFile(s, v, t, fn)
	if err != nil {
		return err
	}
	l, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	fmt.Print(string(l))
	return nil
}

func reportVersions(s string, v string, t bool) (err error) {
	vr := make([]string, 5, 5)
	vs := make([]string, 5, 5)
	for i, j := range img {
		vr[i] = getVer(j)
		vs[i], err = getSrvVer(s, v, t, j)
		if err != nil {
			return err
		}
	}
	prVer(vr, vs)
	return nil
}

func checkChecksums() error { //FIXME
	return nil
}

func doUpgrade(s string, v string, t bool, u bool, d bool,
	e bool, k bool, i bool, a bool, f bool) error {
	fmt.Print("\n")
	fn := ArchiveName
	n, err := getFile(s, v, t, fn)
	if err != nil {
		return fmt.Errorf("Error downloading: %v", err)
	}
	if n < 1000 {
		return fmt.Errorf("Error file too small: %v", err)
	}
	if err := unzip(); err != nil {
		return fmt.Errorf("Error unzipping file: %v", err)
	}

	if a && f { //FIXME CUT THIS
		if err := writeImageAll(); err != nil {
			return fmt.Errorf("*** UPGRADE ERROR! ***: %v", err)
		}
		Reboot_flag = true
	} else { //FIXME turn into a LOOP
		if u {
			if err := upgradeX(s, v, t, "ubo", f); err != nil {
				return fmt.Errorf("Flash write error: %v", err)
			}
		}
		if d {
			if err := upgradeX(s, v, t, "dtb", f); err != nil {
				return fmt.Errorf("Flash write error: %v", err)
			}
		}
		if e {
			if err := upgradeX(s, v, t, "env", f); err != nil {
				return fmt.Errorf("Flash write error: %v", err)
			}
		}
		if k {
			if err := upgradeX(s, v, t, "ker", f); err != nil {
				return fmt.Errorf("Flash write error: %v", err)
			}
		}
		if i {
			if err := upgradeX(s, v, t, "ini", f); err != nil {
				return fmt.Errorf("Flash write error: %v", err)
			}
			Reboot_flag = true
		}
	}
	if Reboot_flag {
		if err := reboot(); err != nil {
			return err
		}
		return nil
	}
	return nil
}
