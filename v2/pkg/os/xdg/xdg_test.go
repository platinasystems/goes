// Copyright © 2022 Platina Systems, Inc. All rights reserved.
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
		program.IsSuperUser.Preload(func(p *bool) {
			*p = true
		})
		program.IsOpt.Preload(func(p *bool) {
			*p = true
		})
		program.IsUsrLocal.Preload(func(p *bool) {
			*p = true
		})
		Root.CacheHome.Preload(func(p *string) {
			*p = "/var/cache"
		})
		t.Run("XDG_CACHE_HOME", unit{
			CacheHome, "/var/cache",
		}.test)
		t.Run("XDG_CONFIG_HOME", unit{
			ConfigHome, "/etc/opt",
		}.test)
		t.Run("XDG_DATA_HOME", unit{
			DataHome, "/usr/local/share",
		}.test)
		Root.RunTimeDir.Preload(func(p *string) {
			*p = "/var/run"
		})
		t.Run("XDG_RUNTIME_DIR", unit{
			RunTimeDir, "/var/run",
		}.test)
		t.Run("XDG_STATE_HOME", unit{
			StateHome, "/var/local",
		}.test)
	})
	t.Run("user", func(t *testing.T) {
		program.IsSuperUser.Preload(func(p *bool) {
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
			xdg.CacheHome.Reload()
			t.Run("XDG_CACHE_HOME", unit{
				CacheHome, "$HOME/.cache",
			}.test)
			xdg.ConfigHome.Reload()
			t.Run("XDG_CONFIG_HOME", unit{
				ConfigHome, "$HOME/.config",
			}.test)
			xdg.DataHome.Reload()
			t.Run("XDG_DATA_HOME", unit{
				DataHome, "$HOME/.local/share",
			}.test)
			xdg.RunTimeDir.Reload()
			t.Run("XDG_RUNTIME_DIR", unit{
				RunTimeDir, "/run/user/$ID",
			}.test)
			xdg.StateHome.Reload()
			t.Run("XDG_STATE_HOME", unit{
				StateHome, "$HOME/.local/state",
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
			xdg.CacheHome.Reload()
			t.Run("XDG_CACHE_HOME", unit{
				CacheHome, "$HOME/.cache",
			}.test)
			xdg.ConfigHome.Reload()
			t.Run("XDG_CONFIG_HOME", unit{
				ConfigHome, "$HOME/.config",
			}.test)
			xdg.DataHome.Reload()
			t.Run("XDG_DATA_HOME", unit{
				DataHome, "$HOME/.local/share",
			}.test)
			xdg.RunTimeDir.Reload()
			t.Run("XDG_RUNTIME_DIR", unit{
				RunTimeDir, "$HOME/.cache",
			}.test)
			xdg.StateHome.Reload()
			t.Run("XDG_STATE_HOME", unit{
				StateHome, "$HOME/.local/state",
			}.test)
		})
	})
}
