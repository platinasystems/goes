// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xpem"
)

// Ed25519 creates a PEM encoded ed25519 signature key file.
func Ed25519(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Create PEM encoded ed25519 signature key file.

{{flags .}}`)

	opts.key = flag.String("o", DefaultKey(),
		"Output file name or “-” for stdout.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if *opts.key == "-" {
		err = xpem.EncodeAll(os.Stdout, block)
	} else {
		err = xpem.Create(*opts.key, 0600, block)
	}
	return err
}
