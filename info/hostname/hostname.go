// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hostname

import (
	"os"
	"syscall"

	"github.com/platinasystems/go/info"
)

const Name = "hostname"

type Info struct{}

func New() Info { return Info{} }

func (Info) String() string { return Name }

func (Info) Main(...string) error {
	value, err := os.Hostname()
	if err != nil {
		value = err.Error()
	}
	info.Publish(Name, value)
	return err
}

func (Info) Close() error { return nil }

func (Info) Del(key string) error { return info.CantDel(key) }

func (Info) Prefixes(...string) []string { return []string{Name} }

func (Info) Set(key, value string) error {
	err := syscall.Sethostname([]byte(value))
	if err == nil {
		info.Publish(Name, value)
	}
	return err
}
