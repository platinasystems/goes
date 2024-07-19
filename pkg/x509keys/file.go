// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509keys

import (
	"encoding/pem"
	"io"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xpem"
)

type File struct {
	Path   string
	mutex  sync.RWMutex
	blocks []*pem.Block
	keys   []Privater
}

func NewFile(path string) (f *File, err error) {
	f = &File{Path: path}
	if path == "-" {
		_, err = f.ReadFrom(os.Stdin)
	} else {
		err = f.ReadFile(path)
	}
	return
}

func (f *File) ReadFile(path string) error {
	r, err := os.Open(path)
	if err == nil {
		defer r.Close()
		_, err = f.ReadFrom(r)
	}
	return err
}

func (f *File) ReadFrom(r io.Reader) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	f.blocks, _ = xpem.DecodeAll(data)
	f.keys, err = Parse(f.blocks...)
	return int64(len(data)), err
}

func (f *File) Add(block *pem.Block, key Privater) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	for i, k := range f.keys {
		if k == nil {
			f.blocks[i] = block
			f.keys[i] = key
			return xpem.Create(f.Path, 0600, f.blocks...)
		}
	}
	f.blocks = append(f.blocks, block)
	f.keys = append(f.keys, key)
	return xpem.Append(f.Path, 0600, block)
}

// Thie returns <nil> if there are no keys.
func (f *File) First() Privater {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	for _, key := range f.keys {
		if key != nil {
			return key
		}
	}
	return nil
}

func (f *File) Show(w io.Writer) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if t, err := Template(); err != nil {
		return err
	} else {
		return t.Execute(w, f.keys)
	}
}
