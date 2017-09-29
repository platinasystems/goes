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
	Usage   = "upgrade [-v VER] [-s SERVER[/dir]] [-r] [-l] [-t] [-a | -g -k -c] [-f]"
	Man     = `
DESCRIPTION
	The upgrade command updates firmware images.

	The default upgrade version is "LATEST". 
	Or specify a version using "-v", in the form YYYYMMDD

	The -l flag display version of selected server and version.

	The -r flag prints a report on current version numbers.

	By default, images are downloaded from "downloads.platina.com".
	Or from a server using "-s" followed by a URL or IPv4 address.

	Upgrade proceeds only if the selected version number is newer,
	unless overridden with the "-f" force flag.

OPTIONS
	-v [VER]          version [YYYYMMDD] or LATEST (default)
	-s [SERVER[/dir]] IP4 or URL, default is downloads.platina.com
	-t                use TFTP instead of HTTP
	-l                display version of selected server and version
	-r                report current versions of goes, kernel, coreboot
	-g                upgrade goes
	-k                upgrade kernel
	-c                upgrade coreboot
	-a                upgrade all
	-f                force upgrade (ignore version check)`

	DfltMod = 0755
	DfltSrv = "downloads.platinasystems.com"
	DfltVer = "LATEST"
	Machine = "platina-mk1"

	//names of server files
	GoesName      = "goes-platina-mk1" //includes non-compressed tag
	GoesInstaller = "goes-platina-mk1-installer"
	KernelName    = "linux-image-platina-mk1"
	CorebootName  = "coreboot-mk1.rom"
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
		"-g", "-c", "-k", "-a")
	parm, args := parms.New(args, "-v", "-s")
	if len(parm.ByName["-v"]) == 0 {
		parm.ByName["-v"] = DfltVer
	}
	if len(parm.ByName["-s"]) == 0 {
		parm.ByName["-s"] = DfltSrv
	}

	if flag.ByName["-a"] {
		flag.ByName["-g"] = true
		flag.ByName["-k"] = true
		flag.ByName["-c"] = true
	}
	if flag.ByName["-l"] == false &&
		flag.ByName["-r"] == false && flag.ByName["-g"] == false &&
		flag.ByName["-k"] == false && flag.ByName["-c"] == false {
		flag.ByName["-g"] = true
	}

	if flag.ByName["-l"] {
		if err := showList(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
			return err
		}
		return nil
	}
	if flag.ByName["-r"] {
		if err := reportVersions(); err != nil {
			return err
		}
		return nil
	}
	if err := doUpgrade(parm.ByName["-s"], parm.ByName["-v"],
		flag.ByName["-t"], flag.ByName["-g"], flag.ByName["-k"],
		flag.ByName["-c"], flag.ByName["-f"]); err != nil {
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
var Install_flag bool = false
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

func reportVersions() error {
	err := printImageInfo()
	if err != nil {
		return err
	}
	return nil
}

func doUpgrade(s string, v string, t bool, g bool, k bool,
	c bool, f bool) error {
	fmt.Print("\n")
	if g {
		if err := upgradeGoes(s, v, t, f); err != nil {
			return err
		}
	}
	if k {
		if err := upgradeKernel(s, v, t, f); err != nil {
			return err
		}
	}
	if c {
		if err := upgradeCoreboot(s, v, t, f); err != nil {
			return err
		}
	}
	if Reboot_flag {
		if err := reboot(); err != nil {
			return err
		}
		return nil
	}
	if Install_flag {
		if err := activateGoes(); err != nil {
			return err
		}
		return nil
	}
	return nil
}
