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
	"io"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/x509keys"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xpem"
)

func generateEd25519Key(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [-]
Generate PEM encoded ed25519 key.

The default output is “{{key}}”;
use “-” for stdout.
`)
	xflag.UsageFuncs["key"] = keyPath
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	k, _ := keyFile()
	if len(args) > 0 {
		k.Path = args[0]
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
	if k.Path == "-" {
		err = xpem.EncodeAll(os.Stdout, block)
	} else {
		err = xpem.Create(k.Path, 0600, block)
	}
	return err
}

func showEd25519Key(ctx context.Context, args []string) error {
	var rc io.ReadCloser
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [-]
Print parsed key confirmation.

The default output is “{{.Key}}”;
use “-” for stdin.
`)
	xflag.UsageFuncs["key"] = keyPath
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	path := keyPath()
	if len(args) > 0 {
		path = args[0]
	}
	if path == "-" {
		rc = os.Stdin
	} else if rc, err = os.Open(path); err != nil {
		return err
	} else {
		defer rc.Close()
	}

	key := &x509keys.File{Path: path}
	if _, err = key.ReadFrom(rc); err == nil {
		err = key.Show(os.Stdout)
	}
	return err
}
