// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package notliner provides an alternative command.Prompter for shell scripts
// and tty's unsupported by liner.
package notliner

import (
	"bufio"
	"fmt"
	"io"
)

type Prompter struct {
	scanner *bufio.Scanner
	w       io.Writer
}

func New(r io.Reader, w io.Writer) *Prompter {
	return &Prompter{bufio.NewScanner(r), w}
}

func (p *Prompter) Close() {
}

func (p *Prompter) Prompt(prompt string) (string, error) {
	if p.w != nil {
		fmt.Fprint(p.w, prompt)
	}
	if p.scanner.Scan() {
		return p.scanner.Text(), nil
	}
	err := p.scanner.Err()
	if err == nil {
		err = io.EOF
	}
	return "", err
}
