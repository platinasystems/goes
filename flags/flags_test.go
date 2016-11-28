// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flags

import (
	"reflect"
	"testing"
)

var lsflags = []string{"-l", "-t", "-r"}

func TestLsDir(t *testing.T) {
	cmd := []string{"ls", "DIR"}
	flag, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(flag, Flag{
		"-l": false,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong args:", args)
	}
}

func TestLsL(t *testing.T) {
	cmd := []string{"ls", "-l"}
	flag, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(flag, Flag{
		"-l": true,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ls"}) {
		t.Error("wrong args:", args)
	}
}

func TestLsLdir(t *testing.T) {
	cmd := []string{"ls", "-l", "DIR"}
	flag, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(flag, Flag{
		"-l": true,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong args:", args)
	}
}

func TestLsLTR(t *testing.T) {
	cmd := []string{"ls", "-ltr"}
	flag, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(flag, Flag{
		"-l": true,
		"-t": true,
		"-r": true,
	}) {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ls"}) {
		t.Error("wrong args:", args)
	}
}

func TestLsLTRdir(t *testing.T) {
	cmd := []string{"ls", "-ltr", "DIR"}
	flag, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(flag, Flag{
		"-l": true,
		"-t": true,
		"-r": true,
	}) {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong args:", args)
	}
}

func TestLnVerbose(t *testing.T) {
	cmd := []string{"ln", "-verbose", "TARGET", "NAME"}
	flag, args := New(cmd, "-v", "-verbose")
	flag.Aka("-v", "-verbose")
	if !flag["-v"] {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, []string{"ln", "TARGET", "NAME"}) {
		t.Error("wrong args:", args)
	}
}

func TestLnVerboSe(t *testing.T) {
	cmd := []string{"ln", "-verboSe", "TARGET", "NAME"}
	flag, args := New(cmd, "-v", "-verbose")
	flag.Aka("-v", "-verbose")
	if flag["-v"] {
		t.Error("wrong flag:", flag)
	}
	if !reflect.DeepEqual(args, cmd) {
		t.Error("wrong args:", args)
	}
}
