// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
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
	X509
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
		} else {
			p.X509.Set(Headers{}, p.Certificate.Leaf)
		}
	}
	return
})}

func (t CachedTLS) Add(cas *x509.CertPool) {
	t.Mutex(func(p *TLS) error {
		if c := p.X509.Certificate; c != nil {
			cas.AddCert(c)
		}
		return nil
	})
}

func (t CachedTLS) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	t.Mutex(func(p *TLS) error {
		algs := p.Certificate.SupportedSignatureAlgorithms
		if n := len(algs); n > 0 {
			fmt.Fprintln(buf, "supported_signature_algoritums:")
			for _, alg := range algs {
				fmt.Fprintln(buf, "  -", alg)
			}
		}
		ctss := p.Certificate.SignedCertificateTimestamps
		if n := len(ctss); n > 0 {
			fmt.Fprintln(buf, "signed_certificate_timestamps:", n)
		}
		fmt.Fprint(buf, p.X509)
		return nil
	})
	return buf.Bytes(), nil
}

func (t CachedTLS) Match(nameOrSKI string) (match *X509, err error) {
	if len(nameOrSKI) == 0 {
		err = errors.New("empty name or SKI")
		return
	}
	err = t.Mutex(func(p *TLS) error {
		if nameOrSKI == p.X509.Name || nameOrSKI == p.X509.SKI {
			match = &p.X509
			return nil
		}
		return fmt.Errorf("%q: neither name nor SKI of %s",
			nameOrSKI, filename.Cert())
	})
	return
}

func (t CachedTLS) Name() (s string) {
	t.Mutex(func(p *TLS) error {
		s = p.Name
		return nil
	})
	return
}

func (t CachedTLS) SKI() (s string) {
	t.Mutex(func(p *TLS) error {
		s = p.SKI
		return nil
	})
	return
}

func (t CachedTLS) TLS() (c tls.Certificate) {
	t.Mutex(func(p *TLS) error {
		c = p.Certificate
		return nil
	})
	return
}
