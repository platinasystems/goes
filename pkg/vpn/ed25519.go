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

type pureEd25519 struct {
	ed25519.PublicKey
}

func (v pureEd25519) verify(msg []byte) bool {
	n := len(msg)
	if n < ed25519.SignatureSize {
		xlog.Errata.Println(n, "<", ed25519.SignatureSize)
		return false
	}
	i := len(msg) - ed25519.SignatureSize
	return ed25519.Verify(v.PublicKey, msg[:i], msg[i:])
}

// NewEd25519 creates a PEM encoded ed25519 signature key file.
func NewEd25519(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Create PEM encoded ed25519 signature key file.

{{flags .}}`)

	DefineConfigFlag()
	DefineSigFlag()

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
	blk := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   der,
	}
	sp := vpnSigPath()
	if sp == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if _, err = os.Stat(sp); err == nil {
		return fmt.Errorf("%s: exists", sp)
	} else if !os.IsNotExist(err) {
		return err
	}
	dn := filepath.Dir(sp)
	if _, err = os.Stat(dn); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dn, 0755); err != nil {
			return err
		}
	}
	w, err := os.OpenFile(sp, oCreate, 0600)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}
