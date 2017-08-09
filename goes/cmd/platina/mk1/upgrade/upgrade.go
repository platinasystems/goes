// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO AUTHENTICATE,  CERT OR EQUIV
//TODO UPGRADE AUTOMATICALLY IF ENABLED, contact boot server

package upgrade

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/kexec"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "upgrade"
	Apropos = "upgrade images"
	Usage   = "upgrade [-v VER] [-s SERVER[/dir]] [-r] [-l] [-t] [-a | -g -k -c] [-f]"
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
	Suffix  = ".zip"
)

const (
	//names of server files
	GoesName      = "goes-platina-mk1"
	GoesInstaller = "goes-platina-mk1-installer"
	KernelName    = "linux-image-platina-mk1"
	CorebootName  = "coreboot-platina-mk1.rom"
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
		if err := reportVersions(parm.ByName["-s"], parm.ByName["-v"],
			flag.ByName["-t"]); err != nil {
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

var reboot_flag bool = false

func showList(s string, v string, t bool) error {
	fn := "LIST"
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	if _, err := getFile(urls, fn); err != nil {
		return err
	}
	l, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	fmt.Print(string(l))
	return nil
}

func reportVersions(s string, v string, t bool) error {
	g := curGoesVer()
	k, err := curKernelVer()
	if err != nil {
		return err
	}
	c, err := curCorebootVer()
	if err != nil {
		return err
	}
	gr, err := srvGoesVer(s, v, t)
	if err != nil {
		return err
	}
	kr, _, err := srvKernelVer(s, v, t)
	if err != nil {
		return err
	}
	cr, err := srvCorebootVer(s, v, t)
	if err != nil {
		return err
	}
	fmt.Print("\n")
	fmt.Print("Currently running:\n")
	fmt.Printf("    Goes version    : %s\n", g)
	fmt.Printf("    Kernel version  : %s\n", k)
	fmt.Printf("    Coreboot version: %s\n", c)
	fmt.Print("\n")
	fmt.Print("Version on server:\n")
	fmt.Printf("    Goes version    : %s\n", gr)
	fmt.Printf("    Kernel version  : %s\n", kr)
	fmt.Printf("    Coreboot version: %s\n", cr)
	fmt.Print("\n")
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
	if reboot_flag {
		reboot()
	}
	return nil
}

func upgradeGoes(s string, v string, t bool, f bool) error {
	fmt.Printf("Update Goes\n")
	if !f {
		g := curGoesVer()
		gr, err := srvGoesVer(s, v, t)
		if err != nil {
			return err
		}
		fmt.Printf("    Goes version currently:  %s\n", g)
		fmt.Printf("    Goes version on server:  %s\n", gr)
		if g == gr {
			fmt.Print("    Versions match, skipping Goes upgrade\n\n")
			return nil
		}
		if len(g) == 0 || len(gr) == 0 {
			fmt.Print("    No tag found, aborting Goes upgrade\n\n")
			return nil
		}
	}

	fn := GoesInstaller
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	n, err := getFile(urls, fn)
	if err != nil {
		return fmt.Errorf("    Error downloading: %v", err)
	}
	if n < 1000 {
		return fmt.Errorf("    Error file too small: %v", err)
	}
	//FIXME install goes
	return nil
}

func upgradeKernel(s string, v string, t bool, f bool) error {
	fmt.Printf("Update Kernel\n")
	if !f {
		k, err := curKernelVer()
		if err != nil {
			return err
		}
		kr, _, err := srvKernelVer(s, v, t)
		if err != nil {
			return err
		}
		fmt.Printf("    Kernel version currently:  %s\n", k)
		fmt.Printf("    Kernel version on server:  %s\n", kr)
		if k == kr {
			fmt.Print("    Versions match, skipping Kernel upgrade\n\n")
			return nil
		}
	}
	//FIXME install kernel
	reboot_flag = true
	return nil
}

func upgradeCoreboot(s string, v string, t bool, f bool) error {
	fmt.Printf("Update Coreboot\n")
	if !f {
		c, err := curCorebootVer()
		if err != nil {
			return err
		}
		cr, err := srvCorebootVer(s, v, t)
		if err != nil {
			return err
		}
		fmt.Printf("    Coreboot version currently:  %s\n", c)
		fmt.Printf("    Coreboot version on server:  %s\n", cr)
		if c == cr {
			fmt.Print("    Versions match, skipping Coreboot upgrade\n\n")
		}
	}
	//FIXME install coreboot
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

func curGoesVer() (v string) {
	ar := "tag"
	maps := []map[string]string{Package}
	if Packages != nil {
		maps = append(maps, Packages()...)
	}
	for _, m := range maps {
		if ip, found := m["importpath"]; found {
			k := strings.TrimLeft(ar, "-")
			if val, found := m[k]; found {
				if strings.Contains(ip, "/go") {
					v = val
				}
			}
		}
	}
	return v
}

func curKernelVer() (string, error) {
	u, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(u)), nil
}

func curCorebootVer() (string, error) { //FIXME
	return "no_tag", nil
}

func srvGoesVer(s string, v string, t bool) (string, error) {
	fn := GoesName
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	n, err := getFile(urls, fn)
	if err != nil {
		return "", fmt.Errorf("Error downloading: %v", err)
	}
	if n < 1000 {
		return "", fmt.Errorf("Error file too small: %v", err)
	}
	x := ""
	a, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	as := string(a)
	ref := regexp.MustCompile("v([0-9])[.]([0-9])-([0-9]+)-g([0-9a-fA-F]+)")
	x = ref.FindString(as)
	if len(x) == 0 {
		ree := regexp.MustCompile("v([0-9])[.]([0-9])")
		x = ree.FindString(as)
	}
	//rmFile(fn) //FIXME cut
	return x, nil
}

func srvKernelVer(s string, v string, t bool) (string, string, error) {
	fn := KernelName
	rmFile(fn)
	urls := "http://" + s + "/" + v + "/" + fn
	if t {
		urls = "tftp://" + s + "/" + v + "/" + fn
	}
	n, err := getFile(urls, fn)
	if err != nil {
		return "", "", fmt.Errorf("Error downloading: %v", err)
	}
	if n < 10 {
		return "", "", fmt.Errorf("Error file too small: %v", err)
	}
	a, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", "", err
	}
	u := strings.Split(string(a), "\n")
	return strings.TrimSpace(u[0]), strings.TrimSpace(u[1]), nil
}

func srvCorebootVer(s string, v string, t bool) (string, error) { //FIXME
	return "no_tag", nil
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
	fmt.Print("\nREBOOTING... Please log back in.\n")
	kexec.Prepare()
	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
