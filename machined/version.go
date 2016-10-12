// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machined

import (
	repos "github.com/platinasystems/go/version"
)

type version struct {
	prefixes []string
}

var Version InfoProvider = &version{
	prefixes: []string{"version"},
}

func (*version) String() string { return "version" }

func (*version) Main(...string) error {
	Publish("version", repos.Version)
	return nil
}

func (*version) Close() error {
	return nil
}

func (*version) Del(key string) error {
	return CantDel(key)
}

func (p *version) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (*version) Set(key, value string) error {
	return CantSet(key)
}
