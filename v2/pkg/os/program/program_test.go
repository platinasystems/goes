// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package program

import "testing"

func Test(t *testing.T) {
	t.Run("executable:", func(t *testing.T) {
		v, err := Executable.Value()
		if err != nil {
			t.Error(err)
		} else if len(v) == 0 {
			t.Error("empty")
		}
	})
	t.Run("base:", func(t *testing.T) {
		v, err := Base.Value()
		if err != nil {
			t.Error(err)
		} else if v != "program.test" {
			t.Errorf("%q", v)
		}
	})
	t.Run("build-id:", func(t *testing.T) {
		v, err := BuildId.Value()
		if err != nil {
			t.Error(err)
		} else if len(v) == 0 {
			t.Error("empty")
		}
	})
	t.Run("build-info:", func(t *testing.T) {
		v, err := BuildInfo.Value()
		if err != nil {
			t.Error(err)
		} else if v == nil {
			t.Error("nil")
		}
	})
	t.Run("is /opt:", func(t *testing.T) {
		v, err := IsOpt.Value()
		if err != nil {
			t.Error(err)
		} else if v {
			t.Error("unexpected")
		}
	})
	t.Run("is /usr/local:", func(t *testing.T) {
		v, err := IsUsrLocal.Value()
		if err != nil {
			t.Error(err)
		} else if v {
			t.Error("unexpected")
		}
	})
	t.Run("is su:", func(t *testing.T) {
		v, err := IsSuperUser.Value()
		if err != nil {
			t.Error(err)
		} else if v {
			t.Error("unexpected")
		}
	})
	t.Run("main reference:", func(t *testing.T) {
		v, err := MainReference.Value()
		if err != ErrUnavailable {
			t.Error("test shouldn't have main reference:", v)
		}
	})
	t.Run("main version:", func(t *testing.T) {
		v, err := MainVersion.Value()
		if err != ErrUnavailable {
			t.Error("test shouldn't have main version:", v)
		}
	})
}
