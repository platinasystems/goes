// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"flag"
	"testing"
)

func TestEnabled(t *testing.T) {
	var enabled bool
	eflag := NewEnable("e", "enabled", func() error {
		enabled = true
		return nil
	})
	fs := flag.NewFlagSet("enabled", flag.ContinueOnError)
	eflag.DefineIn(fs)
	err := fs.Parse([]string{"-e"})
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Fail()
	}
	if !eflag.Value() {
		t.Fail()
	}
}

func TestNotEnabled(t *testing.T) {
	var enabled bool
	eflag := NewEnable("e", "not-enabled", func() error {
		enabled = true
		return nil
	})
	fs := flag.NewFlagSet("notenabled", flag.ContinueOnError)
	eflag.DefineIn(fs)
	err := fs.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if enabled {
		t.Fail()
	}
	if eflag.Value() {
		t.Fail()
	}
}
