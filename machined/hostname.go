// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package machined

import (
	"os"
	"syscall"
)

type hostname struct {
	prefixes []string
}

var Hostname InfoProvider = &hostname{
	prefixes: []string{"hostname"},
}

func (*hostname) Main(...string) error {
	value, err := os.Hostname()
	if err != nil {
		value = err.Error()
	}
	Publish("hostname", value)
	return err
}

func (*hostname) Close() error {
	return nil
}

func (*hostname) Del(key string) error {
	return CantDel(key)
}

func (p *hostname) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (*hostname) Set(key, value string) error {
	err := syscall.Sethostname([]byte(value))
	if err == nil {
		Publish("hostname", value)
	}
	return err
}

func (*hostname) String() string { return "hostname" }
