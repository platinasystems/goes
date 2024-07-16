// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509keys

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

type Privater interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// This returns a list of private keys parsed from the corresponding PKCS#8 DER
// encoded ASN.1 structures of block type matching "PRIVATE KEY".  Any other
// block type will have a corresponding nil key.
func Parse(blocks ...*pem.Block) ([]Privater, error) {
	keys := make([]Privater, len(blocks))
	for i, block := range blocks {
		if !strings.HasSuffix(block.Type, "PRIVATE KEY") {
			continue
		}
		b := block.Bytes
		if k, err := x509.ParsePKCS8PrivateKey(b); err != nil {
			return keys, fmt.Errorf("block[%d]: %w", i, err)
		} else {
			keys[i] = k.(Privater)
		}
	}
	return keys, nil
}
