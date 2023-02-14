// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"bytes"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type File struct{ Name func() string }

var (
	SubscribersFile   = File{filename.Subscribers}
	SubscriptionsFile = File{filename.Subscriptions}
)

type Certs struct {
	FN string
	X  []*keycert.X509
}

type CachedCerts struct{ *cache.Cache[Certs] }

var Subscribers = CachedCerts{cache.New[Certs](SubscribersFile.Load)}
var Subscriptions = CachedCerts{cache.New[Certs](SubscriptionsFile.Load)}

func Match(nameOrSKI string) (match *keycert.X509, err error) {
	match, err = Self.Match(nameOrSKI)
	if err != nil {
		match, err = Subscriptions.Match(nameOrSKI)
		if err != nil {
			match, err = Subscribers.Match(nameOrSKI)
			if err != nil {
				err = fmt.Errorf("cert:%s: %w", nameOrSKI, err)
			}
		}
	}
	return
}

func (file File) Load(c *Certs) error {
	c.FN = file.Name()
	b, err := ioutil.ReadFile(c.FN)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	for {
		var blk *pem.Block
		if blk, b = pem.Decode(b); blk == nil {
			return nil
		}
		x := new(keycert.X509)
		if err = x.UnmarshalPEM(blk); err != nil {
			return err
		}
		c.X = append(c.X, x)
	}
	return nil
}

func (cc CachedCerts) Append(x *keycert.X509) error {
	return cc.Mutex(func(p *Certs) error {
		(*p).X = append((*p).X, x)
		return x.AppendFile(p.FN)
	})
}

func (cc CachedCerts) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	cc.Range(func(x *keycert.X509) bool {
		fmt.Fprint(buf, x)
		return true
	})
	return buf.Bytes(), nil
}

func (cc CachedCerts) Match(nameOrSKI string) (match *keycert.X509, err error) {
	if len(nameOrSKI) == 0 {
		err = errors.New("empty name or SKI")
		return
	}
	cc.Range(func(x *keycert.X509) bool {
		if nameOrSKI == x.Name() || nameOrSKI == x.SKI() {
			match = x
			return false
		}
		return true
	})
	if match == nil {
		err = fmt.Errorf("%q: not found in %s",
			nameOrSKI, cc.Value().FN)
	}
	return
}

func (cc CachedCerts) Names() (names []string) {
	cc.Range(func(x *keycert.X509) bool {
		names = append(names, x.Name())
		return true
	})
	return
}

func (cc CachedCerts) Range(f func(*keycert.X509) bool) {
	cc.Mutex(func(p *Certs) error {
		for _, x := range (*p).X {
			if !f(x) {
				break
			}
		}
		return nil
	})
}

func (cc CachedCerts) SKIs() (skis []string) {
	cc.Range(func(x *keycert.X509) bool {
		skis = append(skis, x.SKI())
		return true
	})
	return
}
