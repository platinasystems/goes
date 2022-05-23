// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package selection

import (
	"bytes"
	"flag"
	"fmt"
	"io"
)

type Path []string

func (p Path) HasComplete() bool {
	return len(p) > 1 && p[1] == "complete"
}

func (p Path) HasHelp() bool {
	return len(p) > 1 && p[1] == "help"
}

func (p Path) Usage(w io.Writer, args ...any) {
	fmt.Fprint(w, "usage:")
	for i, s := range p {
		if i != 1 || s != "help" {
			fmt.Fprint(w, " ", s)
		}
	}
	var buf bytes.Buffer
	p.args(&buf, args...)
	b := buf.Bytes()
	if n := len(b); n > 0 {
		if b[0] != ' ' && b[0] != '\n' {
			fmt.Fprint(w, " ")
		}
		w.Write(b)
		if b[n-1] != '\n' {
			fmt.Fprintln(w)
		}
	}
}

func (p Path) args(w io.Writer, args ...any) {
	for _, v := range args {
		if fs, ok := v.(*flag.FlagSet); ok {
			fmt.Fprintln(w)
			fs.SetOutput(w)
			fs.PrintDefaults()
		} else if slice, ok := v.([]any); ok {
			p.args(w, slice...)
		} else {
			fmt.Fprint(w, v)
		}
	}

}
