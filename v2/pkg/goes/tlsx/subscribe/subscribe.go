// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package subscribe

import (
	"context"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/subscribe"
)

const Usage = `
usage:	{{.}} <name>:<port>
	{{.}} <name> <address>:<port>
Register with exchange.
`

// PEM block headers to subscription certificates.
var Headers = map[string]string{}

func Func(
	ctx context.Context,
	r io.Reader,
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
		return template.Must(template.New("usage").
			Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	return subscribe.Req(ctx, Headers, args)
}
