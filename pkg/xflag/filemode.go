// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"fmt"
	"os"
	"strconv"
)

type FileMode os.FileMode

func (m *FileMode) Mode() os.FileMode {
	return os.FileMode(*m)
}

func (m *FileMode) Set(s string) error {
	u, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid file mode: %w", err)
	}
	*m = FileMode(u)
	return nil
}

func (m *FileMode) String() string {
	return fmt.Sprintf("%#o", *m)
}
