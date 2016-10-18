// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package pidfile records pids in /run/goes/pids
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/platinasystems/go/rundir"
)

const Dir = "/run/goes/pids"

func New() (string, error) {
	pid := os.Getpid()
	fn := filepath.Join(Dir, fmt.Sprint(pid))
	f, err := rundir.Create(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fmt.Fprintln(f, pid)
	return fn, err
}

// Path returns Dir + "/" + name if name isn't already prefaced by Dir
func Path(name string) string {
	if strings.HasPrefix(name, Dir) {
		return name
	}
	return filepath.Join(Dir, name)
}

func RemoveAll() {
	pids, err := filepath.Glob(filepath.Join(Dir, "*"))
	if err == nil {
		for _, fn := range pids {
			os.Remove(fn)
		}
		os.Remove(Dir)
	}
}
