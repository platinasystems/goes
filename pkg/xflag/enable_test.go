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
	fs := flag.NewFlagSet("enabled", flag.ContinueOnError)
	EnableIn(fs, "t", "enabled", func() error {
		enabled = true
		return nil
	})
	err := fs.Parse([]string{"-t"})
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Fail()
	}
}

func TestNotEnabled(t *testing.T) {
	var enabled bool
	fs := flag.NewFlagSet("notenabled", flag.ContinueOnError)
	EnableIn(fs, "t", "enabled", func() error {
		enabled = true
		return nil
	})
	err := fs.Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if enabled {
		t.Fail()
	}
}
