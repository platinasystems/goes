// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build windows || plan9 || ((darwin || ios) && !cgo)

package oslog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func OpenError() (io.WriteCloser, error) {
	return openFile(fmt.Sprint(filepath.Base(os.Args[0]), "_error.log"))
}

func OpenNotice() (io.WriteCloser, error) {
	return openFile(fmt.Sprint(filepath.Base(os.Args[0]), ".log"))
}

func OpenInfo() (io.WriteCloser, error) {
	return openFile(fmt.Sprint(filepath.Base(os.Args[0]), "_info.log"))
}

// If directory exists, create ~/Library/Logs/FN; otherwise, TMPDIR/FN.
func openFile(fn string) (io.WriteCloser, error) {
	dn := os.TempDir()
	if h, err := os.UserHomeDir(); err == nil {
		hl := filepath.Join(h, "/Library/Logs")
		if fi, err := os.Stat(hl); err == nil && fi.IsDir() {
			dn = hl
		}
	}
	return os.Create(filepath.Join(dn, fn))
}
