// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type File struct{ Name func() string }

var (
	SubscribersFile   = File{filename.Subscribers}
	SubscriptionsFile = File{filename.Subscriptions}
)

func Unsubscribed(ex string) error {
	return fmt.Errorf("%q: not found in %s", ex, SubscriptionsFile.Name())
}

type Certs struct {
	FN string
	X  []X509
}

type CachedCerts struct{ *cache.Cache[Certs] }

var Subscribers = CachedCerts{cache.New[Certs](SubscribersFile.Load)}
var Subscriptions = CachedCerts{cache.New[Certs](SubscriptionsFile.Load)}

func (cc CachedCerts) Add(h Headers, c *x509.Certificate) error {
	return cc.Ref(func(p *Certs) error {
		(*p).X = append((*p).X, X509{
			Headers:     h,
			Certificate: c,
			Name:        c.DNSNames[0],
			SKI:         hex.EncodeToString(c.SubjectKeyId),
		})
		return keycert.AppendX509CertificatesFile(p.FN, h, c)
	})
}

func (cc CachedCerts) Format(w fmt.State, verb rune) {
	cc.Range(func(x X509) bool {
		fmt.Fprint(w, x)
		return true
	})
}

func (cc CachedCerts) Lookup(ex string) (
	name, ski string,
	c *x509.Certificate,
) {
	cc.Range(func(x X509) bool {
		if ex == x.SKI {
			ski = ex
			c = x.Certificate
			name = x.DNSNames[0]
			return false
		}
		for _, s := range x.Certificate.DNSNames {
			if ex == s {
				ski = x.SKI
				name = ex
				c = x.Certificate
				return false
			}
		}
		return true
	})
	return
}

func (cc CachedCerts) Names() (names []string) {
	cc.Range(func(x X509) bool {
		names = append(names, x.Name)
		return true
	})
	return
}

func (cc CachedCerts) Range(f func(X509) bool) {
	cc.Ref(func(p *Certs) error {
		for _, x := range (*p).X {
			if !f(x) {
				break
			}
		}
		return nil
	})
}

func (cc CachedCerts) SKIs() (skis []string) {
	cc.Range(func(x X509) bool {
		skis = append(skis, x.SKI)
		return true
	})
	return
}

func (file File) Load(c *Certs) error {
	c.FN = file.Name()
	blocks, err := keycert.DecodeFile(c.FN)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return nil
	}
	c.X = make([]X509, len(blocks))
	cs, err := keycert.ParseX509Certificates(blocks)
	if err == nil {
		for i, block := range blocks {
			c.X[i].Headers = block.Headers
			c.X[i].Certificate = cs[i]
			c.X[i].SKI = hex.
				EncodeToString(cs[i].SubjectKeyId)
		}
	}
	return err
}
