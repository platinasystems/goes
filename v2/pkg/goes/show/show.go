// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/encoding/ensure"
)

// This returns a method that prints its embedded value/Formatter/String.
func New(v any) func(
	context.Context,
	io.Reader,
	io.Writer,
	[]string,
	...string,
) error {
	return show{v}.show
}

type show struct{ v any }

func (sh show) show(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(`
usage: {{.}}
	Print named value.
`[1:])).Execute(w, strings.Join(path, " "))
	}

	nw := ensure.Newline(w)
	defer nw.Close()
	fmt.Fprint(nw, sh.v)
	return ctx.Err()
}
