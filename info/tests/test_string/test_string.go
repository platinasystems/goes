// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test_string

import "github.com/platinasystems/go/info"

const Name = "test.string"

type Info struct{}

func New() Info { return Info{} }

func (Info) String() string { return Name }

func (Info) Main(...string) error {
	info.Publish(Name,
		"The quick brown fox jumped over the lazy dog's back.")
	return nil
}

func (Info) Close() error { return nil }

func (Info) Del(key string) error { return info.CantDel(key) }

func (Info) Prefixes(...string) []string { return []string{Name} }

func (Info) Set(key, value string) (err error) {
	if key != Name {
		err = info.CantSet(key)
	} else {
		info.Publish(key, value)
	}
	return
}
