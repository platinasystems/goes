// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parms

import (
	"reflect"
	"testing"
)

var ddparms = []string{"bs", "count", "seek"}
var ioparms = []string{"<", "<<", ">", ">>"}

func TestDdBs(t *testing.T) {
	cmd := []string{"dd", "bs=4k"}
	parm, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(parm, Parm{
		"bs":    "4k",
		"count": "",
		"seek":  "",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"dd"}) {
		t.Error("wrong args:", args)
	}
}

func TestDdBsCount(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1"}
	parm, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(parm, Parm{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"dd"}) {
		t.Error("wrong args:", args)
	}
}

func TestDdBsCountFile(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1", "FILE"}
	parm, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(parm, Parm{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"dd", "FILE"}) {
		t.Error("wrong args:", args)
	}
}

func TestDdBsCountFileGtOut(t *testing.T) {
	cmd := []string{"dd", "bs=4k", "count", "1", "FILE", ">", "OUT"}
	parm, args := New(cmd, ddparms...)
	if !reflect.DeepEqual(parm, Parm{
		"bs":    "4k",
		"count": "1",
		"seek":  "",
	}) {
		t.Error("wrong parm:", parm)
	}
	ioparm, args := New(args, ioparms...)
	if !reflect.DeepEqual(ioparm, Parm{
		"<":  "",
		"<<": "",
		">":  "OUT",
		">>": "",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"dd", "FILE"}) {
		t.Error("wrong args:", args)
	}
}

func TestConcat(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar", "BAR"}
	parm, args := New(cmd, "-bar")
	if !reflect.DeepEqual(parm, Parm{
		"-bar": "FOO BAR",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"foo"}) {
		t.Error("wrong args:", args)
	}
}

func TestConcatPlus(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar", "BAR", "bar"}
	parm, args := New(cmd, "-bar")
	if !reflect.DeepEqual(parm, Parm{
		"-bar": "FOO BAR",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"foo", "bar"}) {
		t.Error("wrong args:", args)
	}
}

func TestClearAndSet(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar=", "-bar=BAR"}
	parm, args := New(cmd, "-bar")
	if !reflect.DeepEqual(parm, Parm{
		"-bar": "BAR",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"foo"}) {
		t.Error("wrong args:", args)
	}
}

func TestClearAndSetPlus(t *testing.T) {
	cmd := []string{"foo", "-bar=FOO", "-bar=", "-bar=BAR", "bar"}
	parm, args := New(cmd, "-bar")
	if !reflect.DeepEqual(parm, Parm{
		"-bar": "BAR",
	}) {
		t.Error("wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"foo", "bar"}) {
		t.Error("wrong args:", args)
	}
}

func TestKexecBug(t *testing.T) {
	cmd := []string{"kexec", "-k", "foo.vmlinuz", "-i", "foo.initrd"}
	parm, args := New(cmd, "-c", "-i", "-k", "-l", "-x")
	if !reflect.DeepEqual(parm, Parm{
		"-c": "",
		"-i": "foo.initrd",
		"-k": "foo.vmlinuz",
		"-l": "",
		"-x": "",
	}) {
		t.Error("Wrong parm:", parm)
	}
	if !reflect.DeepEqual(args, []string{"kexec"}) {
		t.Error("wrong args:", args)
	}
}
