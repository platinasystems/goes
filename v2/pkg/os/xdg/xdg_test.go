// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdg

import (
	"testing"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

type unit struct {
	get  func() string
	want string
}

func (ut unit) test(t *testing.T) {
	t.Helper()
	if got := ut.get(); got != ut.want {
		if len(got) == 0 {
			got = "\"\""
		}
		t.Error(got)
	}
}

func Test(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		program.Opt.Preload(func(p *bool) {
			*p = true
		})
		program.SuperUser.Preload(func(p *bool) {
			*p = true
		})
		program.UsrLocal.Preload(func(p *bool) {
			*p = true
		})
		SU.CacheHome.Preload(func(p *string) {
			*p = "/var/cache"
		})
		t.Run("XDG_CACHE_HOME", unit{
			CacheHome.Value, "/var/cache",
		}.test)
		t.Run("XDG_CONFIG_HOME", unit{
			ConfigHome.Value, "/etc/opt",
		}.test)
		t.Run("XDG_DATA_HOME", unit{
			DataHome.Value, "/usr/local/share",
		}.test)
		SU.RunTimeDir.Preload(func(p *string) {
			*p = "/var/run"
		})
		t.Run("XDG_RUNTIME_DIR", unit{
			RunTimeDir.Value, "/var/run",
		}.test)
		t.Run("XDG_STATE_HOME", unit{
			StateHome.Value, "/var/local",
		}.test)
	})
	t.Run("user", func(t *testing.T) {
		program.SuperUser.Preload(func(p *bool) {
			*p = false
		})
		t.Run("env", func(t *testing.T) {
			Getenv = func(name string) string {
				return map[string]string{
					"XDG_CACHE_HOME":  "$HOME/.cache",
					"XDG_CONFIG_HOME": "$HOME/.config",
					"XDG_DATA_HOME":   "$HOME/.local/share",
					"XDG_RUNTIME_DIR": "/run/user/$ID",
					"XDG_STATE_HOME":  "$HOME/.local/state",
				}[name]
			}
			CacheHome.Invalidate()
			t.Run("XDG_CACHE_HOME", unit{
				CacheHome.Value, "$HOME/.cache",
			}.test)
			ConfigHome.Invalidate()
			t.Run("XDG_CONFIG_HOME", unit{
				ConfigHome.Value, "$HOME/.config",
			}.test)
			DataHome.Invalidate()
			t.Run("XDG_DATA_HOME", unit{
				DataHome.Value, "$HOME/.local/share",
			}.test)
			RunTimeDir.Invalidate()
			t.Run("XDG_RUNTIME_DIR", unit{
				RunTimeDir.Value, "/run/user/$ID",
			}.test)
			StateHome.Invalidate()
			t.Run("XDG_STATE_HOME", unit{
				StateHome.Value, "$HOME/.local/state",
			}.test)
		})
		t.Run("noenv", func(t *testing.T) {
			Getenv = func(name string) string { return "" }
			UserCacheDir = func() (string, error) {
				return "$HOME/.cache", nil
			}
			UserConfigDir = func() (string, error) {
				return "$HOME/.config", nil
			}
			UserHomeDir = func() (string, error) {
				return "$HOME", nil
			}
			CacheHome.Invalidate()
			t.Run("XDG_CACHE_HOME", unit{
				CacheHome.Value, "$HOME/.cache",
			}.test)
			ConfigHome.Invalidate()
			t.Run("XDG_CONFIG_HOME", unit{
				ConfigHome.Value, "$HOME/.config",
			}.test)
			DataHome.Invalidate()
			t.Run("XDG_DATA_HOME", unit{
				DataHome.Value, "$HOME/.local/share",
			}.test)
			RunTimeDir.Invalidate()
			t.Run("XDG_RUNTIME_DIR", unit{
				RunTimeDir.Value, "$HOME/.cache",
			}.test)
			StateHome.Invalidate()
			t.Run("XDG_STATE_HOME", unit{
				StateHome.Value, "$HOME/.local/state",
			}.test)
		})
	})
}
