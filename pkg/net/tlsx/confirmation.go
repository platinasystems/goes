// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

type Confirmation uint64

func (p *Confirmation) ReadFrom(r io.Reader) (int64, error) {
	var buf [8]byte
	n, err := r.Read(buf[:])
	if err != nil {
		return int64(n), egress.Mark(err)
	}
	if n != len(buf[:]) {
		return int64(n), egress.Mark(ErrUnderrun)
	}
	err = p.UnmarshalBinary(buf[:])
	return int64(n), err
}

func (cno Confirmation) String() string { return fmt.Sprintf("%d", cno) }

func (p *Confirmation) UnmarshalBinary(data []byte) error {
	*p = Confirmation(ByteOrder.Uint64(data))
	return nil
}

func (cno Confirmation) WriteTo(w io.Writer) (int64, error) {
	var buf [8]byte
	ByteOrder.PutUint64(buf[:], uint64(cno))
	n, err := w.Write(buf[:])
	return int64(n), err
}
