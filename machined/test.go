// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machined

type test struct{}

var Test InfoProvider = &test{}

func (*test) String() string { return "test" }
func (*test) Close() error   { return nil }

func (*test) Del(key string) error {
	return CantDel(key)
}

func (*test) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		return prefixes
	}
	return []string{}
}

func (*test) Set(key, value string) error {
	return CantSet(key)
}
