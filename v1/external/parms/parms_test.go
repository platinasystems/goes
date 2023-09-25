// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parms

import (
	"reflect"
	"testing"
)

var ddparms = []interface{}{"bs", "count", "seek"}
var ioparms = []interface{}{"<", "<<", ">", ">>"}

func TestDdBs(t *testing.T) {
	cmd := []string{"dd", "bs=4k"}
	p, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"bs":    "4k",
		"count": "",
		"seek":  "",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"dd"}) {
		t.Error("wrong:", args)
	}
}

func TestDdBsCount(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1"}
	p, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"dd"}) {
		t.Error("wrong:", args)
	}
}

func TestDdBsCountFile(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1", "FILE"}
	p, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"dd", "FILE"}) {
		t.Error("wrong:", args)
	}
}

func TestDdBsCountFileGtOut(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1", "FILE", ">", "OUT"}
	p, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(p.ByName, ByName{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong:", p.ByName)
	}
	iop, args := New(args, ioparms...)
	if !reflect.DeepEqual(iop.ByName, ByName{
		"<":  "",
		"<<": "",
		">":  "OUT",
		">>": "",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"dd", "FILE"}) {
		t.Error("wrong:", args)
	}
}

func TestConcat(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar", "BAR"}
	p, args := New(cmd, "-bar")
	if !reflect.DeepEqual(p.ByName, ByName{
		"-bar": "FOO BAR",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"foo"}) {
		t.Error("wrong:", args)
	}
}

func TestConcatPlus(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar", "BAR", "bar"}
	p, args := New(cmd, "-bar")
	if !reflect.DeepEqual(p.ByName, ByName{
		"-bar": "FOO BAR",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"foo", "bar"}) {
		t.Error("wrong:", args)
	}
}

func TestClearAndSet(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar=", "-bar=BAR"}
	p, args := New(cmd, "-bar")
	if !reflect.DeepEqual(p.ByName, ByName{
		"-bar": "BAR",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"foo"}) {
		t.Error("wrong:", args)
	}
}

func TestClearAndSetPlus(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar=", "-bar=BAR", "bar"}
	p, args := New(cmd, "-bar")
	if !reflect.DeepEqual(p.ByName, ByName{
		"-bar": "BAR",
	}) {
		t.Error("wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"foo", "bar"}) {
		t.Error("wrong:", args)
	}
}

func TestKexecBug(t *testing.T) {
	cmd := []string{"kexec", "-k", "foo.vmlinuz", "-i", "foo.initrd"}
	p, args := New(cmd, "-c", "-i", "-k", "-l", "-x")
	if !reflect.DeepEqual(p.ByName, ByName{
		"-c": "",
		"-i": "foo.initrd",
		"-k": "foo.vmlinuz",
		"-l": "",
		"-x": "",
	}) {
		t.Error("Wrong:", p.ByName)
	}
	if !reflect.DeepEqual(args, []string{"kexec"}) {
		t.Error("wrong:", args)
	}
}
