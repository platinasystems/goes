// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// build goes machine(s)
package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/cavaliercoder/go-cpio"
)

const (
	goesExample             = "goes-example"
	goesExampleArm          = "goes-example-arm"
	goesBoot                = "goes-boot"
	goesBootArm             = "goes-boot-arm"
	goesIP                  = "goes-ip"
	goesIPTest              = "goes-ip.test"
	goesPlatinaMk1          = "goes-platina-mk1"
	goesPlatinaMk1Installer = "goes-platina-mk1-installer"
	goesPlatinaMk1Test      = "goes-platina-mk1.test"
	goesPlatinaMk1Bmc       = "goes-platina-mk1-bmc"
	goesPlatinaMk2Lc1Bmc    = "goes-platina-mk2-lc1-bmc"
	goesPlatinaMk2Mc1Bmc    = "goes-platina-mk2-mc1-bmc"

	fe1so = "fe1.so"

	fe1     = "github.com/platinasystems/fe1"
	mainFe1 = "github.com/platinasystems/go/main/fe1"

	mainGoesPrefix           = "github.com/platinasystems/go/main/goes-"
	mainGoesExample          = mainGoesPrefix + "example"
	mainGoesBoot             = mainGoesPrefix + "boot"
	mainGoesInstaller        = mainGoesPrefix + "installer"
	mainIP                   = "github.com/platinasystems/go/main/ip"
	mainGoesPlatinaMk1       = mainGoesPrefix + "platina-mk1"
	mainGoesPlatinaMk1Bmc    = mainGoesPrefix + "platina-mk1-bmc"
	mainGoesPlatinaMk2Lc1Bmc = mainGoesPrefix + "platina-mk2-lc1-bmc"
	mainGoesPlatinaMk2Mc1Bmc = mainGoesPrefix + "platina-mk2-mc1-bmc"

	systemBuildSubmodules = "../system-build/src"

	corebootExampleAmd64        = "coreboot-example-amd64"
	corebootExampleAmd64Config  = "example-amd64_defconfig"
	corebootExampleAmd64Machine = "example-amd64"
	corebootExampleAmd64Rom     = "coreboot-example-amd64.rom"
	corebootPlatinaMk1          = "coreboot-platina-mk1"
	corebootPlatinaMk1Config    = "platina-mk1_defconfig"
	corebootPlatinaMk1Machine   = "platina-mk1"
	corebootPlatinaMk1Rom       = "coreboot-platina-mk1.rom"

	linuxKernelExampleAmd64     = "example-amd64"
	linuxKernelPlatinaMk1       = "platina-mk1"
	linuxKernelPlatinaMk1Bmc    = "platina-mk1-bmc"
	linuxKernelPlatinaMk2Lc1Bmc = "platina-mk2-lc1-bmc"
	linuxKernelPlatinaMk2Mc1Bmc = "platina-mk2-mc1-bmc"

	linuxConfigExampleAmd64     = "platina-example-amd64_defconfig"
	linuxConfigPlatinaMk1       = "platina-mk1_defconfig"
	linuxConfigPlatinaMk1Bmc    = "platina-mk1-bmc_defconfig"
	linuxConfigPlatinaMk2Lc1Bmc = "platina-mk2-lc1-bmc_defconfig"
	linuxConfigPlatinaMk2Mc1Bmc = "platina-mk2-mc1-bmc_defconfig"

	ubootPlatinaMk1Bmc       = "u-boot-platina-mk1-bmc"
	ubootPlatinaMk1BmcConfig = "platinamx6boards_sd_defconfig"
)

type goenv struct {
	goarch           string
	goos             string
	gnuPrefix        string
	kernelMakeTarget string
	kernelPath       string
	kernelConfigPath string
	kernelArch       string
	boot             string
}

