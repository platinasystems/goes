// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
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

type TLS struct{ *cache.Cache[keycert.TLS] }

var Self = TLS{cache.New[keycert.TLS](func(p *keycert.TLS) error {
	return p.LoadX509KeyPair(filename.Cert(), filename.PrivateKey())
})}

func (t TLS) Add(cas *x509.CertPool) {
	t.Mutex(func(p *keycert.TLS) error {
		if c := p.X509.Certificate; c != nil {
			cas.AddCert(c)
		}
		return nil
	})
}

func (t TLS) Format(w fmt.State, verb rune) {
	if buf, err := t.MarshalText(); err == nil {
		w.Write(buf)
	} else {
		fmt.Fprint(w, err)
	}
}

func (t TLS) MarshalPEM() (data []byte, err error) {
	t.Mutex(func(p *keycert.TLS) error {
		data, err = p.MarshalPEM()
		return err
	})
	return
}

func (t TLS) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.Mutex(func(p *keycert.TLS) error {
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
		fmt.Fprint(buf, &p.X509)
		return nil
	})
	return buf.Bytes(), err
}

func (t TLS) Match(nameOrSKI string) (match *keycert.X509, err error) {
	if len(nameOrSKI) == 0 {
		err = errors.New("empty name or SKI")
		return
	}
	err = t.Mutex(func(p *keycert.TLS) error {
		if nameOrSKI == p.X509.Name() || nameOrSKI == p.X509.SKI() {
			match = &p.X509
			return nil
		}
		return fmt.Errorf("%q: neither name nor SKI of %s",
			nameOrSKI, filename.Cert())
	})
	return
}

func (t TLS) Name() (s string) {
	t.Mutex(func(p *keycert.TLS) error {
		s = p.Name()
		return nil
	})
	return
}

func (t TLS) SKI() (s string) {
	t.Mutex(func(p *keycert.TLS) error {
		s = p.SKI()
		return nil
	})
	return
}

func (t TLS) TLS() (c tls.Certificate) {
	t.Mutex(func(p *keycert.TLS) error {
		c = p.Certificate
		return nil
	})
	return
}
