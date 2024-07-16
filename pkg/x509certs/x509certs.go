// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// This returns a list of certificates parsed from the corresponding DER
// encoded ASN.1 structures of matching "CERTIFICATE" block type.  Any other
// blocks type will have a corresponding nil certificate.
func Parse(blocks ...*pem.Block) ([]*x509.Certificate, error) {
	certs := make([]*x509.Certificate, len(blocks))
	for i, block := range blocks {
		if !strings.HasSuffix(block.Type, "CERTIFICATE") {
			continue
		}
		b := block.Bytes
		if c, err := x509.ParseCertificate(b); err != nil {
			return certs, fmt.Errorf("block[%d]: %w", i, err)
		} else {
			certs[i] = c
		}
	}
	return certs, nil
}
