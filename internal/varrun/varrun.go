// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package varrun creates and removes [/var]/run/goes/... files.
package varrun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/platinasystems/go/internal/group"
)

const Dir = "/run/goes"

var cached = struct {
	ready bool
	adm   int
	euid  int
	perms os.FileMode
}{
	adm:   -1,
	euid:  -1,
	perms: os.FileMode(0644),
}

var perms = os.FileMode(0644)
var ErrNotRoot = errors.New("you aren't root")

func cache() {
	if !cached.ready {
		cached.ready = true
		cached.euid = os.Geteuid()
		if cached.adm = group.Parse()["adm"].Gid(); cached.adm > 0 {
			cached.perms = os.FileMode(0664)
		}
	}
}

// Create the named file within Dir with proper permissions.
func Create(name string) (*os.File, error) {
	const flags = syscall.O_RDWR | syscall.O_CREAT | syscall.O_TRUNC
	if !strings.HasPrefix(name, Dir) {
		return nil, fmt.Errorf("%s: not in %q", name, Dir)
	}
	err := New(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(name, flags, perms)
	if err != nil {
		return nil, err
	}
	if cached.adm > 0 {
		f.Chown(cached.euid, cached.adm)
	}
	return f, nil
}

// New creates dir within Dir if it doesn't exist.
func New(dir string) error {
	cache()
	if cached.euid != 0 {
		return ErrNotRoot
	}
	if !strings.HasPrefix(dir, Dir) {
		return fmt.Errorf("%s: not in %q", dir, Dir)
	}
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	_, err = os.Stat(Dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(Dir, os.FileMode(0755))
		if err != nil {
			return err
		}
		if cached.adm > 0 {
			os.Chown(Dir, cached.euid, cached.adm)
		}
	}
	err = os.Mkdir(dir, os.FileMode(0775))
	if cached.adm > 0 {
		err = os.Chown(dir, cached.euid, cached.adm)
	}
	return err
}

// Path returns Dir + "/" + name if name isn't already prefaced by Dir
func Path(name string) string {
	if strings.HasPrefix(name, Dir) {
		return name
	}
	return filepath.Join(Dir, name)
}
