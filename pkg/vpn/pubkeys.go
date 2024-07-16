// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type pubkey struct {
	raw   []byte
	block *pem.Block
	key   crypto.PublicKey
}

func (p *pubkey) UnmarshalBinary(data []byte) error {
	var err error
	p.raw = data
	p.block, _ = pem.Decode(data)
	if p.block == nil {
		return xerrors.Invalid("encoding")
	}
	if !strings.HasSuffix(p.block.Type, "PUBLIC KEY") {
		return fmt.Errorf(`type %q !=  "PUBLIC KEY"`, p.block.Type)
	}
	der := p.block.Bytes
	p.key, err = x509.ParsePKCS1PublicKey(der)
	if err != nil {
		p.key, err = x509.ParsePKIXPublicKey(der)
	}
	return err
}
