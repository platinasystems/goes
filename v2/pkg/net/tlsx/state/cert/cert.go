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

var ErrNilLeaf = errors.New("nil leaf")

var SKI = cache.New[string](func(p *string) error {
	c, err := cert.ValErr()
	if err == nil {
		if c.Leaf != nil {
			*p = hex.EncodeToString(c.Leaf.SubjectKeyId)
		} else {
			err = ErrNilLeaf
		}
	}
	return err
})

var cert = cache.New[tls.Certificate](func(
	p *tls.Certificate,
) error {
	cfn, err := filename.Cert.ValErr()
	if err != nil {
		return err
	}
	kfn, err := filename.PrivateKey.ValErr()
	if err != nil {
		return err
	}
	*p, err = keycert.ReadTLSCertificate(cfn, kfn)
	return err
})

func Preload(c tls.Certificate, x *x509.Certificate) {
	cert.Preload(func(p *tls.Certificate) {
		*p = c
		p.Leaf = x
	})
}

func ValErr() (tls.Certificate, error) { return cert.ValErr() }
