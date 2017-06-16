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

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "status"
	Apropos = "print status of goes daemons"
	Usage   = "status"
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	vnetd_down bool = false
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Kind() cmd.Kind    { return cmd.DontFork }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func checkForChip() bool {
	args := []string{"/usr/bin/lspci"}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", cmdOut)
		os.Exit(1)
	}

	match, err := regexp.MatchString("Broadcom Corporation Device b96[05]",
		string(cmdOut))

	if err == nil && match == true {
		return true
	} else {
		return false
	}
}

func checkForKmod() bool {
	args := []string{"/bin/lsmod"}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
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

	args := []string{"/bin/ps", "-C", "goes", "-o", "pid="}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
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

		args = []string{"/bin/ps", "-p", pid, "-o", "cmd="}
		cmdOut, err = exec.Command(args[0], args[1:]...).Output()
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
	args := []string{"/usr/bin/goes", "vnet", "show", "hardware"}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "out =", cmdOut)
		return false
	}

	return true
}

func (Command) Main(args ...string) error {
	if os.Getuid() != 0 {
		fmt.Println("must be run as root")
		os.Exit(1)
	}
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
		os.Exit(1)
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
