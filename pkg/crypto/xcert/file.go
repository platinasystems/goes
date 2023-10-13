// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcert

import (
	"encoding/pem"
	"io/fs"
	"os"
	"sync"
)

type NameX509File func() string

func (f NameX509File) Load() *X509File {
	x := &X509File{
		name: f(),
	}
	data, err := os.ReadFile(x.name)
	if err != nil {
		if os.IsNotExist(err) {
			return x
		}
		panic(err)
	}
	x.head = new(X509)
	if err = x.head.UnmarshalText(data); err != nil {
		panic(err)
	} else if x.head.Block == nil {
		panic(ErrInvalid)
	}
	return x
}

type X509File struct {
	mutex sync.RWMutex
	name  string
	head  *X509
}

func (x *X509File) Append(sub *X509) error {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	if x.head == nil {
		x.head = sub
	} else {
		x.head.Append(sub)
	}
	const fflags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	fmode := fs.FileMode(0600)
	if fi, err := os.Stat(x.name); err == nil {
		fmode = fi.Mode()
	}
	f, err := os.OpenFile(x.name, fflags, fmode)
	if err != nil {
		return err
	}
	defer f.Close()
	x.head.Range(func(xx *X509) bool {
		return pem.Encode(f, xx.Block) == nil
	})
	return nil
}

func (x *X509File) Index(nameOrSKI string) int {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	return x.head.Index(nameOrSKI)
}

func (x *X509File) Match(nameOrSKI string) *X509 {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	return x.head.WhoIs(nameOrSKI)
}

func (x *X509File) Names() []string {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	return x.head.Names()
}

func (x *X509File) Range(f func(*X509) bool) {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	x.head.Range(f)
}

func (x *X509File) SKIs() []string {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	return x.head.SKIs()
}
