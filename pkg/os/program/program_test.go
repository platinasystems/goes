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
	t.Run("is_ko_app:", func(t *testing.T) {
		if IsKoApp() {
			t.Error("unexpected")
		}
	})
	t.Run("is_opt:", func(t *testing.T) {
		if IsOpt() {
			t.Error("unexpected")
		}
	})
	t.Run("is_usr_local:", func(t *testing.T) {
		if IsUsrLocal() {
			t.Error("unexpected")
		}
	})
	t.Run("build_id:", func(t *testing.T) {
		if s := BuildId(); len(s) == 0 {
			t.Error("empty")
		}
	})
	t.Run("build_nfo:", func(t *testing.T) {
		if bi := BuildInfo(); bi == nil {
			t.Error("nil")
		}
	})
	t.Run("main_reference:", func(t *testing.T) {
		if s := MainReference(); len(s) > 0 {
			t.Error("test shouldn't have main reference:", s)
		}
	})
	t.Run("main_version:", func(t *testing.T) {
		if s := MainVersion(); len(s) > 0 {
			t.Error("test shouldn't have main version:", s)
		}
	})
}
