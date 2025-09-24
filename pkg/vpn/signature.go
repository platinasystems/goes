// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
)

var (
	sign     func([]byte) []byte
	signPriv crypto.PrivateKey
	signPub  crypto.PublicKey
)

var SigFile = xflag.New[string]("sig", `
Signature file w/in current or config directory.
`[1:], func() string {
	s, ok := xmain.LookupEnv("SIG")
	if !ok {
		s = "sig.pk8"
	}
	return s
})

func SigPath() string {
	return xmain.Config.File(SigFile.Value())
}

// ShowSignature prints algorithm.
func ShowSignature(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [<filename> | -]
Print algorithm.

{{flags .}}`)

	xmain.Config.Define()
	SigFile.Define()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flag.CommandLine.NArg() > 0 {
		SigFile.Override(flag.CommandLine.Arg(0))
	}
	if err = signInit(); err != nil {
		return err
	}
	switch t := signPriv.(type) {
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

func signInit() error {
	var data []byte
	var err error

	if signPriv != nil {
		return err
	}
	input := SigPath()
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
	signPriv, err = x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return xerrors.Label(err, input)
	}
	switch t := signPriv.(type) {
	// case *rsa.PrivateKey:
	// case *ecdsa.PrivateKey:
	case ed25519.PrivateKey:
		sign = signEd25519
		signPub = t.Public()
	// case *ecdh.PrivateKey:
	default:
		return fmt.Errorf("%T: %w", t, xerrors.ErrUnsupported)
	}
	return nil
}

func signEd25519(data []byte) []byte {
	return ed25519.Sign(signPriv.(ed25519.PrivateKey), data)
}
