// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

// NewEd25519 creates a PEM encoded ed25519 signature key file.
func NewEd25519(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Create PEM encoded ed25519 signature key file.

{{flags .}}`)

	defineCommonFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if vpnQuiet {
		xlog.MuteErrata()
	} else if vpnVerbose {
		xlog.UnmuteInfo()
	}

	sfn := filepath.Join(vpnConfigDir, vpnSig)

	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if sfn == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if _, err = os.Stat(sfn); err == nil {
		return fmt.Errorf("%s: exists", sfn)
	} else if !os.IsNotExist(err) {
		return err
	}
	dn := filepath.Dir(sfn)
	if _, err = os.Stat(dn); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dn, 0755); err != nil {
			return err
		}
	}
	w, err := os.OpenFile(sfn, oCreate, 0600)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}
