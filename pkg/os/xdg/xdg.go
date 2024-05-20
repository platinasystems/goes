// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides [XDG] or, if Unix super user, [FSHS] defined paths.
//
//	XDG  https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
//	FSHS https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard
package xdg

import (
	"os"
)

// test overrides
var (
	Getenv        = os.Getenv
	UserCacheDir  = os.UserCacheDir
	UserConfigDir = os.UserConfigDir
	UserHomeDir   = os.UserHomeDir
)
