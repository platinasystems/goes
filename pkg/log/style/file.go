// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build windows || plan9 || ((darwin || ios) && !cgo)

package style

import (
	"fmt"
	"os"
	"path/filepath"
)

var file *os.File

// If directory exists, log to ~/Library/Logs/PROGRAM.log; otherwise,
// log to TMPDIR/PGROGRAM.log instead of Std{out|err}.
func (style Style) System() {
	var prefix string
	if file == nil {
		bn := fmt.Sprint(filepath.Base(os.Args[0]), ".log")
		dn := os.TempDir()
		if h, err := os.UserHomeDir(); err == nil {
			hl := filepath.Join(h, "/Library/Logs")
			if fi, err := os.Stat(hl); err == nil && fi.IsDir() {
				dn = hl
			}
		}
		fn := filepath.Join(dn, bn)
		if f, err := os.Create(fn); err != nil {
			return
		} else {
			file = f
		}
	}
	if style.Level != Notice {
		prefix = fmt.Sprint(style.Level, ":")
		if style.Flags() == plain {
			prefix += " "
		}
	}
	style.SetOutput(file)
	style.SetPrefix(prefix)
}
