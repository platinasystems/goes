// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package unprompted provied a command.GetLiner for scripts.
package unprompted

import (
	"bufio"
	"io"
)

type Unprompted bufio.Reader

func New(r io.Reader) *Unprompted {
	return (*Unprompted)(bufio.NewReader(r))
}

// Use Unprompted.GetLine to source script files.
func (r *Unprompted) GetLine(prompt string) (string, error) {
	return (*bufio.Reader)(r).ReadString('\n')
}
