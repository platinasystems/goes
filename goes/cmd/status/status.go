// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package status

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "status"
	Apropos = "print status of goes daemons"
	Usage   = "status"
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

var vnetd_down bool = false

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Kind() goes.Kind { return goes.DontFork }

func checkForChip() bool {
	var (
		cmdOut []byte
		err    error
	)
	cmd := "/usr/bin/lspci"
	args := []string{}
	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", cmdOut)
		os.Exit(1)
	}

	match, err := regexp.MatchString("Broadcom Corporation Device b960",
		string(cmdOut))
	if err == nil && match == true {
		return true
	} else {
		return false
	}
}

func checkForKmod() bool {
	var (
		cmdOut []byte
		err    error
	)
	cmd := "/bin/lsmod"
	args := []string{}
	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", cmdOut)
		os.Exit(1)
	}

	match, err := regexp.MatchString("uio_pci_dma", string(cmdOut))
	if err == nil && match == true {
		return true
	} else {
		return false
	}
}

func checkDaemons() bool {
	var (
		cmdOut []byte
		err    error
	)

	daemons := map[string]bool{
		"goes-daemons": true,
		"vnetd":        true,
		"redisd":       true,
		"qsfp":         true,
		"uptimed":      true,
		"i2cd":         true,
	}

	mypid := os.Getpid()
	status := true

	cmd := "/bin/ps"
	args := []string{"-C", "goes", "-o", "pid="}
	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", string(cmdOut))
		os.Exit(1)
	}
	pids := strings.Split(string(cmdOut), "\n")

	for _, pid := range pids {
		if pid == "" {
			continue
		}
		pid = strings.Replace(pid, " ", "", -1)
		pid_i, err := strconv.Atoi(pid)
		if err != nil {
			fmt.Printf("err converting [%s]: %v\n", pid, err)
			continue
		}

		if pid_i == mypid {
			continue
		}

		p, err := os.FindProcess(pid_i)
		if err != nil {
			fmt.Println("FindProcess error", err)
			continue
		}

		args = []string{"-p", pid, "-o", "cmd="}
		cmdOut, err = exec.Command(cmd, args...).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "out =",
				string(cmdOut))
			os.Exit(1)
		}
		daemon := string(cmdOut)
		daemon = strings.Replace(daemon, "\n", "", -1)

		if err = p.Signal(os.Signal(syscall.Signal(0))); err != nil {
			fmt.Printf("Daemon [%s] not responding to signal: %s",
				daemon, err)
			status = false
			continue
		}

		if _, ok := daemons[daemon]; ok == true {
			fmt.Printf("    %-13s - OK\n", daemon)
			delete(daemons, daemon)
		} else {
			fmt.Println("map NOT found for", daemon, ok)
			status = false
		}
	}
	for k := range daemons {
		fmt.Printf("    %-13s - not running\n", k)
		if k == "vnetd" {
			vnetd_down = true
		}
		status = false

	}
	return status
}

func checkRedis() bool {
	s, err := redis.Hget("platina", "redis.ready")
	if err != nil {
		fmt.Println("redis error:", err)
		return false
	}
	if s == "true" {
		return true
	}
	return true
}

func checkVnetdHung() bool {
	var (
		cmdOut []byte
		err    error
	)
	cmd := "/usr/bin/goes"
	args := []string{"vnet", "show", "hardware"}
	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", cmdOut)
		return false
	}

	return true
}

func (cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	fmt.Println("GOES status")
	fmt.Println("======================")

	fmt.Printf("  %-15s - ", "PCI")
	if checkForChip() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf(" TH not found on PCI bus\n")
		os.Exit(2)
	}

	fmt.Printf("  %-15s - ", "Kernel module")
	if checkForKmod() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("uio_pci_dma module not loaded\n")
		os.Exit(3)
	}

	fmt.Printf("  Check daemons:\n")
	if !checkDaemons() {
		os.Exit(4)
	}

	fmt.Printf("  %-15s - ", "Check Redis")
	if checkRedis() {
		fmt.Printf("OK\n")
	} else {
		os.Exit(4)
	}

	if vnetd_down == false {
		fmt.Printf("  %-15s - ", "Check vnet")
		if checkVnetdHung() {
			fmt.Printf("OK\n")
		} else {
			fmt.Printf("VNET not responding\n")
			os.Exit(5)
		}
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
