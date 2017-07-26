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
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func checkForChip() error {
	args := []string{"/usr/bin/lspci"}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return err
	}

	match, err := regexp.MatchString("Broadcom Corporation Device b96[05]",
		string(cmdOut))
	if err != nil {
		return err
	}

	if !match {
		err = fmt.Errorf("TH missing")
	}
	return err
}

func checkForKmod() error {
	args := []string{"/bin/lsmod"}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return err
	}

	match, err := regexp.MatchString("uio_pci_dma", string(cmdOut))
	if err != nil {
		return err
	}

	if !match {
		err = fmt.Errorf("not loaded")
	}
	return err
}

func checkDaemons() error {
	daemons := map[string]bool{
		"goes-daemons": true,
		"goes":         true,
		"vnetd":        true,
		"redisd":       true,
		"qsfp":         true,
		"uptimed":      true,
		"i2cd":         true,
	}

	mypid := os.Getpid()

	args := []string{"/bin/ps", "-C", "goes", "-o", "pid="}
	cmdOut, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return err
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
			return err
		}
		daemon := string(cmdOut)
		daemon = strings.Replace(daemon, "\n", "", -1)

		if err = p.Signal(os.Signal(syscall.Signal(0))); err != nil {
			fmt.Printf("Daemon [%s] not responding to signal: %s",
				daemon, err)
			continue
		}

		if _, ok := daemons[daemon]; ok == true {
			delete(daemons, daemon)
		} else {
			fmt.Printf("map NOT found for [%s]\n", daemon)
		}
	}
	for k := range daemons {
		if k == "goes" {
			continue // another instance of goes
		}
		return fmt.Errorf("%s daemon not running", k)
	}
	return err
}

func checkRedis() error {
	s, err := redis.Hget("platina", "redis.ready")
	if err != nil {
		return err
	}
	if s == "true" {
		return nil
	}
	return nil
}

func checkVnetdHung() error {
	args := []string{"/usr/bin/timeout", "30", "/usr/bin/goes",
		"vnet", "show", "hardware"}
	_, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return fmt.Errorf("vnetd daemon not responding")
	}
	return nil
}

func (Command) Main(args ...string) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("must be run as root")
	}
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	fmt.Println("GOES status")
	fmt.Println("======================")

	for _, x := range []struct {
		header string
		f      func() error
	}{
		{"PCI", checkForChip},
		{"Kernel module", checkForKmod},
		{"Check daemons", checkDaemons},
		{"Check Redis", checkRedis},
		{"Check vnet", checkVnetdHung},
	} {
		fmt.Printf("  %-15s - ", x.header)
		if err := x.f(); err == nil {
			fmt.Println("OK")
		} else {
			fmt.Printf("%s\n", err)
			return err
		}
	}

	return nil
}
