// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package admin

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/registry"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

const Usage = `
usage: {{.}} [<subject-key-id(s)>]
List or approve/deny subscriber requests.`

// PEM block headers to subscriber certificates.
var Headers = make(map[string]string)

func Func(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	usage := func() error {
		return template.Must(template.New("usage").
			Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}
	if len(args) == 0 {
		registry.Range(func(c *x509.Certificate, ski string) bool {
			fmt.Fprint(w, ski, ": ", c.DNSNames, "\n")
			return true
		})
		return nil
	}
	approve := path[len(path)-1] == "approve"
	for _, arg := range args {
		c := registry.Extract(
			func(c *x509.Certificate, ski string) bool {
				return ski == arg
			})
		if c == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		if approve {
			err := certs.Subscribers.Add(Headers, c)
			if err != nil {
				return err
			}
			certs.ClientCAs.Add(c)
		}
	}
	return nil
}
