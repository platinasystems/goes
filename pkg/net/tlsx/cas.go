// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"crypto/x509"
	"sync"
)

// If Restricted is true, the Self certificate is the only permitted Client.
var Restricted bool

type CAs struct {
	sync.RWMutex
	pool *x509.CertPool
}

var ClientCAs, RootCAs CAs

var ClientCAsDependencies = []func() error{
	InitSelf,
	InitSubscribers,
	InitSubscriptions,
}

var RootCAsDependencies = []func() error{
	InitSelf,
	InitSubscriptions,
}

var InitClientCAs = sync.OnceValue(func() error {
	for _, f := range ClientCAsDependencies {
		if err := f(); err != nil {
			return err
		}
	}
	ClientCAs.pool = x509.NewCertPool()
	ClientCAs.pool.AddCert(Self.X509.Certificate)
	if Restricted {
		return nil
	}
	Subscriptions.Range(func(x *X509) bool {
		ClientCAs.pool.AddCert(x.Certificate)
		return true
	})
	Subscribers.Range(func(x *X509) bool {
		ClientCAs.pool.AddCert(x.Certificate)
		return true
	})
	return nil
})

var InitRootCAs = sync.OnceValue(func() error {
	for _, f := range RootCAsDependencies {
		if err := f(); err != nil {
			return err
		}
	}
	RootCAs.pool = x509.NewCertPool()
	RootCAs.pool.AddCert(Self.X509.Certificate)
	Subscriptions.Range(func(x *X509) bool {
		RootCAs.pool.AddCert(x.Certificate)
		return true
	})
	return nil
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
