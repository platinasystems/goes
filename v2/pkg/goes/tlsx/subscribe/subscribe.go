// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package subscribe

import (
	"context"
	"errors"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
)

const Usage = `
usage:	{{.}} <name>[@<address>][:<port>]
Register with exchange.
`

var ErrIncomplete = errors.New("incomplete")

func Func(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) (err error) {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	return tlsx.Subscribe(ctx, args[0])
}
