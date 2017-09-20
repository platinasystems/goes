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
	Usage   = "upgrade [-v VER] [-s SERVER[/dir]] [-r] [-l] [-c] [-t] [-f]"
	Man     = `
DESCRIPTION
	The upgrade command updates firmware images.

	The default upgrade version is "LATEST". 
	Or specify a version using "-v", in the form YYYYMMDD

	The -l flag display version of selected server and version.

	The -r flag reports QSPI version numbers and booted from.

	By default, images are downloaded from "downloads.platina.com".
	Or from a server using "-s" followed by a URL or IPv4 address.

	Upgrade proceeds only if the selected version number is newer,
	unless overridden with the "-f" force flag.

OPTIONS
	-v [VER]          version [YYYYMMDD] or LATEST (default)
	-s [SERVER[/dir]] IP4 or URL, default is downloads.platina.com
	-t                use TFTP instead of HTTP
	-l                display version of selected server and version
	-r                report QSPI installed versions, QSPI booted from
	-c                check SHA-1's of flash
	-f                force upgrade (ignore version check)`

	DfltMod     = 0755
	DfltSrv     = "downloads.platinasystems.com"
	DfltVer     = "LATEST"
	Machine     = "platina-mk1-bmc"
	ArchiveName = "platina-mk1-bmc.zip"
	VersionName = "platina-mk1-bmc-ver.bin"
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
	initQfmt()
	flag, args := flags.New(args, "-t", "-l", "-f", "-r", "-c", "-1")
	parm, args := parms.New(args, "-v", "-s")
	if len(parm.ByName["-v"]) == 0 {
		parm.ByName["-v"] = DfltVer
	}
	if len(parm.ByName["-s"]) == 0 {
		parm.ByName["-s"] = DfltSrv
	}

	if flag.ByName["-l"] {
		if err := reportVerServer(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
			return err
		}
		return nil
	}
	if flag.ByName["-r"] {
		if err := reportVerQSPIdetail(); err != nil {
			return err
		}
		if err := reportVerQSPI(); err != nil {
			return err
		}
		return nil
	}
	if flag.ByName["-c"] {
		if err := compareChecksums(); err != nil {
			return err
		}
		return nil
	}

	if err := doUpgrade(parm.ByName["-s"], parm.ByName["-v"],
		flag.ByName["-t"], flag.ByName["-f"],
		flag.ByName["-1"]); err != nil {
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

func reportVerServer(s string, v string, t bool) (err error) {
	n, err := getFile(s, v, t, ArchiveName)
	if err != nil {
		return fmt.Errorf("Server unreachable\n")
	}
	if n < 1000 {
		return fmt.Errorf("Server unreachable\n")
	}
	if err := unzip(); err != nil {
		return fmt.Errorf("Server error: unzipping file: %\n", err)
	}
	defer rmFiles()

	l, err := ioutil.ReadFile(VersionName)
	if err != nil {
		fmt.Printf("Image version not found on server\n")
		return nil
	}
	sv := string(l[VERSION_OFFSET:VERSION_LEN])
	if string(l[VERSION_OFFSET:VERSION_DEV]) == "dev" {
		sv = "dev"
	}
	printVerServer(s, v, sv)
	return nil
}

func reportVerQSPIdetail() (err error) {
	err = printJSON(false)
	if err != nil {
		return err
	}
	err = printJSON(true)
	if err != nil {
		return err
	}
	return nil
}

func reportVerQSPI() (err error) {
	iv, err := getInstalledVersions()
	if err != nil {
		return err
	}
	q, err := getBootedQSPI()
	if err != nil {
		return err
	}
	printVerQSPI(iv, q)
	return nil
}

func compareChecksums() (err error) {
	if err = cmpSums(false); err != nil {
		return err
	}
	if err = cmpSums(true); err != nil {
		return err
	}
	return nil
}

func doUpgrade(s string, v string, t bool, f bool, q bool) error {
	fmt.Print("\n")

	n, err := getFile(s, v, t, ArchiveName)
	if err != nil {
		return fmt.Errorf("Server unreachable\n")
	}
	if n < 1000 {
		return fmt.Errorf("Server unreachable\n")
	}
	if err := unzip(); err != nil {
		return fmt.Errorf("Server error: unzipping file: %v\n", err)
	}
	defer rmFiles()

	if !f {
		qv, err := getVerQSPI(q)
		if err != nil {
			return err
		}
		if len(qv) == 0 {
			fmt.Printf("Aborting, couldn't find version in QSPI\n")
			fmt.Printf("Use -f to force upgrade.\n")
			return nil
		}

		l, err := ioutil.ReadFile(VersionName)
		if err != nil {
			fmt.Printf("Aborting, couldn't find version number on server\n")
			fmt.Printf("Use -f to force upgrade.\n")
			return nil
		}
		sv := string(l[VERSION_OFFSET:VERSION_LEN])
		if string(l[VERSION_OFFSET:VERSION_DEV]) == "dev" {
			sv = "dev"
		}
		if sv != "dev" && qv != "dev" {
			newer, err := isVersionNewer(qv, sv)
			if err != nil {
				fmt.Printf("Aborting, server version error\n", sv)
				fmt.Printf("Use -f to force upgrade.\n")
				return nil
			}
			if !newer {
				fmt.Printf("Aborting, server version %s is not newer\n", sv)
				fmt.Printf("Use -f to force upgrade.\n")
				return nil
			}
		}
	}

	selectQSPI(q)
	if q == true {
		fmt.Println("Upgrading QSPI1...\n")
	}
	if err := writeImageAll(); err != nil {
		return fmt.Errorf("*** UPGRADE ERROR! ***: %v\n", err)
	}
	if err := reboot(); err != nil {
		return err
	}
	return nil
}
