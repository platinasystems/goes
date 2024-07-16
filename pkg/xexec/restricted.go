// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xexec

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("not found")

func RestrictedLookPath(name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) {
		if IsExecutable(name) {
			return name, nil
		}
		return "", ErrNotFound
	}
	for _, d := range RestrictedCurrentUserPath() {
		if full := filepath.Join(d, name); IsExecutable(full) {
			return full, nil
		}
	}
	return "", ErrNotFound
}
