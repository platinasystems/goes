// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package goesd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/platinasystems/go/initutils/internal"
	"github.com/platinasystems/go/sockfile"
)

type goesd struct{}

func New() goesd { return goesd{} }

func (goesd) String() string { return "/usr/sbin/goesd" }
func (goesd) Usage() string  { return "/usr/sbin/goesd" }

func (goesd) Daemon() int { return -1 }

func (goesd goesd) Main(args ...string) error {
	if len(args) > 0 && args[0] == "stop" {
		return goesd.stop()
	}
	if err := internal.Init.Start(); err != nil {
		return err
	}
	internal.Init.Reg.Srvr.Wait()
	internal.Init.Redisd.Handler("daemons").(internal.Killaller).Killall()
	internal.Init.Redisd.Wait()
	os.Remove(internal.RunGoesPidsGoesd)
	return nil
}

func (p *goesd) stop() error {
	fns, err := filepath.Glob(filepath.Join(internal.RunGoesPids, "*"))
	if err != nil {
		return err
	}
	if len(fns) == 0 {
		return nil
	}
	// Kill goesd last
	for i := 0; i < len(fns)-1; {
		if fns[i] == internal.RunGoesPidsGoesd {
			copy(fns[i:], fns[i+1:])
			fns[len(fns)-1] = internal.RunGoesPidsGoesd
			break
		}
	}
	pids := make([]int, 0, len(fns))
	for _, fn := range fns {
		var pid int
		f, err := os.Open(fn)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if _, err = fmt.Fscan(f, &pid); err != nil {
			return err
		}
		f.Close()
		os.Remove(fn)
		_, err = os.Stat(fmt.Sprint("/proc/", pid, "/stat"))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		syscall.Kill(pid, syscall.SIGTERM)
		pids = append(pids, pid)
	}
	time.Sleep(2 * time.Second)
	for _, pid := range pids {
		_, err = os.Stat(fmt.Sprint("/proc/", pid, "/stat"))
		if err == nil {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	fns, err = filepath.Glob(filepath.Join(sockfile.Dir, "*"))
	if err == nil {
		for _, fn := range fns {
			os.Remove(fn)
		}
	}
	return nil
}
