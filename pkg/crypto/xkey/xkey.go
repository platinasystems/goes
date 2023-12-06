// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xkey

import (
	"context"
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

type Privateer interface {
	Equal(x crypto.PrivateKey) bool
	Public() crypto.PublicKey
}

type Private struct {
	*pem.Block
	Privateer
}

var (
	ErrInvalid     = fs.ErrInvalid
	ErrUnsupported = errors.ErrUnsupported
)

const GenerateUsageTemplate = `
usage: {{.}} [<algorithm>]
Generate PEM encoded private signature key to stdout with given or
default, ed25519 algorithm.

Algorithms
  ecdsa
  ed25519
  x25519
  rsa`

func GenerateUsageData(ctx context.Context) any {
	return strings.Join(ctxparm.Strings.In(ctx), " ")
}

func Generate(ctx context.Context, args ...string) error {
	if *complete.Help {
		if len(args) > 0 {
			return complete.Last(args, []string{
				"ecdsa",
				"ed25519",
				"x25519",
				"rsa",
			})

		}
		return nil
	}
	if *usage.Help {
		return usage.Error(GenerateUsageTemplate[1:],
			GenerateUsageData(ctx))
	}
	w := ctxparm.Writer.In(ctx)
	alg := "ed25519"
	if len(args) > 0 {
		alg = args[0]
	}
	var priv Privateer
	var err error
	random := rand.Reader
	t := "PRIVATE KEY"
	switch alg {
	case "ecdsa":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), random)
	case "ed25519":
		_, priv, err = ed25519.GenerateKey(random)
	case "x25519":
		priv, err = ecdh.X25519().GenerateKey(random)
	case "rsa":
		priv, err = rsa.GenerateKey(random, 4096)
		t = "RSA " + t
	default:
		err = fmt.Errorf("%q: %w", alg, ErrUnsupported)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    t,
		Headers: map[string]string{},
		Bytes:   der,
	}
	return pem.Encode(w, blk)
}

const ShowUsageTemplate = `
usage: {{.}} [<name>]
Print algorithm of the named private key file or stdin.`

func ShowUsageData(ctx context.Context) any {
	return strings.Join(ctxparm.Strings.In(ctx), " ")
}

func Show(ctx context.Context, args ...string) error {
	if *complete.Help {
		return complete.Last(args, "*.pem")
	}
	if *usage.Help {
		return usage.Error(ShowUsageTemplate[1:], ShowUsageData(ctx))
	}
	r := ctxparm.Reader.In(ctx)
	if len(args) > 0 && args[0] != "-" {
		if f, err := os.Open(args[0]); err != nil {
			return err
		} else {
			defer f.Close()
			r = f
		}
	}
	var priv Private
	if _, err := priv.ReadFrom(r); err != nil {
		return err
	}
	fmt.Fprintln(ctxparm.Writer.In(ctx), priv)
	return nil
}

func (priv Private) Private() Privateer { return priv.Privateer }

func (priv *Private) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, egress.Mark(err)
	}
	n := int64(len(data))
	for {
		var rest []byte
		if priv.Block, rest = pem.Decode(data); priv.Block == nil {
			return n, egress.Mark(ErrInvalid)
		}
		if strings.HasSuffix(priv.Block.Type, "PRIVATE KEY") {
			break
		}
		data = rest
	}
	v, err := x509.ParsePKCS8PrivateKey(priv.Block.Bytes)
	if err != nil {
		return n, egress.Mark(err)
	}
	priv.Privateer = v.(Privateer)
	return n, nil
}

func (priv Private) SignatureAlgorithm() x509.SignatureAlgorithm {
	switch t := priv.Privateer.(type) {
	case *ecdsa.PrivateKey:
		switch t.Public().(*ecdsa.PublicKey).Curve {
		case elliptic.P224(), elliptic.P256():
			return x509.ECDSAWithSHA256
		case elliptic.P384():
			return x509.ECDSAWithSHA384
		case elliptic.P521():
			return x509.ECDSAWithSHA512
		}
	case ed25519.PrivateKey:
		return x509.PureEd25519
	case *rsa.PrivateKey:
		return x509.SHA256WithRSA
	}
	return x509.UnknownSignatureAlgorithm
}

func (priv Private) String() string {
	alg := priv.SignatureAlgorithm()
	if alg != x509.UnknownSignatureAlgorithm {
		return alg.String()
	}
	if k, ok := priv.Privateer.(*ecdh.PrivateKey); ok {
		curve := k.Curve()
		switch {
		case curve == ecdh.P256():
			return "P256"
		case curve == ecdh.P384():
			return "P384"
		case curve == ecdh.P521():
			return "P521"
		case curve == ecdh.X25519():
			return "X25519"
		}
	}
	return fmt.Sprintf("%T", priv.Privateer)
}
