// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO AUTHENTICATE,  CERT OR EQUIV
//TODO UPGRADE AUTOMATICALLY IF ENABLED, contact boot server

/* TODO
   upgrade      -- all (same as -g -k -c), with version check
   upgrade -g   -- goes only, with version check
   upgrade -k   -- kernel only, with verion check
   upgrade -c   -- coreboot only, with version check
   upgrade -f   -- force upgrade, don't do version check
   upgrade -v   -- version report of current versions
   upgrade -r   -- remote version report
*/

package upgrade

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "upgrade"
	Apropos = "1.1 upgrade images"
	Usage   = "1.1 upgrade [-v VER] [-s SERVER[/dir]] [-l] [-t] [-d]"
	Man     = `
DESCRIPTION
	The upgrade command updates firmware images.

	The upgrade version is "LATEST" by default. Or, the upgrade
	version can be specified in the form of:  v0.0[.0][.0]

	The -l flag will list available upgrade versions.

	Images are downloaded from "downloads.platina.com",
	or from a user specified URL or IPv4 address.

OPTIONS
	-v [VER]          version number or hash, default is LATEST
	-s [SERVER[/dir]] IP4 or URL, default is downloads.platina.com
	-t                use TFTP instead of HTTP
	-l                shows list of available upgrade versions`

	DfltMod = 0755
	DfltSrv = "downloads.platinasystems.com"
	DfltVer = "LATEST"
	Machine = "platina-mk1"
	Suffix  = ".zip"
)

const (
	GoesName     = "goes-platina-mk1-installer"
	CorebootName = "coreboot-platina-mk1.rom"
	LinuxName    = "linux-mk1-filename"
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
	flag, args := flags.New(args, "-t", "-l")
	parm, args := parms.New(args, "-v", "-s")

	if len(parm.ByName["-v"]) == 0 {
		parm.ByName["-v"] = DfltVer
	}
	if len(parm.ByName["-s"]) == 0 {
		parm.ByName["-s"] = DfltSrv
	}
	if flag.ByName["-l"] {
		if err := showList(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
			return err
		}
		return nil
	}
	if err := doUpgrade(parm.ByName["-s"], parm.ByName["-v"],
		flag.ByName["-t"]); err != nil {
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

func showList(s string, v string, tftp bool) error {
	f := "LIST"
	rmFile("/" + f)
	urls := "http://" + s + "/" + v + "/" + f
	if tftp {
		urls = "tftp://" + s + "/" + v + "/" + f
	}
	if _, err := getFile(urls, f); err != nil {
		return err
	}
	l, err := ioutil.ReadFile("/" + f)
	if err != nil {
		return err
	}
	fmt.Print(string(l))
	return nil
}

func doUpgrade(s string, v string, tftp bool) error {
	return nil

	f := GoesName
	rmFile("/" + f)
	urls := "http://" + s + "/" + v + "/" + f
	if tftp {
		urls = "tftp://" + s + "/" + v + "/" + f
	}
	n, err := getFile(urls, f)
	if err != nil {
		return fmt.Errorf("Error downloading: %v", err)
	}
	if n < 1000 {
		return fmt.Errorf("Error file too small: %v", err)
	}
	//RUN INSTALLER
	//reboot()
	return nil
}

func getFile(urls string, fn string) (int, error) {
	r, err := url.Open(urls)
	if err != nil {
		return 0, err
	}
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DfltMod)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	syscall.Fsync(int(os.Stdout.Fd()))
	return int(n), nil
}

func unzip() error {
	archive := Machine + Suffix
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	target := "."
	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}
		r, err := file.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		t, err := os.OpenFile(
			path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer t.Close()
		if _, err := io.Copy(t, r); err != nil {
			return err
		}
	}
	return nil
}

func rmFile(f string) error {
	if _, err := os.Stat(f); err != nil {
		return err
	}
	if err := os.Remove(f); err != nil {
		return err
	}
	return nil
}

func reboot() error {
	kexec.Prepare()
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
