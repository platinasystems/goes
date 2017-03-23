// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package boot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/platinasystems/go/internal/fields"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/liner"
)

const Name = "boot"

type bootSet struct {
	kernel string
	initrd string
}

type bootMnt struct {
	mnt   string
	cmd   string
	err   error
	files []bootSet
}

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.DontFork }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name + " [OPTIONS]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) (err error) {
	byName := goes.ByName(*c)
	parm, args := parms.New(args, "-t")

	if len(args) == 0 {
		args = append(args, "/boot")
	}

	timeout := time.Duration(0)
	if parm["-t"] != "" {
		t, err := strconv.ParseUint(parm["-t"], 10, 8)
		if err != nil {
			return err
		}
		timeout = time.Duration(t) * time.Second
	}

	cnt := 0

	done := make(chan *bootMnt, len(args))

	for _, arg := range args {
		fields := strings.Split(arg, ":")
		m := &bootMnt{}
		m.mnt = fields[0]
		if len(fields) > 0 {
			m.cmd = fields[1]
		}
		go c.tryScanFiles(m, done)
		cnt++
	}

	line := liner.NewLiner()
	defer line.Close()
	err = line.SetDuration(timeout)
	if err != nil {
		return err
	}
	line.SetCtrlCAborts(true)

	defBoot := ""

	for i := 0; i < cnt; i++ {
		m := <-done
		for _, file := range m.files {
			c := fmt.Sprintf("kexec -k %s/%s -i %s/%s -e",
				m.mnt, file.kernel, m.mnt, file.initrd)
			if m.cmd != "" {
				c = fmt.Sprintf("%s -c %s", c, m.cmd)
			}
			if c > defBoot {
				defBoot = c
			}
			line.AppendHistory(c)
		}
	}

	resp, err := line.PromptWithSuggestion("Boot command: ",
		defBoot, -1)

	if err != nil {
		if err == liner.ErrTimeOut {
			resp = defBoot
		} else {
			return err
		}
	}
	kCmd := fields.New(resp)

	err = byName.Main(kCmd...)

	return err
}

func (c *cmd) tryScanFiles(m *bootMnt, done chan *bootMnt) {
	files, err := ioutil.ReadDir(m.mnt)
	if err != nil {
		m.err = err
		done <- m
		return
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "vmlinuz") {
			i := strings.Replace(file.Name(), "vmlinuz",
				"initrd.img", 1)
			if _, err := os.Stat(m.mnt + "/" + file.Name()); err != nil {
				i = ""
			}
			b := bootSet{kernel: file.Name(), initrd: i}
			m.files = append(m.files, b)
		}
	}
	done <- m
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "boot another operating system",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	boot - Boot another operating system

SYNOPSIS
	boot [-t seconds] [path]...

DESCRIPTION

The boot command finds other operating systems to load, and chooses an
appropriate one to execute.

Boot is a high level interface to the kexec command. While kexec performs
the actual work, boot is a higher level interface that simplifies the process
of selecting a kernel to execute.

OPTIONS
	-t	Specify a timeout in seconds`,
	}
}
