// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/show"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var Show = selection.Map{
	"build-id":       show.Func(program.BuildId),
	"build-info":     show.Func(program.BuildInfo),
	"cert":           showCert,
	"clients":        showClients,
	"exchanges":      showExchanges,
	"main-reference": show.Func(program.MainReference),
	"version":        show.Func(program.MainVersion),
}

func showCert(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "\n",
			"Print local certificate.",
		)
		return nil
	}
	tlsc, err := cert.Value()
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
	certs.Fsequent(w, certs.Headers{}, tlsc.Leaf)
	return nil
}

func showClients(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "\nList client certificates of exchange.")
		return nil
	}
	certs.Clients.Range(func(
		headers certs.Headers,
		cl *x509.Certificate,
	) bool {
		certs.Fsequent(w, headers, cl)
		return true
	})
	return nil
}

func showExchanges(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "\nList exchange certificates of "+
			"host or consumer.")
		return nil
	}
	certs.Exchanges.Range(func(
		headers certs.Headers,
		ex *x509.Certificate,
	) bool {
		certs.Fsequent(w, headers, ex)
		return true
	})
	return nil
}
