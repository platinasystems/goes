// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sig

import (
	"context"
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

var (
	Priv crypto.PrivateKey
	Pub  crypto.PublicKey

	Same func(cert *x509.Certificate) bool
	Sign func([]byte) []byte

	Schemes = []tls.SignatureScheme{
		tls.Ed25519,
	}
)

var (
	File = "sig.pk8"
	Flag = xflag.Label{"sig", `
Signature file w/in current or config directory.
(or $<main>_KEY_FILE, $SSL_KEY_FILE)`[1:], func() any {
		if s, ok := xmain.LookupEnv("KEY_FILE"); ok {
			File = s
		} else if s, ok = os.LookupEnv("SSL_KEY_FILE"); ok {
			File = s
		}
		return &File
	}}
	Path     = func() string { return xmain.ConfigFile(File) }
	Features = map[string]any{
		"new": map[string]any{
			"signature": New,
		},
		"show": map[string]any{
			"signature": Show,
		},
	}
)

// Initialize [Sign], [Priv], and [Pub] from [File].
func Init() error {
	var data []byte
	var err error

	if Priv != nil {
		return nil
	}
	input := Path()
	if input == "-" {
		input = "input"
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(input)
	}
	if err != nil {
		return err
	}
	blk, _ := pem.Decode(data)
	if blk == nil || !strings.HasSuffix(blk.Type, "PRIVATE KEY") {
		return xerrors.Invalid(input)
	}
	Priv, err = x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return xerrors.Label(err, input)
	}
	switch t := Priv.(type) {
	// case *rsa.PrivateKey:
	// case *ecdsa.PrivateKey:
	case ed25519.PrivateKey:
		Pub = t.Public()
		Same = sameEd25519
		Sign = signEd25519
	// case *ecdh.PrivateKey:
	default:
		return fmt.Errorf("%T: %w", t, xerrors.ErrUnsupported)
	}
	return nil
}

// New creates a PEM encoded signature key file.
func New(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Create PEM encoded  signature key file.

{{flags .}}`)
	kind := "ed25519"
	err := xflag.Labels{
		xmain.ConfigFlag,
		Flag,
		// {"kind", "only ed25519", &kind},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	var priv any
	switch kind {
	case "ed25519":
		_, priv, err = ed25519.GenerateKey(nil)
	default:
		return xerrors.Unsupported(kind)
	}
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
	sp := Path()
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
	w, err := xos.TruncFile(sp, 0600)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Prints algorithm of signature file.
func Show(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [<filename> | -]
Print algorithm.

{{flags .}}`)
	err := xflag.Labels{
		xmain.ConfigFlag,
		Flag,
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	if flag.CommandLine.NArg() > 0 {
		File = flag.CommandLine.Arg(0)
	}
	if err = Init(); err != nil {
		return err
	}
	switch t := Priv.(type) {
	case *rsa.PrivateKey:
		fmt.Println("RSA")
	case *ecdsa.PrivateKey:
		fmt.Println("ECDSA")
	case ed25519.PrivateKey:
		fmt.Println("Ed25519")
	case *ecdh.PrivateKey:
		c := t.Curve()
		if cs, ok := c.(fmt.Stringer); ok {
			fmt.Println(cs)
		} else {
			fmt.Printf("%T\n", c)
		}
	default:
		fmt.Printf("%T\n", t)
	}
	return nil
}
