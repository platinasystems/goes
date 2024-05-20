// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
)

func labelHeader(blk *pem.Block) (label.Label, error) {
	return label.Parse(blk.Headers["label"])
}

func viaHeader(blk *pem.Block) (label.Label, error) {
	return label.Parse(blk.Headers["via"])
}

func addressHeader(blk *pem.Block) (netip.Addr, error) {
	return netip.ParseAddr(blk.Headers["address"])
}

func nonceHeader(blk *pem.Block) ([]byte, error) {
	return hex.DecodeString(blk.Headers["nonce"])
}

func prefixHeader(blk *pem.Block) (netip.Prefix, error) {
	return netip.ParsePrefix(blk.Headers["prefix"])
}

func serviceHeader(blk *pem.Block) (netip.AddrPort, error) {
	return netip.ParseAddrPort(blk.Headers["service"])
}

func ecdhPublicKey(blk *pem.Block) (*ecdh.PublicKey, error) {
	if blk.Type != "PUBLIC KEY" {
		return nil, ErrInvalidBlockType
	}
	anyk, err := x509.ParsePKIXPublicKey(blk.Bytes)
	if err != nil {
		return nil, err
	}
	pubkey, ok := anyk.(*ecdh.PublicKey)
	if !ok {
		return nil, ErrInvalidKey
	}
	return pubkey, nil
}
