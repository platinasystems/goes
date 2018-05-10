package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	goesExampleAmd64        = "goes-example-amd64"
	goesExampleArm          = "goes-example-arm"
	goesCoreboot            = "goes-coreboot"
	goesIP                  = "goes-ip"
	goesIPTest              = "goes-ip.test"
	goesPlatinaMk1          = "goes-platina-mk1"
	goesPlatinaMk1Installer = "goes-platina-mk1-installer"
	goesPlatinaMk1Test      = "goes-platina-mk1.test"
	goesPlatinaMk1Bmc       = "goes-platina-mk1-bmc"
	goesPlatinaMk2Lc1Bmc    = "goes-platina-mk2-lc1-bmc"
	goesPlatinaMk2Mc1Bmc    = "goes-platina-mk2-mc1-bmc"

	fe1so = "fe1.so"

	top     = "github.com/platinasystems/go"
	fe1     = "github.com/platinasystems/fe1"
	mainFe1 = "github.com/platinasystems/go/main/fe1"

	mainGoesPrefix           = "github.com/platinasystems/go/main/goes-"
	mainGoesExample          = mainGoesPrefix + "example"
	mainGoesCoreboot         = mainGoesPrefix + "coreboot"
	mainGoesInstaller        = mainGoesPrefix + "installer"
	mainIP                   = "github.com/platinasystems/go/main/ip"
	mainGoesPlatinaMk1       = mainGoesPrefix + "platina-mk1"
	mainGoesPlatinaMk1Bmc    = mainGoesPrefix + "platina-mk1-bmc"
	mainGoesPlatinaMk2Lc1Bmc = mainGoesPrefix + "platina-mk2-lc1-bmc"
	mainGoesPlatinaMk2Mc1Bmc = mainGoesPrefix + "platina-mk2-mc1-bmc"
)

type goenv struct {
	goarch string
	goos   string
}

var (
	defaultTargets = []string{
		goesExampleAmd64,
		goesExampleArm,
		goesCoreboot,
		goesIP,
		goesPlatinaMk1,
		goesPlatinaMk1Bmc,
	}
	nFlag = flag.Bool("n", false,
		"print 'go build' commands but do not run them.")
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
		goarch: "amd64",
		goos:   "linux",
	}
	armLinux = goenv{
		goarch: "arm",
		goos:   "linux",
	}
	make = map[string]func() error{
		goesExampleAmd64:        makeGoesExampleAmd64,
		goesExampleArm:          makeGoesExampleArm,
		goesCoreboot:            makeGoesCoreboot,
		goesIP:                  makeIP,
		goesIPTest:              makeIPTest,
		goesPlatinaMk1:          makeGoesPlatinaMk1,
		goesPlatinaMk1Test:      makeGoesPlatinaMk1Test,
		goesPlatinaMk1Installer: makeGoesPlatinaMk1Installer,
		goesPlatinaMk1Bmc:       makeGoesPlatinaMk1Bmc,
		goesPlatinaMk2Lc1Bmc:    makeGoesPlatinaMk2Lc1Bmc,
		goesPlatinaMk2Mc1Bmc:    makeGoesPlatinaMk2Mc1Bmc,
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()
	targets := flag.Args()
	if len(targets) == 0 {
		targets = defaultTargets
	}
	for _, target := range targets {
		if f, found := make[target]; !found {
			fmt.Fprintf(os.Stderr, "%q: unknown\n", target)
			os.Exit(1)
		} else if err := f(); err != nil {
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "usage:", os.Args[0],
		"[ OPTION... ] [ TARGET... ]")
	fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
	flag.PrintDefaults()
	fmt.Fprintln(flag.CommandLine.Output(), "\nDefault Targets:")
	for _, target := range defaultTargets {
		fmt.Fprint(flag.CommandLine.Output(), "\t", target, "\n")
	}
	fmt.Fprintln(flag.CommandLine.Output(), "\nAll Targets:")
	for target := range make {
		fmt.Fprint(flag.CommandLine.Output(), "\t", target, "\n")
	}
}

func makeGoesCoreboot() error {
	return amd64Linux.godo("build", mainGoesCoreboot)
}

func makeGoesExampleAmd64() error {
	return amd64Linux.godo("build", "-o", goesExampleAmd64,
		mainGoesExample)
}

func makeGoesExampleArm() error {
	return armLinux.godo("build", "-tags", "netgo", "-ldflags", "-d",
		"-o", goesExampleArm, mainGoesExample)
}

func makeIP() error {
	if err := host.godo("generate", top); err != nil {
		return err
	}
	return host.godo("build", "-o", goesIP, mainIP)
}

func makeIPTest() error {
	return host.godo("test", "-c", "-o", goesIPTest, mainIP)
}

func makeGoesPlatinaMk1() error {
	if err := host.godo("generate", top); err != nil {
		return err
	}
	args := []string{"build"}
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
	return amd64Linux.godo(append(args, mainGoesPlatinaMk1)...)
}

func makeGoesPlatinaMk1Bmc() error {
	return armLinux.godo("build", "-tags", "netgo", "-ldflags", "-d",
		mainGoesPlatinaMk1Bmc)
}

func makeGoesPlatinaMk1Installer() error {
	const (
		tinstaller = goesPlatinaMk1Installer + ".tmp"
		tzip       = goesPlatinaMk1 + ".zip"
	)
	var zfiles []string
	err := makeGoesPlatinaMk1()
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
	err = catto(goesPlatinaMk1Installer, tinstaller, tzip)
	if err != nil {
		return err
	}
	if err = rm(tinstaller, tzip); err != nil {
		return err
	}
	if err = zipa(goesPlatinaMk1Installer); err != nil {
		return err
	}
	return chmodx(goesPlatinaMk1Installer)
}

func makeGoesPlatinaMk1Test() error {
	return amd64Linux.godo("test", "-c", "-o", goesPlatinaMk1Test,
		mainGoesPlatinaMk1)
}

func makeGoesPlatinaMk2Lc1Bmc() error {
	return armLinux.godo("build", "-tags", "netgo", "-ldflags", "-d",
		mainGoesPlatinaMk2Lc1Bmc)
}

func makeGoesPlatinaMk2Mc1Bmc() error {
	return armLinux.godo("build", "-tags", "netgo", "-ldflags", "-d",
		mainGoesPlatinaMk2Mc1Bmc)
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
