// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"io"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

type TLS struct {
	X509
	tls.Certificate
}

func (x *TLS) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = x.X509.ReadFrom(r)
	if err == nil {
		err = x.validate()
	}
	return
}

func (x *TLS) UnmarshalText(data []byte) error {
	err := x.X509.UnmarshalText(data)
	if err == nil {
		err = x.validate()
	}
	return err
}

func (x *TLS) UnmarshalDER(der []byte) error {
	err := x.X509.UnmarshalDER(der)
	if err == nil {
		err = x.validate()
	}
	return err
}

func (x *TLS) validate() error {
	if x.Certificate.PrivateKey == nil {
		return egress.Marked(ErrPrivateKey)
	}
	if len(x.X509.Certificate.DNSNames) == 0 {
		return egress.Marked(ErrNoDNSNames)
	}
	if x.X509.Block.Type != "CERTIFICATE" {
		return egress.Marked(ErrInvalid)
	}
	switch pub := x.X509.Certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := x.Certificate.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return egress.Marked(ErrKeyType)
		}
		if pub.N.Cmp(priv.N) != 0 {
			return egress.Marked(ErrKeyParm)
		}
	case *ecdsa.PublicKey:
		priv, ok := x.Certificate.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return egress.Marked(ErrKeyType)
		}
		if pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0 {
			return egress.Marked(ErrKeyParm)
		}
	case ed25519.PublicKey:
		priv, ok := x.Certificate.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			return egress.Marked(ErrKeyType)
		}
		if !bytes.Equal(priv.Public().(ed25519.PublicKey), pub) {
			return egress.Marked(ErrKeyParm)
		}
	default:
		return egress.Marked(ErrInvalid)
	}
	x.Certificate.Certificate =
		append(x.Certificate.Certificate, x.X509.Block.Bytes)
	for next := x.X509.Next; next != nil; next = x.Next {
		x.Certificate.Certificate =
			append(x.Certificate.Certificate, next.Block.Bytes)
	}
	return nil
}
