// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flags

import (
	"reflect"
	"testing"
)

var lsflags = []interface{}{"-l", "-t", "-r"}

func TestLsDir(t *testing.T) {
	cmd := []string{"ls", "DIR"}
	p, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"-l": false,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong:", args)
	}
}

func TestLsL(t *testing.T) {
	cmd := []string{"ls", "-l"}
	p, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"-l": true,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ls"}) {
		t.Error("wrong:", args)
	}
}

func TestLsLdir(t *testing.T) {
	cmd := []string{"ls", "-l", "DIR"}
	p, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"-l": true,
		"-t": false,
		"-r": false,
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong:", args)
	}
}

func TestLsLTR(t *testing.T) {
	cmd := []string{"ls", "-ltr"}
	p, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"-l": true,
		"-t": true,
		"-r": true,
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ls"}) {
		t.Error("wrong:", args)
	}
}

func TestLsLTRdir(t *testing.T) {
	cmd := []string{"ls", "-ltr", "DIR"}
	p, args := New(cmd, lsflags...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"-l": true,
		"-t": true,
		"-r": true,
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ls", "DIR"}) {
		t.Error("wrong:", args)
	}
}

func TestLnVerbose(t *testing.T) {
	cmd := []string{"ln", "-verbose", "TARGET", "NAME"}
	p, args := New(cmd, []string{"-v", "-verbose"})
	if !p.ByName["-v"] {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"ln", "TARGET", "NAME"}) {
		t.Error("wrong:", args)
	}
}

func TestLnVerboSe(t *testing.T) {
	cmd := []string{"ln", "-verboSe", "TARGET", "NAME"}
	p, args := New(cmd, []string{"-v", "-verbose"})
	if p.ByName["-v"] {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, cmd) {
		t.Error("wrong:", args)
	}
}
