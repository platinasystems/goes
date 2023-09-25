// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grubd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g      *goes.Goes
	mounts []*bootMnt
}

type bootMnt struct {
	mnt     string
	dir     string
	err     error
	hasGrub bool
}

func (*Command) String() string { return "grubd" }

func (*Command) Usage() string {
	return "grubd [PATH]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "boot another operating system",
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(args ...string) (err error) {
	mp := "/mountd"
	if len(args) > 0 {
		mp = args[0]
	}

	done := make(chan *bootMnt, len(args))
	if c.mounts == nil {
		c.mounts = make([]*bootMnt, 0)
	}
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		dirs, err := ioutil.ReadDir(mp)
		if err != nil {
			fmt.Printf("Error reading %s dir: %s", mp, err)
		}
		cnt := 0
		c.mounts = c.mounts[:0]
		for _, dir := range dirs {
			for _, sd := range []string{"", "/boot", "/d-i"} {
				m := &bootMnt{
					mnt: filepath.Join(mp, dir.Name()),
					dir: sd,
				}
				c.mounts = append(c.mounts, m)
				goes.WG.Add(1)
				go func() {
					defer goes.WG.Done()
					c.tryScanFiles(m, done)
				}()
				cnt++
			}
		}
		for i := 0; i < cnt; i++ {
			<-done
		}

		sort.Slice(c.mounts, func(i, j int) bool {
			return c.mounts[i].mnt > c.mounts[j].mnt
		})

		for _, m := range c.mounts {
			if m.hasGrub {
				done := make(chan struct{}, 1)
				args := []string{"grub", "--daemon"}
				args = append(args, m.mnt,
					filepath.Join(m.dir, "grub/grub.cfg"))
				fmt.Printf("%v\n", args)
				x := c.g.Fork(args...)
				x.Stdin = os.Stdin
				x.Stdout = os.Stdout
				x.Stderr = os.Stderr
				x.Dir = "/"
				goes.WG.Add(1)
				go func() {
					defer goes.WG.Done()
					for {
						select {
						case <-done:
							return
						case <-goes.Stop:
							p := x.Process
							if p != nil {
								p.Kill()
							}
							return
						}
					}
				}()
				err := x.Run()
				if err != nil {
					fmt.Printf("grub returned %s\n", err)
				}
				close(done)
			}
		}

		select {
		case <-goes.Stop:
			return nil
		case <-t.C:
		}
	}
}

func (*Command) tryScanFiles(m *bootMnt, done chan *bootMnt) {
	files, err := ioutil.ReadDir(filepath.Join(m.mnt, m.dir))
	if err != nil {
		m.err = err
		done <- m
		return
	}

	for _, file := range files {
		name := file.Name()
		if file.Mode().IsDir() && name == "grub" {
			if _, err := os.Stat(filepath.Join(m.mnt,
				m.dir, "grub/grub.cfg")); err == nil {
				m.hasGrub = true
			} else {
				fmt.Printf("os.stat err %s\n", err)
			}
			continue
		}
	}
	done <- m
}
