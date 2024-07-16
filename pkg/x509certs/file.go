// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xpem"
)

type File struct {
	Path   string
	mutex  sync.RWMutex
	blocks []*pem.Block
	certs  []*x509.Certificate
	// FIXME does this need to be []int for updated certs?
	named map[string]int
}

func NewFile(path string) (f *File, err error) {
	f = &File{
		Path:  path,
		named: make(map[string]int),
	}
	r, err := os.Open(path)
	if err == nil {
		defer r.Close()
		_, err = f.ReadFrom(r)
	}
	return
}

func (f *File) ReadFrom(r io.Reader) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	f.blocks, _ = xpem.DecodeAll(data)
	f.certs, err = Parse(f.blocks...)
	if err == nil {
		if f.named == nil {
			f.named = make(map[string]int)
		}
		for i, cert := range f.certs {
			if cert == nil {
				continue
			}
			f.named[cert.Subject.CommonName] = i
		}
	}
	return int64(len(data)), err
}

func (f *File) Add(block *pem.Block, cert *x509.Certificate) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	cn := cert.Subject.CommonName
	if f.named == nil {
		f.named = make(map[string]int)
	} else if _, ok := f.named[cn]; ok {
		return fmt.Errorf("%s: already exists in %s", cn, f.Path)
	}
	for i, c := range f.certs {
		if c == nil {
			f.blocks[i] = block
			f.certs[i] = cert
			f.named[cn] = i
			return xpem.Create(f.Path, 0644, f.blocks...)
		}
	}
	f.blocks = append(f.blocks, block)
	f.certs = append(f.certs, cert)
	f.named[cn] = len(f.certs) - 1
	return xpem.Append(f.Path, 0644, block)
}

func (f *File) DERs() [][]byte {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	ders := make([][]byte, len(f.blocks))
	for i, block := range f.blocks {
		ders[i] = block.Bytes
	}
	return ders
}

func (f *File) Dump(w io.Writer) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return xpem.EncodeAll(w, f.blocks...)
}

// Thie returns <nil> if there are no certificates.
func (f *File) First() *x509.Certificate {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, c := range f.certs {
		if c != nil {
			return c
		}
	}
	return nil
}

func (f *File) Has(peer *x509.Certificate) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, cert := range f.certs {
		if cert != nil && cert.Equal(peer) {
			return true
		}
	}
	return false
}

func (f *File) Join(pool *x509.CertPool) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, cert := range f.certs {
		pool.AddCert(cert)
	}
}

func (f *File) Named(cn string) (*pem.Block, *x509.Certificate, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if i, ok := f.named[cn]; ok {
		return f.blocks[i], f.certs[i], nil
	}
	return nil, nil, fmt.Errorf("%s: not found", cn)
}

func (f *File) Remove(cn string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if i, ok := f.named[cn]; ok {
		f.blocks[i] = nil
		f.certs[i] = nil
		delete(f.named, cn)
		return xpem.Create(f.Path, 0644, f.blocks...)
	}
	return fmt.Errorf("%s: not found", cn)
}

func (f *File) Show(w io.Writer) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if t, err := Template(); err != nil {
		return err
	} else {
		return t.Execute(w, f.certs)
	}
}
