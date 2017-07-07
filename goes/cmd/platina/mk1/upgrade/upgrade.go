// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO AUTH WGET, CERT OR EQUIV
//TODO UPGRADE AUTOMATICALLY IF ENB., contact boot server

package upgrade

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "upgrade"
	Apropos = "upgrade images"
	Usage   = "upgrade [-v VER] [-s SERVER[/dir]] [-l]"
	Man     = `
DESCRIPTION
	The upgrade command updates firmware images.

	The default version is "LATEST".  Optionally, a version
	number can be supplied in the form:  v0.0[.0][.0]

	The -l flag will list available versions.

	Images are downloaded from "downloads.platina.com",
	or, from a user specified URL or IPv4 address.

	For the BMC the images in QSPI are:
	   1. goes
	   2. coreboot
	   3. debian linux distro

OPTIONS
	-v [VER]          version number or hash, default is LATEST
	-s [SERVER[/dir]] IP4 or URL, default is downloads.platina.com
	-f                use FTP instead of wget/http for downloading
	-l                lists available upgrade hashes`

	DfltMod = 0755
	DfltSrv = "downloads.platinasystems.com"
	DfltVer = "LATEST"
	Machine = "platina-mk1"
	Suffix  = ".zip"
)

var imageNames = []string{
	"goes-platina-mk1-installer",
	"coreboot-platina-mk1.rom",
	"linux-image-platina-mk1-4.11.0.deb"} //TODO filename agnostic get

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
	flag, args := flags.New(args, "-l")
	parm, args := parms.New(args, "-v", "-s")

	if len(parm["-v"]) == 0 {
		parm["-v"] = DfltVer
	}
	if len(parm["-s"]) == 0 {
		parm["-s"] = DfltSrv
	}
	if flag["-l"] {
		if err := dispList(parm["-s"], parm["-v"]); err != nil {
			return err
		}
		return nil
	}
	if err := doUpgrade(parm["-s"], parm["-v"], flag["-f"]); err != nil {
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

func doUpgrade(s string, v string, f bool) error {
	err, size := getFile(s, v)
	if err != nil {
		return fmt.Errorf("Error: Could not download file")
	}
	if size < 1000 {
		return fmt.Errorf("Error: File too small")
	}
	if err := unzip(); err != nil {
		return fmt.Errorf("Error: Could not unzip file")
	}
	if err := copyFiles(); err != nil {
		return fmt.Errorf("Error: Could not copy")
	}
	if err := rmFiles(); err != nil {
		return fmt.Errorf("Error: Could not remove files")
	}
	reboot()
	return nil
}

func reboot() error {
	kexec.Prepare()
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
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

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func dispList(s string, v string) error {
	rmFile("/LIST")
	files := []string{
		"http://" + s + "/" + v + "/" + "LIST",
	}
	err := wgetFiles(files)
	if err != nil {
		return err
	}
	l, err := ioutil.ReadFile("/LIST")
	if err != nil {
		return err
	}
	fmt.Print(string(l))
	return nil
}

func getFile(s string, v string) (error, int) {
	rmFile(Machine + Suffix)
	files := []string{
		"http://" + s + "/" + v + "/" + Machine + Suffix,
	}
	err := wgetFiles(files)
	if err != nil {
		return err, 0
	}
	f, err := os.Open(Machine + Suffix)
	if err != nil {
		return err, 0
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err, 0
	}
	filesize := int(stat.Size())
	return nil, filesize
}

func wgetFiles(urls []string) error {
	reqs := make([]*grab.Request, 0)
	for _, url := range urls {
		req, err := grab.NewRequest(url)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}

	successes, err := url.FetchReqs(0, reqs)
	if successes == 0 && err != nil {
		return err
	}
	return nil
}

func copyFiles() error {
	for _, f := range imageNames {
		err := copyFile("/"+Machine+"-"+f+
			".bin", MmcDir+"/"+f+".bin")
		if err != nil {
			return err
		}
		err = copyFile("/"+Machine+"-"+f+
			".crc", MmcDir+"/"+f+".crc")
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(f string, d string) error {
	sFile, err := os.Open(f)
	if err != nil {
		return err
	}
	defer sFile.Close()

	eFile, err := os.Create(d)
	if err != nil {
		return err
	}
	defer eFile.Close()

	_, err = io.Copy(eFile, sFile)
	if err != nil {
		return err
	}

	err = eFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func rmFiles() error {
	for _, f := range imageNames {
		err := rmFile("/" + Machine + "-" + f + ".bin")
		if err != nil {
			return err
		}
		err = rmFile("/" + Machine + "-" + f + ".crc")
		if err != nil {
			return err
		}
	}
	err := rmFile("/" + Machine + Suffix)
	if err != nil {
		return err
	}
	return nil
}

func rmFile(f string) error {
	_, err := os.Stat(f)
	if err != nil {
		return err
	}

	if err = os.Remove(f); err != nil {
		return err
	}
	return nil
}
