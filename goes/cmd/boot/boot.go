// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package boot

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/cmdline"
	"github.com/platinasystems/go/internal/fields"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/liner"
)

type Command struct {
	g      *goes.Goes
	mounts []*bootMnt
}

type bootKernel struct {
	kernel string
	initrd string
}

type bootMnt struct {
	mnt   string
	cl    cmdline.Cmdline
	err   error
	files []bootKernel
}

func (*Command) String() string { return "boot" }

func (*Command) Usage() string {
	return "boot [-t TIMEOUT] [PATH]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "boot another operating system",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The boot command finds other operating systems to load, and chooses
	an appropriate one to execute.

	Boot is a high level interface to the kexec command. While kexec
	performs the actual work, boot is a higher level interface that
	simplifies the process of selecting a kernel to execute.

OPTIONS
	-t	Specify a timeout`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Main(args ...string) (err error) {
	parm, args := parms.New(args, "-t")

	if len(args) == 0 {
		args = []string{"/boot"}
	}

	cnt := 0

	done := make(chan *bootMnt, len(args))
	if c.mounts == nil {
		c.mounts = make([]*bootMnt, 0)
	} else {
		c.mounts = c.mounts[:0]
	}
	for _, arg := range args {
		_, cl, err := cmdline.New()
		if err != nil {
			return err
		}
		fields := strings.Split(arg, ":")
		m := &bootMnt{}
		m.mnt = fields[0]
		m.cl = cl
		if len(fields) > 1 {
			m.cl.Set(fields[1])
		}
		c.mounts = append(c.mounts, m)
		go c.tryScanFiles(m, done)
		cnt++
	}

	line := liner.NewLiner()
	defer line.Close()
	if parm.ByName["-t"] != "" {
		timeout, err := time.ParseDuration(parm.ByName["-t"])
		if err != nil {
			return err
		}
		err = line.SetDuration(timeout)
		if err != nil {
			return err
		}
	}
	line.SetCtrlCAborts(true)

	defBoot := ""

	for i := 0; i < cnt; i++ {
		<-done
	}

	re := regexp.MustCompile("([0-9]+)\\.([0-9]+)\\.([0-9]+)-([0-9]+)")

	for _, m := range c.mounts {
		sort.Slice(m.files, func(i, j int) bool {
			r1 := re.FindStringSubmatch(m.files[i].kernel)
			r2 := re.FindStringSubmatch(m.files[j].kernel)
			if len(r1) == 5 && len(r2) == 5 {
				for k := 1; k < len(r1); k++ {
					v1, _ := strconv.Atoi(r1[k])
					v2, _ := strconv.Atoi(r2[k])
					if v1 < v2 {
						return true
					}
					if v1 > v2 {
						return false
					}
				}
				return false
			}
			return m.files[i].kernel < m.files[j].kernel
		})
	}

	for _, m := range c.mounts {
		for _, file := range m.files {
			cl := fmt.Sprintf(`kexec -k %s/%s -i %s/%s -e -c "%s"`,
				m.mnt, file.kernel, m.mnt, file.initrd, m.cl)
			line.AppendHistory(cl)
			defBoot = cl
		}
	}

	resp, err := line.PromptWithSuggestion("Boot command: ",
		defBoot, -1)

	if err != nil {
		if err == liner.ErrTimeOut {
			resp = defBoot
			fmt.Println("<timeout>")
		} else {
			return err
		}
	}
	kCmd := fields.New(resp)

	if len(kCmd) > 0 {
		return c.g.Main(kCmd...)
	}
	return nil
}

func (*Command) tryScanFiles(m *bootMnt, done chan *bootMnt) {
	files, err := ioutil.ReadDir(m.mnt)
	if err != nil {
		m.err = err
		done <- m
		return
	}

	for _, file := range files {
		name := file.Name()
		if file.Mode().IsRegular() {
			if strings.Contains(name, "vmlinuz") {
				if _, err := os.Stat(m.mnt + "/" + name); err == nil {
					for _, ird := range []string{
						"initrd.img",
						"initrd.gz",
						"initrd.xz",
						"initrd.lzma"} {
						i := strings.Replace(name, "vmlinuz",
							ird, 1)
						if _, err := os.Stat(m.mnt + "/" + i); err == nil {
							b := bootKernel{kernel: name, initrd: i}
							m.files = append(m.files, b)
						}
					}
				}
			}
		}
	}
	done <- m
}
