// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package version

import (
	"github.com/platinasystems/go/machined/info"
	"github.com/platinasystems/go/version"
)

const Name = "version"

type Info struct {
	prefixes []string
}

func New() *Info { return &Info{[]string{Name}} }

func (*Info) String() string { return Name }

func (*Info) Main(...string) error {
	info.Publish(Name, version.Version)
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

func (*Info) Set(key, value string) error {
	return info.CantSet(key)
}
