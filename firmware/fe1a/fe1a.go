// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

var Eucode, Fucode struct {
	Version, Crc uint16
	Data         []byte
}

func Load() error {
	err := load("fe1a-e.ucode", &Eucode)
	if err == nil {
		err = load("fe1a-f.ucode", &Fucode)
	}
	return err
}

func load(fn string, ucode *struct {
	Version, Crc uint16
	Data         []byte
}) error {
	var f *os.File
	var err error
	for _, dir := range []string{
		".",
		"firmware/fe1a",
		"/usr/share/goes",
	} {
		f, err = os.Open(filepath.Join(dir, fn))
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("%s: not found", fn)
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(ucode)
}
