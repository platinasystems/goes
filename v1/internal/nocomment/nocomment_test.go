// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nocomment

import "testing"

func Test(t *testing.T) {
	var s string
	if s = New("hello # world"); s != "hello" {
		t.Errorf("unexpected: %q\n", s)
	}
	if s = New("# hello world"); s != "" {
		t.Errorf("unexpected: %q\n", s)
	}
	if s = New("#hello world"); s != "" {
		t.Errorf("unexpected: %q\n", s)
	}
	if s = New("hello#world"); s != "hello#world" {
		t.Errorf("unexpected: %q\n", s)
	}
	if s = New("hello #world"); s != "hello" {
		t.Errorf("unexpected: %q\n", s)
	}
}
