// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cmdline

import (
	"fmt"

	"github.com/platinasystems/go/cmdline"
	"github.com/platinasystems/go/machined/info"
)

const Name = "cmdline"

type Info struct {
	prefixes []string
}

func New() *Info { return &Info{[]string{Name}} }

func (*Info) String() string { return Name }

func (*Info) Main(...string) error {
	keys, m, err := cmdline.New()
	if err != nil {
		return err
	}
	for _, k := range keys {
		info.Publish(fmt.Sprint(Name, ".", k), m[k])
	}
	return nil
}

func (*Info) Close() error {
	return nil
}

func (*Info) Del(key string) error {
	return info.CantDel(key)
}

func (p *Info) Prefixes(...string) []string {
	return p.prefixes
}

func (Info) Set(key, value string) error {
	return info.CantSet(key)
}
