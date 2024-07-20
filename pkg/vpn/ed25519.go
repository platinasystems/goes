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

func newEd25519Key(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Generate PEM encoded ed25519 key.

{{flags .}}`)

	opts.key = flag.String("o", defaultKey(),
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

func showEd25519Key(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Print parsed key confirmation.

{{flags .}}`)

	opts.key = flag.String("i", defaultKey(),
		"Input file name or “-” for stdin.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	key, err := keyFile()
	if err == nil {
		err = key.Show(os.Stdout)
	}
	return err
}
