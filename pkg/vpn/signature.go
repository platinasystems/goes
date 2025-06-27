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
	"path/filepath"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

type Sign = func([]byte) []byte

type PrivateSignatureKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

var PrivSigFileName = sync.OnceValue(func() string {
	return filepath.Join(vpnConfigDir, vpnSig)
})

var PrivSigFileData = sync.OnceValues(func() ([]byte, error) {
	if fn := PrivSigFileName(); fn != "-" {
		return os.ReadFile(fn)
	}
	return io.ReadAll(os.Stdin)
})

var PrivSigFilePEMBlocks = sync.OnceValues(func() ([]*pem.Block, error) {
	var blks []*pem.Block
	data, err := PrivSigFileData()
	if err != nil {
		return blks, err
	}
	for blk, r := pem.Decode(data); blk != nil; blk, r = pem.Decode(r) {
		if strings.HasSuffix(blk.Type, "PRIVATE KEY") {
			blks = append(blks, blk)
		}
	}
	return blks, err
})

var PrivSigFileKeys = sync.OnceValues(func() ([]PrivateSignatureKey, error) {
	var keys []PrivateSignatureKey
	blks, err := PrivSigFilePEMBlocks()
	if err != nil {
		return keys, err
	}
	for i, blk := range blks {
		k, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
		if err != nil {
			return nil, xerrors.Label(err, "block", i)
		} else {
			keys = append(keys, k.(PrivateSignatureKey))
		}
	}
	return keys, err
})

var FirstPrivSigFileKey = sync.OnceValues(func() (PrivateSignatureKey, error) {
	keys, err := PrivSigFileKeys()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, xerrors.Invalid(PrivSigFileName())
	}
	return keys[0], nil
})

var FirstPrivSigFileSign = sync.OnceValues(func() (Sign, error) {
	k, err := FirstPrivSigFileKey()
	if err != nil {
		return nil, err
	}
	switch t := k.(type) {
	// case *rsa.PrivateKey:
	// case *ecdsa.PrivateKey:
	case ed25519.PrivateKey:
		return func(msg []byte) []byte {
			return ed25519.Sign(t, msg)
		}, nil
	// case *ecdh.PrivateKey:
	default:
		xlog.Errata.Printf("%T: unsupported", t)
		return nil, fmt.Errorf("%T: %w", t, xerrors.ErrUnsupported)
	}
})

// ShowSignature prints algorithm.
func ShowSignature(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [<filename> | -]
Print algorithm.

{{flags .}}`)

	defineConfigDir()
	defineSig()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flag.CommandLine.NArg() > 0 {
		PrivSigFileName = func() string {
			return flag.CommandLine.Arg(0)
		}
	}
	keys, err := PrivSigFileKeys()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Println("# none")
		return nil
	}
	for _, k := range keys {
		var s string
		switch t := k.(type) {
		case *rsa.PrivateKey:
			s = "RSA"
		case *ecdsa.PrivateKey:
			s = "ECDSA"
		case ed25519.PrivateKey:
			s = "Ed25519"
		case *ecdh.PrivateKey:
			c := t.Curve()
			if cs, ok := c.(fmt.Stringer); ok {
				s = cs.String()
			} else {
				s = fmt.Sprintf("%T", c)
			}
		default:
			s = fmt.Sprintf("%T", t)
		}
		if len(keys) > 1 {
			fmt.Print("- ")
		}
		fmt.Println(s)
	}
	return nil
}

var Verify = ed25519.Verify
