// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	ErrNilLeaf    = errors.New("nil leaf")
	ErrNoDNSNames = errors.New("no DNS names")
)

func Preload(c tls.Certificate, x *x509.Certificate) {
	cert.Preload(func(p *tls.Certificate) {
		*p = c
		p.Leaf = x
	})
}

var SKI = cache.New[string](func(p *string) error {
	*p = hex.EncodeToString(cert.Value().Leaf.SubjectKeyId)
	return nil
}).Value

func ValErr() (tls.Certificate, error) { return cert.ValErr() }

func Value() tls.Certificate { return cert.Value() }

var cert = cache.New[tls.Certificate](func(p *tls.Certificate) (err error) {
	if *p, err = keycert.ReadTLSCertificate(
		filename.Cert(),
		filename.PrivateKey(),
	); err == nil {
		if p.Leaf == nil {
			err = ErrNilLeaf
		} else if len(p.Leaf.DNSNames) == 0 {
			err = ErrNoDNSNames
		}
	}
	return
})
