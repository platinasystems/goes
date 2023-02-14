// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keycert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	ErrNilLeaf    = errors.New("nil leaf")
	ErrNoDNSNames = errors.New("no DNS names")
)

type TLS struct {
	tls.Certificate
	X509
}

func (t *TLS) LoadX509KeyPair(certFn, keyFn string) (err error) {
	t.Certificate, err = tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		return
	}
	if t.Certificate.Leaf == nil {
		t.Certificate.Leaf, err = x509.
			ParseCertificate(t.Certificate.Certificate[0])
		if err != nil {
			return
		}
	}
	if t.Certificate.Leaf == nil {
		err = ErrNilLeaf
	} else if len(t.Certificate.Leaf.DNSNames) == 0 {
		err = ErrNoDNSNames
	} else {
		t.X509.Headers = Headers{}
		t.X509.Certificate = t.Certificate.Leaf
	}
	return
}

func NewTLSCertificate(cert, key *pem.Block) (tls.Certificate, error) {
	certEnc := pem.EncodeToMemory(cert)
	keyEnc := pem.EncodeToMemory(key)
	tlscert, err := tls.X509KeyPair(certEnc, keyEnc)
	if err == nil && tlscert.Leaf == nil {
		tlscert.Leaf, err =
			x509.ParseCertificate(tlscert.Certificate[0])
	}
	return tlscert, err
}
