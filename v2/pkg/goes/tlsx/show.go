// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

func ShowCert(
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
Print local certificate.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	tlsc, err := cert.ValErr()
	if err != nil {
		return err
	}
	if n := len(tlsc.SupportedSignatureAlgorithms); n > 0 {
		fmt.Fprintln(w, "supported_signature_algoritums:")
		for _, alg := range tlsc.SupportedSignatureAlgorithms {
			fmt.Fprintln(w, "  -", alg)
		}
	}
	if n := len(tlsc.SignedCertificateTimestamps); n > 0 {
		fmt.Fprintln(w, "signed_certificate_timestamps:", n)
	}
	fmt.Fprint(w, certs.NewCert(tlsc.Leaf))
	return nil
}

func ShowSubs(
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
List certificates.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	subs := certs.Subscribers
	if path[len(path)-1] == "subscriptions" {
		subs = certs.Subscriptions
	}
	fmt.Fprint(w, subs)
	return nil
}
