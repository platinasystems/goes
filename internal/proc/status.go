// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package proc

import (
	"bufio"
	"bytes"
	"io"
)

// Linux /proc/<PID|"self">/status
type Status map[string]string

func (m Status) Load(r io.Reader) error {
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		b := scan.Bytes()
		colon := bytes.Index(b, []byte(":"))
		if colon > 0 {
			m[string(b[:colon])] =
				string(bytes.TrimSpace(b[colon+1:]))
		}
	}
	return scan.Err()
}
