// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sig

import (
	"crypto/ed25519"

	"github.com/platinasystems/goes/v2/pkg/xlog"
)

type pureEd25519 struct {
	ed25519.PublicKey
}

func (v pureEd25519) verify(msg []byte) bool {
	n := len(msg)
	if n < ed25519.SignatureSize {
		xlog.Errata.Println(n, "<", ed25519.SignatureSize)
		return false
	}
	i := len(msg) - ed25519.SignatureSize
	return ed25519.Verify(v.PublicKey, msg[:i], msg[i:])
}

func generateEd25519() (any, error) {
	_, priv, err := ed25519.GenerateKey(nil)
	return priv, err
}

func signEd25519(data []byte) []byte {
	return ed25519.Sign(Priv.(ed25519.PrivateKey), data)
}
