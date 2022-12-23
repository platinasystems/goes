// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package program

import "testing"

func Test(t *testing.T) {
	t.Run("executable:", func(t *testing.T) {
		if v := Executable(); len(v) == 0 {
			t.Error("empty")
		}
	})
	t.Run("base:", func(t *testing.T) {
		if v := Base(); v != "program.test" {
			t.Errorf("%q", v)
		}
	})
	t.Run("is /opt:", func(t *testing.T) {
		if Is.Opt() {
			t.Error("unexpected")
		}
	})
	t.Run("is /usr/local:", func(t *testing.T) {
		if Is.UsrLocal() {
			t.Error("unexpected")
		}
	})
	t.Run("is su:", func(t *testing.T) {
		if Is.SuperUser() {
			t.Error("unexpected")
		}
	})
	t.Run("build:", func(t *testing.T) {
		t.Run("id:", func(t *testing.T) {
			if v, err := BuildId.ValErr(); err != nil {
				t.Error(err)
			} else if len(v) == 0 {
				t.Error("empty")
			}
		})
		t.Run("info:", func(t *testing.T) {
			if v, err := BuildInfo.ValErr(); err != nil {
				t.Error(err)
			} else if v == nil {
				t.Error("nil")
			}
		})
	})
	t.Run("main:", func(t *testing.T) {
		t.Run("reference:", func(t *testing.T) {
			v, err := MainReference.ValErr()
			if err != ErrUnavailable {
				t.Error("test shouldn't have main reference:", v)
			}
		})
		t.Run("version:", func(t *testing.T) {
			v, err := MainVersion.ValErr()
			if err != ErrUnavailable {
				t.Error("test shouldn't have main version:", v)
			}
		})
	})
}
