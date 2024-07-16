// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chunk

import "testing"

func Test4(t *testing.T) {
	data := Alloc(4)
	if n := len(data); n != 4 {
		t.Fatal("len:", n)
	}
	if n := cap(data); n != 16 {
		t.Fatal("cap:", n)
	}
	Free(data)
	if Cap16.free.data.Load() == nil {
		t.Fatal("not pooled")
	}
	data = Alloc(8)
	if n := len(data); n != 8 {
		t.Fatal("len:", n)
	}
	if n := cap(data); n != 16 {
		t.Fatal("cap:", n)
	}
	if Cap16.free.data.Load() != nil {
		t.Fatal("pool not empty")
	}
}

func TestPage(t *testing.T) {
	page := Page()
	data := page.Alloc(4)
	if n := len(data); n != 4 {
		t.Fatal("len:", n)
	}
	if n := cap(data); n != page.Cap {
		t.Fatal("cap:", n)
	}
	page.Free(data)
	if page.free.data.Load() == nil {
		t.Fatal("not pooled")
	}
	data = page.Alloc(8)
	if n := len(data); n != 8 {
		t.Fatal("len:", n)
	}
	if n := cap(data); n != page.Cap {
		t.Fatal("cap:", n)
	}
	if page.free.data.Load() != nil {
		t.Fatal("pool not empty")
	}
}
