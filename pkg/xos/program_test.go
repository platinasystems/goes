// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xos

import "testing"

func TestProgram(t *testing.T) {
	t.Run("program:", func(t *testing.T) {
		if v := Program(); len(v) == 0 {
			t.Error("empty")
		}
	})
	t.Run("base:", func(t *testing.T) {
		if v := Arg0Base(); v != "xos.test" {
			t.Errorf("%q", v)
		}
	})
	t.Run("is_ko_app:", func(t *testing.T) {
		if ProgramIsKoApp() {
			t.Error("unexpected")
		}
	})
	t.Run("is_opt:", func(t *testing.T) {
		if ProgramIsOpt() {
			t.Error("unexpected")
		}
	})
	t.Run("is_usr_local:", func(t *testing.T) {
		if ProgramIsUsrLocal() {
			t.Error("unexpected")
		}
	})
}
