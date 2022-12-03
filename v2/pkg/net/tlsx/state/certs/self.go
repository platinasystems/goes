// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	ErrNilLeaf    = errors.New("nil leaf")
	ErrNoDNSNames = errors.New("no DNS names")
)

type TLS struct {
	tls.Certificate
	Name, SKI string
}

type CachedTLS struct{ *cache.Cache[TLS] }

var Self = CachedTLS{cache.New[TLS](func(p *TLS) (err error) {
	if p.Certificate, err = keycert.ReadTLSCertificate(
		filename.Cert(),
		filename.PrivateKey(),
	); err == nil {
		if p.Certificate.Leaf == nil {
			err = ErrNilLeaf
		} else if len(p.Certificate.Leaf.DNSNames) == 0 {
			err = ErrNoDNSNames
		}
	}
	p.Name = p.Certificate.Leaf.DNSNames[0]
	p.SKI = hex.EncodeToString(p.Certificate.Leaf.SubjectKeyId)
	return
})}

func (t CachedTLS) Format(w fmt.State, verb rune) {
	t.Ref(func(p *TLS) error {
		algs := p.Certificate.SupportedSignatureAlgorithms
		if n := len(algs); n > 0 {
			fmt.Fprintln(w, "supported_signature_algoritums:")
			for _, alg := range algs {
				fmt.Fprintln(w, "  -", alg)
			}
		}
		ctss := p.Certificate.SignedCertificateTimestamps
		if n := len(ctss); n > 0 {
			fmt.Fprintln(w, "signed_certificate_timestamps:", n)
		}
		fmt.Fprint(w, X509{
			Headers:     make(Headers),
			Certificate: p.Certificate.Leaf,
			Name:        p.Name,
			SKI:         p.SKI,
		})
		return nil
	})
}

func (t CachedTLS) DNSNames() (l []string) {
	t.Ref(func(p *TLS) error {
		l = p.Certificate.Leaf.DNSNames
		return nil
	})
	return
}

func (t CachedTLS) Leaf() (x *x509.Certificate) {
	t.Ref(func(p *TLS) error {
		x = p.Certificate.Leaf
		return nil
	})
	return
}

func (t CachedTLS) Name() (s string) {
	t.Ref(func(p *TLS) error {
		s = p.Name
		return nil
	})
	return
}

func (t CachedTLS) SKI() (s string) {
	t.Ref(func(p *TLS) error {
		s = p.SKI
		return nil
	})
	return
}

func (t CachedTLS) TLS() (c tls.Certificate) {
	t.Ref(func(p *TLS) error {
		c = p.Certificate
		return nil
	})
	return
}
