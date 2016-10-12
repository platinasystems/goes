// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package slice_args

import (
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	pl := New("|")
	pl.Slice("ls", "-lR", "|", "more")
	if !reflect.DeepEqual(pl.Slices, [][]string{
		{"ls", "-lR"},
		{"more"},
	}) {
		t.Error("unexpected:", pl.Slices)
	}
	if pl.More {
		t.Error("more?")
	}

	pl.Reset()
	if !reflect.DeepEqual(pl.Slices, [][]string{}) {
		t.Error("unexpected:", pl.Slices)
	}

	pl.Slice("ls", "-lR", "|")
	if !reflect.DeepEqual(pl.Slices, [][]string{
		{"ls", "-lR"},
	}) {
		t.Error("unexpected:", pl.Slices)
	}
	if !pl.More {
		t.Error("no more?")
	}
	pl.Slice("more")
	if pl.More {
		t.Error("more?")
	}
	if !reflect.DeepEqual(pl.Slices, [][]string{
		{"ls", "-lR"},
		{"more"},
	}) {
		t.Error("unexpected:", pl.Slices)
	}

}
