// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package goes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var Path = func() string {
	var p []string
	if hd, err := os.UserHomeDir(); err == nil {
		for _, s := range []string{
			"bin",
			"go/bin",
		} {
			ps := filepath.Join(hd, filepath.FromSlash(s))
			if _, err := os.Stat(ps); err == nil {
				p = append(p, ps)
			}
		}
	}
	for _, s := range []string{
		"/ko-app",
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	} {
		ps := filepath.FromSlash(s)
		if _, err := os.Stat(ps); err == nil {
			p = append(p, ps)
		}
	}
	return fmt.Sprint("PATH=", strings.
		Join(p, string(filepath.ListSeparator)))
}
