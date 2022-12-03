// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package address

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/address"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

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
		return template.Must(template.New("usage").Parse(`
usage {{.}} [<exchange> [<address>]]
Assign or print exchange address.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	switch len(args) {
	case 0:
		address.Range(func(ski, addr string) bool {
			fmt.Fprint(w, ski, ": ", addr, "\n")
			return true
		})
	case 1:
		if name, ski, c := tlsx.Lookup(args[0]); c == nil {
			err = certs.Unsubscribed(args[0])
		} else if addr, ok := address.Load(ski); !ok {
			err = address.Unaddressed(ski, name)
		} else {
			fmt.Fprintln(w, addr)
		}
	case 2:
		if _, ski, c := tlsx.Lookup(args[0]); c == nil {
			err = certs.Unsubscribed(args[0])
		} else {
			err = address.Store(ski, args[1])
		}
	default:
		err = fmt.Errorf("%v: unexpected", args[2:])
	}
	return
}
