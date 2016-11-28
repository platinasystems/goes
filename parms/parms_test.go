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
