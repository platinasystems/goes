// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
)

type CertsFile struct{ Name func() string }

var (
	SubscribersFile   = CertsFile{SubscribersFileName}
	SubscriptionsFile = CertsFile{SubscriptionsFileName}
)

func (cf CertsFile) Load() *Certs {
	certs := &Certs{FN: cf.Name()}
	b, err := ioutil.ReadFile(certs.FN)
	if err != nil {
		if os.IsNotExist(err) {
			return certs
		}
		panic(err)
	}
	for {
		var blk *pem.Block
		if blk, b = pem.Decode(b); blk == nil {
			return nil
		}
		x := new(keycert.X509)
		if err = x.UnmarshalPEM(blk); err != nil {
			panic(err)
		}
		certs.X = append(certs.X, x)
	}
	return certs
}

type Certs struct {
	sync.RWMutex
	FN string
	X  []*keycert.X509
}

var Subscribers = sync.OnceValue(SubscribersFile.Load)
var Subscriptions = sync.OnceValue(SubscriptionsFile.Load)

func Match(nameOrSKI string) (match *keycert.X509, err error) {
	if self := Self(); nameOrSKI == self.Name() || nameOrSKI == self.SKI() {
		return &self.X509, nil
	}
	if err != nil {
		match, err = Subscriptions().Match(nameOrSKI)
		if err != nil {
			match, err = Subscribers().Match(nameOrSKI)
			if err != nil {
				err = fmt.Errorf("cert:%s: %w", nameOrSKI, err)
			}
		}
	}
	return
}

func (c *Certs) Append(x *keycert.X509) {
	c.Lock()
	defer c.Unlock()
	c.X = append(c.X, x)
	//FIXME writeback FN
}

func (c *Certs) Index(nameOrSKI string) (int, error) {
	if len(nameOrSKI) == 0 {
		return -1, ErrUnnamed
	}
	c.RLock()
	defer c.RUnlock()
	for i, x := range c.X {
		if nameOrSKI == x.Name() || nameOrSKI == x.SKI() {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%q: %w in %s", nameOrSKI, ErrNotFound, c.FN)
}

func (c *Certs) MarshalText() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	buf := new(bytes.Buffer)
	for _, x := range c.X {
		fmt.Fprint(buf, x)
	}
	return buf.Bytes(), nil
}

func (c *Certs) Match(nameOrSKI string) (*keycert.X509, error) {
	if len(nameOrSKI) == 0 {
		return nil, ErrUnnamed
	}
	c.RLock()
	defer c.RUnlock()
	for _, x := range c.X {
		if nameOrSKI == x.Name() || nameOrSKI == x.SKI() {
			return x, nil
		}
	}
	return nil, fmt.Errorf("%q: %w in %s", nameOrSKI, ErrNotFound, c.FN)
}

func (c *Certs) Names() []string {
	c.RLock()
	defer c.RUnlock()
	names := make([]string, len(c.X))
	for i, x := range c.X {
		names[i] = x.Name()
	}
	return names
}

func (c *Certs) Range(f func(*keycert.X509) bool) {
	c.RLock()
	defer c.RUnlock()
	for _, x := range c.X {
		if !f(x) {
			break
		}
	}
}

func (c *Certs) SKIs() []string {
	c.RLock()
	defer c.RUnlock()
	skis := make([]string, len(c.X))
	for i, x := range c.X {
		skis[i] = x.SKI()
	}
	return skis
}
