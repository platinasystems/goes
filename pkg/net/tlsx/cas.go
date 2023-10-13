// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"crypto/x509"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
)

// If Restricted is true, the Self certificate is the only permitted Client.
var Restricted bool

type CAs struct {
	sync.RWMutex
	pool *x509.CertPool
}

// ClientCAs includes self plus all Subscribers and Subscriptions.
var ClientCAs = sync.OnceValue(func() *CAs {
	cas := &CAs{pool: x509.NewCertPool()}
	cas.pool.AddCert(Self().X509.Certificate)
	if Restricted {
		return cas
	}
	Subscriptions().Range(func(x *xcert.X509) bool {
		cas.pool.AddCert(x.Certificate)
		return true
	})
	Subscribers().Range(func(x *xcert.X509) bool {
		cas.pool.AddCert(x.Certificate)
		return true
	})
	return cas
})

// RootCAs includes self plus all Subscriptions.
var RootCAs = sync.OnceValue(func() *CAs {
	cas := &CAs{pool: x509.NewCertPool()}
	cas.pool.AddCert(Self().X509.Certificate)
	Subscriptions().Range(func(x *xcert.X509) bool {
		cas.pool.AddCert(x.Certificate)
		return true
	})
	return cas
})

func (cas *CAs) Add(c *x509.Certificate) {
	cas.Lock()
	defer cas.Unlock()
	cas.pool.AddCert(c)
}

func (cas *CAs) Clone() *x509.CertPool {
	cas.RLock()
	defer cas.RUnlock()
	return cas.pool.Clone()
}
