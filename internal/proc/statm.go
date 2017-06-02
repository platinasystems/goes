// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package proc

import (
	"fmt"
	"io"
)

// Linux /proc/<PID|"self">/statm
type Statm struct {
	Size, Resident, Share, Text, Lib, Data, Dt uint64
}

func (p *Statm) Load(r io.Reader) error {
	for _, x := range []struct {
		s string
		v interface{}
	}{
		{"Size", &p.Size},
		{"Resident", &p.Resident},
		{"Share", &p.Share},
		{"Text", &p.Text},
		{"Lib", &p.Lib},
		{"Data", &p.Data},
		{"Dt", &p.Dt},
	} {
		if _, err := fmt.Fscan(r, x.v); err != nil {
			return fmt.Errorf("%s: %v", x.s, err)
		}
	}
	return nil
}
