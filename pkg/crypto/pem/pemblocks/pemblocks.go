// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pemblocks

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Append(path string, mode os.FileMode, blocks ...*pem.Block) error {
	if err := AssurePath(path); err != nil {
		return err
	}
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, mode)
	if err != nil {
		return err
	}
	defer w.Close()
	return Encode(w, blocks...)
}

func Create(path string, mode os.FileMode, blocks ...*pem.Block) error {
	if err := AssurePath(path); err != nil {
		return err
	}
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer w.Close()
	return Encode(w, blocks...)
}

func Decode(data []byte) (blocks []*pem.Block, r []byte) {
	var b *pem.Block
	for b, r = pem.Decode(data); b != nil; b, r = pem.Decode(r) {
		blocks = append(blocks, b)
	}
	return
}

// Encode all non-nil blocks to writer.
func Encode(w io.Writer, blocks ...*pem.Block) error {
	for _, block := range blocks {
		if block == nil {
			continue
		}
		if err := pem.Encode(w, block); err != nil {
			return err
		}
	}
	return nil
}

func AssurePath(path string) error {
	dir := filepath.Dir(path)
	fi, err := os.Stat(dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("%s: isn't a directory", dir)
	}
	return os.MkdirAll(dir, 0755)
}

func Parse(path string) (blocks []*pem.Block, err error) {
	data, err := os.ReadFile(path)
	if err == nil {
		blocks, _ = Decode(data)
	}
	return
}
