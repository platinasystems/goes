// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/x509"

	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type CAs struct{ cache *cache.Cache[*x509.CertPool] }

// ClientCAs includes self plus all Subscribers and Subscriptions.
var ClientCAs = CAs{cache.New[*x509.CertPool](func(p **x509.CertPool) error {
	*p = x509.NewCertPool()
	(*p).AddCert(Self.Leaf())
	Subscriptions.Range(func(x X509) bool {
		(*p).AddCert(x.Certificate)
		return true
	})
	Subscribers.Range(func(x X509) bool {
		(*p).AddCert(x.Certificate)
		return true
	})
	return nil
})}

// RootCAs includes self plus all Subscriptions.
var RootCAs = CAs{cache.New[*x509.CertPool](func(p **x509.CertPool) error {
	*p = x509.NewCertPool()
	(*p).AddCert(Self.Leaf())
	Subscriptions.Range(func(x X509) bool {
		(*p).AddCert(x.Certificate)
		return true
	})
	return nil
})}

func (cached CAs) Add(c *x509.Certificate) {
	cached.cache.Ref(func(p **x509.CertPool) error {
		(*p).AddCert(c)
		return nil
	})
}

func (cached CAs) Clone() (cas *x509.CertPool) {
	cached.cache.Ref(func(p **x509.CertPool) error {
		cas = (*p).Clone()
		return nil
	})
	return
}
