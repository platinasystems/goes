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

	"github.com/platinasystems/go/group"
)

const Dir = "/run/goes"

var adm = -1
var euid = -1
var perms = os.FileMode(0644)
var ErrNotRoot = errors.New("you aren't root")

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
	if adm > 0 {
		f.Chown(euid, adm)
	}
	return f, nil
}

// New creates dir within Dir if it doesn't exist.
func New(dir string) error {
	if !strings.HasPrefix(dir, Dir) {
		return fmt.Errorf("%s: not in %q", dir, Dir)
	}
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if euid < 0 {
		euid = os.Geteuid()
		if euid != 0 {
			return ErrNotRoot
		}
	}
	if adm < 0 {
		adm = group.Parse()["adm"].Gid()
		if adm > 0 {
			perms = os.FileMode(0664)
		}
	}
	_, err = os.Stat(Dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(Dir, os.FileMode(0755))
		if err != nil {
			return err
		}
		if adm > 0 {
			os.Chown(Dir, euid, adm)
		}
	}
	err = os.Mkdir(dir, os.FileMode(0775))
	if adm > 0 {
		err = os.Chown(dir, euid, adm)
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
