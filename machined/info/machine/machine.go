// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machine

import "github.com/platinasystems/go/machined/info"

const Name = "machine"

type Info string

func New(name string) Info { return Info(name) }

func (Info) String() string { return Name }

func (name Info) Main(...string) error {
	info.Publish(Name, string(name))
	return nil
}

func (Info) Close() error { return nil }

func (Info) Del(key string) error { return info.CantDel(key) }

func (Info) Prefixes(...string) []string { return []string{Name} }

func (Info) Set(key, value string) error { return info.CantSet(key) }
