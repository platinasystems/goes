// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"errors"
	"flag"
	"testing"
)

func TestCatchNilItem(t *testing.T) {
	fs := flag.NewFlagSet("catch nil item", flag.ContinueOnError)
	err := Labels{
		{"nil", fs.Name(), nil},
	}.DefineIn(fs)
	if err == nil {
		t.Error("uncaught nil item")
	}
}

func TestCatchNonPointer(t *testing.T) {
	fs := flag.NewFlagSet("catch non pointer", flag.ContinueOnError)
	err := Labels{
		{"val", fs.Name(), true},
	}.DefineIn(fs)
	if err == nil {
		t.Error("uncaught non pointer")
	} else {
		t.Log("caught:", err)
	}
}

func TestCatchInitFault(t *testing.T) {
	fs := flag.NewFlagSet("catch init fault", flag.ContinueOnError)
	err := Labels{
		{"val", fs.Name(), func() any {
			return errors.New("init fault")
		}},
	}.DefineIn(fs)
	if err == nil {
		t.Error("uncaught init fault")
	} else {
		t.Log("caught:", err)
	}
}

func TestCatchInitNonPointer(t *testing.T) {
	fs := flag.NewFlagSet("catch init non pointer", flag.ContinueOnError)
	err := Labels{
		{"val", fs.Name(), func() any {
			return true
		}},
	}.DefineIn(fs)
	if err == nil {
		t.Error("uncaught init non pointer")
	} else {
		t.Log("caught:", err)
	}
}