var (
	defaultTargets = []string{
		goesExample,
		linuxKernelExampleAmd64,
		corebootExampleAmd64,
		goesExampleArm,
		goesBoot,
		goesBootArm,
		goesIP,
		goesPlatinaMk1,
		linuxKernelPlatinaMk1,
		corebootPlatinaMk1,
		goesPlatinaMk1Bmc,
		ubootPlatinaMk1Bmc,
		linuxKernelPlatinaMk1Bmc,
		linuxKernelPlatinaMk2Lc1Bmc,
		linuxKernelPlatinaMk2Mc1Bmc,
		corebootExampleAmd64Rom,
		corebootPlatinaMk1Rom,
	}
	goarchFlag = flag.String("goarch", runtime.GOARCH,
		"GOARCH of PACKAGE build")
	goosFlag = flag.String("goos", runtime.GOOS,
		"GOOS of PACKAGE build")
	nFlag = flag.Bool("n", false,
		"print 'go build' commands but do not run them.")
	oFlag    = flag.String("o", "", "output file name of PACKAGE build")
	tagsFlag = flag.String("tags", "", `
debug	disable optimizer and increase vnet log
diag	include manufacturing diagnostics with BMC
plugin	use pre-compiled proprietary packages
`)
	xFlag = flag.Bool("x", false, "print 'go build' commands.")
	vFlag = flag.Bool("v", false,
		"print the names of packages as they are compiled.")
	zFlag = flag.Bool("z", false, "print 'goes-build' commands.")
	host  = goenv{
		goarch: runtime.GOARCH,
		goos:   runtime.GOOS,
	}
	amd64Linux = goenv{
		goarch:           "amd64",
		goos:             "linux",
		gnuPrefix:        "x86_64-linux-gnu-",
		kernelMakeTarget: "bzImage",
		kernelPath:       "arch/x86/boot/bzImage",
		kernelConfigPath: "arch/x86/configs",
		kernelArch:       "x86_64",
		boot:             "coreboot",
	}
	armLinux = goenv{
		goarch:           "arm",
		goos:             "linux",
		gnuPrefix:        "arm-linux-gnueabi-",
		kernelPath:       "arch/arm/boot/zImage",
		kernelConfigPath: "arch/arm/configs",
		kernelArch:       "arm",
		boot:             "u-boot",
	}
	mainPkg = map[string]string{
		goesExample:                 mainGoesExample,
		linuxKernelExampleAmd64:     linuxConfigExampleAmd64,
		corebootExampleAmd64:        corebootExampleAmd64Config,
		corebootExampleAmd64Rom:     corebootExampleAmd64Machine,
		goesExampleArm:              mainGoesExample,
		goesBoot:                    mainGoesBoot,
		goesBootArm:                 mainGoesBoot,
		goesIP:                      mainIP,
		goesIPTest:                  mainIP,
		goesPlatinaMk1:              mainGoesPlatinaMk1,
		linuxKernelPlatinaMk1:       linuxConfigPlatinaMk1,
		corebootPlatinaMk1:          corebootPlatinaMk1Config,
		corebootPlatinaMk1Rom:       corebootPlatinaMk1Machine,
		goesPlatinaMk1Test:          mainGoesPlatinaMk1,
		goesPlatinaMk1Installer:     mainGoesPlatinaMk1,
		goesPlatinaMk1Bmc:           mainGoesPlatinaMk1Bmc,
		linuxKernelPlatinaMk1Bmc:    linuxConfigPlatinaMk1Bmc,
		ubootPlatinaMk1Bmc:          ubootPlatinaMk1BmcConfig,
		goesPlatinaMk2Lc1Bmc:        mainGoesPlatinaMk2Lc1Bmc,
		linuxKernelPlatinaMk2Lc1Bmc: linuxConfigPlatinaMk2Lc1Bmc,
		goesPlatinaMk2Mc1Bmc:        mainGoesPlatinaMk2Mc1Bmc,
		linuxKernelPlatinaMk2Mc1Bmc: linuxConfigPlatinaMk2Mc1Bmc,
	}
	make = map[string]func(out, name string) error{
		goesExample:                 makeHost,
		linuxKernelExampleAmd64:     makeAmd64LinuxKernel,
		corebootExampleAmd64:        makeAmd64Boot,
		corebootExampleAmd64Rom:     makeAmd64CorebootRom,
		goesExampleArm:              makeArmLinuxStatic,
		goesBoot:                    makeAmd64LinuxInitramfs,
		goesBootArm:                 makeArmLinuxInitramfs,
		goesIP:                      makeHost,
		goesIPTest:                  makeHostTest,
		goesPlatinaMk1:              makeGoesPlatinaMk1,
		linuxKernelPlatinaMk1:       makeAmd64LinuxKernel,
		corebootPlatinaMk1:          makeAmd64Boot,
		corebootPlatinaMk1Rom:       makeAmd64CorebootRom,
		goesPlatinaMk1Installer:     makeGoesPlatinaMk1Installer,
		goesPlatinaMk1Test:          makeAmd64LinuxTest,
		goesPlatinaMk1Bmc:           makeArmLinuxStatic,
		linuxKernelPlatinaMk1Bmc:    makeArmLinuxKernel,
		ubootPlatinaMk1Bmc:          makeArmBoot,
		goesPlatinaMk2Lc1Bmc:        makeArmLinuxStatic,
		linuxKernelPlatinaMk2Lc1Bmc: makeArmLinuxKernel,
		goesPlatinaMk2Mc1Bmc:        makeArmLinuxStatic,
		linuxKernelPlatinaMk2Mc1Bmc: makeArmLinuxKernel,
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()
	targets := flag.Args()
	if len(targets) == 0 {
		targets = defaultTargets
	} else if targets[0] == "all" {
		targets = targets[:0]
		for target := range make {
			targets = append(targets, target)
		}
	}
	err := host.godo("generate", "github.com/platinasystems/go")
	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()
	if err != nil {
		return
	}
	for _, target := range targets {
		if f, found := make[target]; found {
			err = f(target, mainPkg[target])
		} else {
			err = makePackage(target)
		}
		if err != nil {
			return
		}
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:", os.Args[0],
		"[ OPTION... ] [ TARGET... | PACKAGE ]")
	fmt.Fprintln(os.Stderr, "\nOptions:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nDefault Targets:")
	for _, target := range defaultTargets {
		fmt.Fprint(os.Stderr, "\t", target, "\n")
	}
	fmt.Fprintln(os.Stderr, "\n\"all\" Targets:")
	for target := range make {
		fmt.Fprint(os.Stderr, "\t", target, "\n")
	}
}

func makeArmLinuxStatic(out, name string) error {
	return armLinux.godo("build", "-o", out, "-tags", "netgo",
		"-ldflags", "-d", name)
}

func makeArmBoot(out, name string) (err error) {
	return armLinux.makeboot(out, "make "+name)
}

func makeArmLinuxKernel(out, name string) (err error) {
	return armLinux.makeLinux(out, name)
}

func makeArmLinuxInitramfs(out, name string) (err error) {
	err = makeArmLinuxStatic(out, name)
	if err != nil {
		return
	}
	return armLinux.makeCpioArchive(out)
}

func makeAmd64Boot(out, name string) (err error) {
	return amd64Linux.makeboot(out, "make crossgcc-i386 && make "+name)
}

func makeAmd64Linux(out, name string) error {
	return amd64Linux.godo("build", "-o", out, name)
}

func makeAmd64LinuxStatic(out, name string) error {
	return amd64Linux.godo("build", "-o", out, "-tags", "netgo", name)
}

func makeAmd64LinuxTest(out, name string) error {
	return amd64Linux.godo("test", "-c", "-o", out, name)
}

func makeAmd64CorebootRom(romfile, machine string) (err error) {
	dir := "worktrees/coreboot/" + machine
	build := dir + "/build"
	cbfstool := build + "/cbfstool"
	tmprom := romfile + ".tmp"

	cmdline := "cp " + build + "/coreboot.rom " + tmprom +
		" && " + cbfstool + " " + tmprom + " add-payload" +
		" -f " + machine + ".vmlinuz" +
		" -I goes-boot.cpio.xz" +
		" -C console=ttyS0" +
		" -n fallback/payload -t payload -c none -r COREBOOT" +
		" && mv " + tmprom + " " + romfile +
		" && " + cbfstool + " " + romfile + " print"
	if err := shellCommand(cmdline); err != nil {
		return err
	}
	return
}

func makeAmd64LinuxKernel(out, name string) (err error) {
	return amd64Linux.makeLinux(out, name)
}

func makeAmd64LinuxInitramfs(out, name string) (err error) {
	err = makeAmd64LinuxStatic(out, name)
	if err != nil {
		return
	}
	return amd64Linux.makeCpioArchive(out)
}

func makeHost(out, name string) error {
	return host.godo("build", "-o", out, name)
}

func makeHostTest(out, name string) error {
	return host.godo("test", "-c", "-o", out, name)
}

func makePackage(name string) error {
	args := []string{"build"}
	if len(*oFlag) > 0 {
		args = append(args, "-o", *oFlag)
	}
	return (&goenv{goarch: *goarchFlag, goos: *goosFlag}).godo(append(args, name)...)
}

func makeGoesPlatinaMk1(out, name string) error {
	args := []string{"build", "-o", out}
	if have(fe1) {
		if err := host.godo("generate", fe1); err != nil {
			return err
		}
	} else if strings.Index(*tagsFlag, "plugin") < 0 {
		args = append(args, "-tags", "plugin")
	}
	if strings.Index(*tagsFlag, "debug") > 0 {
		args = append(args, "-gcflags", "-N -l")
	}
	return amd64Linux.godo(append(args, name)...)
}

func makeGoesPlatinaMk1Installer(out, name string) error {
	var zfiles []string
	tinstaller := out + ".tmp"
	tzip := goesPlatinaMk1 + ".zip"
	err := makeGoesPlatinaMk1(goesPlatinaMk1, name)
	if err != nil {
		return err
	}
	if have(fe1) && strings.Index(*tagsFlag, "plugin") >= 0 {
		err = host.godo("generate", fe1)
		if err != nil {
			return err
		}
		err = amd64Linux.godo("build", "-buildmode=plugin", mainFe1)
		if err != nil {
			return err
		}
	}
	err = amd64Linux.godo("build", "-o", tinstaller, mainGoesInstaller)
	if err != nil {
		return err
	}
	if !have(fe1) || strings.Index(*tagsFlag, "plugin") >= 0 {
		fi, fierr := os.Stat(fe1so)
		if fierr != nil {
			fi, fierr = os.Stat("/usr/lib/goes/" + fe1so)
			if fierr != nil {
				return fmt.Errorf("can't find " + fe1so)
			}
		}
		zfiles = append(zfiles, fi.Name())
	}
	err = zipfile(tzip, append(zfiles, goesPlatinaMk1))
	if err != nil {
		return err
	}
	err = catto(out, tinstaller, tzip)
	if err != nil {
		return err
	}
	if err = rm(tinstaller, tzip); err != nil {
		return err
	}
	if err = zipa(out); err != nil {
		return err
	}
	return chmodx(out)
}

func (goenv *goenv) makeCpioArchive(name string) (err error) {
	if *nFlag {
		return nil
	}
	f, err := os.Create(name + ".cpio.xz.tmp")
	if err != nil {
		return
	}
	defer func() {
		f.Close()
		if err == nil {
			mv(name+".cpio.xz.tmp", name+".cpio.xz")
		} else {
			rm(name + ".cpio.xz.tmp")
		}
	}()
	rp, wp := io.Pipe()

	w := cpio.NewWriter(wp)

	cmd, err := filterCommand(rp, f, "xz", "--stdout", "--check=crc32", "-9")
	defer func() {
		errcmd := cmd.Wait()
		if err == nil {
			err = errcmd
		}
	}()
	defer func() {
		errclose := wp.Close()
		if err == nil {
			err = errclose
		}
	}()
	defer func() {
		errclose := w.Close()
		if err == nil {
			err = errclose
		}
	}()
	if err != nil {
		return err
	}
	for _, dir := range []struct {
		name string
		mode os.FileMode
	}{
		{".", 0775},
		{"etc", 0775},
		{"etc/ssl", 0775},
		{"etc/ssl/certs", 0775},
		{"sbin", 0775},
		{"usr", 0775},
		{"usr/bin", 0775},
	} {
		err = mkdirCpio(w, dir.name, dir.mode)
		if err != nil {
			return
		}
	}
	for _, file := range []struct {
		tname string
		mode  os.FileMode
		hname string
	}{
		{"etc/ssl/certs/ca-certificates.crt", 0644, "/etc/ssl/certs/ca-certificates.crt"},
	} {
		if err = mkfileFromHostCpio(w, file.tname, file.mode, file.hname); err != nil {
			return
		}
	}

	goesbin, err := goenv.stripBinary(name)
	if err != nil {
		return
	}
	if err = mkfileFromSliceCpio(w, "/usr/bin/goes", 0755, "(stripped)"+name, goesbin); err != nil {
		return
	}
	for _, link := range []struct {
		hname string
		tname string
	}{
		{"init", "/usr/bin/goes"},
	} {
		if err = mklinkCpio(w, link.hname, link.tname); err != nil {
			return
		}
	}
	return
}

func mkdirCpio(w *cpio.Writer, name string, perm os.FileMode) (err error) {
	host.log("{archive}mkdir", "-m", fmt.Sprintf("%o", perm), name)
	hdr := &cpio.Header{
		Name: name,
		Mode: cpio.ModeDir | cpio.FileMode(perm),
	}
	err = w.WriteHeader(hdr)
	return
}

func mklinkCpio(w *cpio.Writer, name string, target string) (err error) {
	host.log("{archive}ln", "-s", name, target)
	link := []byte(target)
	hdr := &cpio.Header{
		Name: name,
		Mode: 0120777,
		Size: int64(len(link)),
	}
	if err = w.WriteHeader(hdr); err != nil {
		return
	}
	_, err = w.Write(link)
	return
}

func mkfileFromSliceCpio(w *cpio.Writer, tname string, mode os.FileMode, hname string, data []byte) (err error) {
	hdr := &cpio.Header{
		Name: tname,
		Mode: 0100000 | cpio.FileMode(mode),
		Size: int64(len(data)),
	}
	if err = w.WriteHeader(hdr); err != nil {
		return
	}
	if _, err = w.Write(data); err != nil {
		return
	}
	host.log("{archive}cp", hname, tname)
	return
}

func mkfileFromHostCpio(w *cpio.Writer, tname string, mode os.FileMode, hname string) (err error) {
	data, err := ioutil.ReadFile(hname)
	if err != nil {
		return
	}
	return mkfileFromSliceCpio(w, tname, mode, hname, data)
}

func (goenv *goenv) godo(args ...string) error {
	if len(*tagsFlag) > 0 {
		done := false
		for i, arg := range args {
			if arg == "-tags" {
				args[i+1] = fmt.Sprint(args[i+1], " ",
					*tagsFlag)
				done = true
			}
		}
		if !done {
			args = append([]string{args[0], "-tags", *tagsFlag},
				args[1:]...)
		}
	}
	if *nFlag {
		args = append([]string{args[0], "-n"}, args[1:]...)
	}
	if *vFlag {
		args = append([]string{args[0], "-v"}, args[1:]...)
	}
	if *xFlag {
		args = append([]string{args[0], "-x"}, args[1:]...)
	}
	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	if goenv.goarch != runtime.GOARCH {
		cmd.Env = append(cmd.Env, fmt.Sprint("GOARCH=", goenv.goarch))
	}
	if goenv.goos != runtime.GOOS {
		cmd.Env = append(cmd.Env, fmt.Sprint("GOOS", goenv.goos))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	goenv.log(cmd.Args...)
	return cmd.Run()
}

func (goenv *goenv) log(args ...string) {
	if !*zFlag {
		return
	}
	fmt.Print("#")
	if goenv.goarch != runtime.GOARCH || goenv.goos != runtime.GOOS {
		fmt.Print(" {", goenv.goarch, ",", goenv.goos, "}")
	}
	for _, arg := range args {
		format := " %s"
		if strings.ContainsAny(arg, " \t") {
			format = " %q"
		}
		fmt.Printf(format, arg)
	}
	fmt.Println()
}

func catto(target string, fns ...string) error {
	host.log(append(append([]string{"cat"}, fns...), ">>", target)...)
	w, err := os.Create(target)
	if err != nil {
		return err
	}
	defer w.Close()
	for _, fn := range fns {
		r, err := os.Open(fn)
		if err != nil {
			w.Close()
			return err
		}
		io.Copy(w, r)
		r.Close()
	}
	return nil
}

func chmodx(fn string) error {
	host.log("chmod", "+x", fn)
	fi, err := os.Stat(fn)
	if err != nil {
		return err
	}
	return os.Chmod(fn, fi.Mode()|
		os.FileMode(syscall.S_IXUSR|syscall.S_IXGRP|syscall.S_IXOTH))
}

func have(pkg string) bool {
	buf, err := exec.Command("go", "list", pkg).Output()
	return err == nil && bytes.Equal(bytes.TrimSpace(buf), []byte(fe1))
}

func mv(from, to string) error {
	host.log("mv", from, to)
	return os.Rename(from, to)
}

func rm(fns ...string) error {
	host.log(append([]string{"rm"}, fns...)...)
	for _, fn := range fns {
		if err := os.Remove(fn); err != nil {
			return err
		}
	}
	return nil
}

// FIXME write a go method to prefix the self extractor header.
func zipa(fn string) error {
	cmd := exec.Command("zip", "-q", "-A", fn)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	host.log(cmd.Args...)
	if *nFlag {
		return nil
	}
	return cmd.Run()
}

func zipfile(zfn string, fns []string) error {
	host.log(append([]string{"zip", zfn}, fns...)...)
	f, err := os.Create(zfn)
	if err != nil {
		return err
	}
	defer f.Close()
	z := zip.NewWriter(f)
	defer z.Close()
	for _, fn := range fns {
		w, err := z.Create(filepath.Base(fn))
		if err != nil {
			return err
		}
		r, err := os.Open(fn)
		if err != nil {
			return err
		}
		io.Copy(w, r)
		r.Close()
	}
	return nil
}

func filterCommand(in io.Reader, out io.Writer, name string, args ...string) (cmd *exec.Cmd, err error) {
	host.log(append([]string{name}, args...)...)
	if *nFlag {
		return nil, nil
	}
	cmd = exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd, cmd.Start()
}

func (goenv *goenv) stripBinary(in string) (out []byte, err error) {
	outfile := in + ".strip.tmp"
	cmdline := []string{"-o", outfile, in}
	stripper := goenv.gnuPrefix + "strip"
	host.log(append([]string{stripper}, cmdline...)...)
	if *nFlag {
		return nil, nil
	}
	defer rm(outfile)
	cmd := exec.Command(stripper, cmdline...)
	err = cmd.Run()
	if err != nil {
		return
	}
	out, err = ioutil.ReadFile(outfile)
	return
}

func shellCommand(cmdline string) (err error) {
	args := []string{"-c", cmdline}
	if *xFlag {
		args = append(args, "-x")
	}
	host.log(append([]string{"sh"}, args...)...)
	if *nFlag {
		return nil
	}
	cmd := exec.Command("sh", args...)
	cmd.Env = os.Environ()
	if *zFlag {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}

func configWorktree(gitpath string, repo string, machine string, config string) (dir string, err error) {
	dir = "worktrees/" + repo + "/" + machine
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := shellCommand("mkdir -p " + dir +
			" && cd " + dir +
			" && p=`pwd` " +
			" && cd ../../../" + gitpath + "/" + repo +
			" && ( git worktree prune ; git branch -d " + machine +
			" ; git worktree add $p || git clone . $p )" +
			" && cd $p" +
			" && " + config); err != nil {
			return "", err
		}
	}
	return
}

func (goenv *goenv) makeboot(out string, configCommand string) (err error) {
	machine := strings.TrimPrefix(out, goenv.boot+"-")
	dir, err := configWorktree(systemBuildSubmodules, goenv.boot,
		machine, configCommand)
	if err != nil {
		return
	}
	cmdline := "make -C " + dir +
		" ARCH=" + goenv.kernelArch +
		" CROSS_COMPILE=" + goenv.gnuPrefix
	if !*zFlag { // quiet "Skipping submodule and Created CBFS" messages
		cmdline += " 2>/dev/null"
	}
	if err := shellCommand(cmdline); err != nil {
		return err
	}
	return
}

func (goenv *goenv) makeLinux(machine string, config string) (err error) {
	configCommand := "cp " + goenv.kernelConfigPath + "/" + config +
		" .config" +
		" && make oldconfig ARCH=" + goenv.kernelArch

	dir, err := configWorktree(systemBuildSubmodules, "linux",
		machine, configCommand)
	if err != nil {
		return
	}
	if err := shellCommand("make -C " + dir + " " + goenv.kernelMakeTarget +
		" ARCH=" + goenv.kernelArch +
		" CROSS_COMPILE=" + goenv.gnuPrefix); err != nil {
		return err
	}
	cmdline := "cp " + dir + "/" + goenv.kernelPath + " " + machine + ".vmlinuz"
	if err := shellCommand(cmdline); err != nil {
		return err
	}
	return
}
