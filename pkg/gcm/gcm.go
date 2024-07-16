// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"

	"github.com/platinasystems/goes/v2/pkg/nonce"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const Overhead = 16

// Cipher contains the Galois Counter Mode (GCM) parameters for authenticated
// encryption with a shared Elliptic Curve Diffie-Hellman (ECDH) keys.
type Cipher struct {
	gcm   cipher.AEAD
	nonce []byte
}

// Assert peer.
func New(
	priv *ecdh.PrivateKey,
	pub *ecdh.PublicKey,
	local, remote []byte,
) (*Cipher, error) {
	shared, err := xerrors.MarkResult(priv.ECDH(pub))
	if err != nil {
		return nil, err
	}
	cb, err := xerrors.MarkResult(aes.NewCipher(shared))
	if err != nil {
		return nil, err
	}
	gcm, err := xerrors.MarkResult(cipher.NewGCM(cb))
	if err != nil {
		return nil, err
	}
	return &Cipher{gcm, nonce.Xor(local, remote)}, nil
}

func (c *Cipher) Open(dst, iv, ciphertext, data []byte) (
	[]byte, error,
) {
	return c.gcm.Open(dst, nonce.Sum(c.nonce, iv), ciphertext, data)
}

func (c *Cipher) Seal(dst, iv, plaintext, data []byte) []byte {
	return c.gcm.Seal(dst, nonce.Sum(c.nonce, iv), plaintext, data)
}

/*LEGACY
type Cipher struct {
	PublicKey struct{ Local, Remote *pem.Block }
	nonce     struct{ local, remote, shared Nonce }
	priv      *ecdh.PrivateKey
	gcm       cipher.AEAD
}

// Generate local x25519 key and nonce.
func New() (*Cipher, error) {
	var nonce Nonce
	if n, err := rand.Read(nonce[:]); err != nil {
		return nil, xerrors.Mark(err)
	} else if n != NonceSize {
		return nil, xerrors.Mark(ErrIncomplete)
	}
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	der, err := x509.MarshalPKIXPublicKey(priv.PublicKey())
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	blk := &pem.Block{
		Type: "PUBLIC KEY",
		Headers: map[string]string{
			"nonce": hex.EncodeToString(nonce[:]),
		},
		Bytes: der,
	}
	c := new(Cipher)
	copy(c.nonce.local[:], nonce[:])
	c.priv = priv
	c.PublicKey.Local = blk
	return c, nil
}

// Clone Cipher and peer with the remote PEM block containing its public key
// and hex nonce header.
func (c *Cipher) Peer(remote *pem.Block) (*Cipher, error) {
	clone := new(Cipher)
	clone.PublicKey.Local = c.PublicKey.Local
	clone.PublicKey.Remote = remote
	copy(clone.nonce.local[:], c.nonce.local[:])
	clone.priv = c.priv
	if s := remote.Headers["nonce"]; len(s) == 0 {
		return nil, xerrors.Mark(ErrIncomplete)
	} else if nonce, err := hex.DecodeString(s); err != nil {
		return nil, xerrors.Mark(err)
	} else if len(nonce) != NonceSize {
		return nil, xerrors.Mark(ErrIncomplete)
	} else {
		copy(clone.nonce.remote[:], nonce)
		copy(clone.nonce.shared[:], clone.nonce.local[:])
		for i, b := range nonce {
			clone.nonce.shared[i] ^= b
		}
	}
	rk, err := x509.ParsePKIXPublicKey(remote.Bytes)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	shared, err := clone.priv.ECDH(rk.(*ecdh.PublicKey))
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	if cb, err := aes.NewCipher(shared); err != nil {
		return nil, xerrors.Mark(err)
	} else if gcm, err := cipher.NewGCM(cb); err != nil {
		return nil, xerrors.Mark(err)
	} else {
		clone.gcm = gcm
	}
	return clone, nil
}
*/
