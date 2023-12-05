// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

const (
	CipherOverhead = 16
	NonceSize      = 12
)

var ErrIncomplete = errors.New("incomplete")

func NewNonce() []byte { return make([]byte, NonceSize, NonceSize) }

// Returns a new nonce sized buffer that contains the sum of IV (incremental
// value) with the most significant 8 bytes of nonce.
func NoncePlusIV(nonce, iv []byte) []byte {
	sum := NewNonce()
	copy(sum, nonce)
	be := ByteOrder.Uint64(nonce)
	be += ByteOrder.Uint64(iv)
	ByteOrder.PutUint64(sum, be)
	return sum
}

// Cipher contains the Galois Counter Mode (GCM) parameters for authenticated
// encryption with a shared Elliptic Curve Diffie-Hellman (ECDH) keys.
type Cipher struct {
	PublicKey struct{ Local, Remote *pem.Block }
	nonce     struct{ local, remote, shared []byte }
	priv      *ecdh.PrivateKey
	gcm       cipher.AEAD
}

// Without opts, the local key and nonce are a randomly generated x25519 and a
// [NonceSize]byte respectively.  These may be preempted with *ecdh.PrivateKey
// and []byte  arguments.  Also, a *pem.Block argument will be decoded for the
// peer public key and nonce.
func NewCipher(opts ...any) (*Cipher, error) {
	var err error
	var remote *pem.Block
	c := new(Cipher)
	for _, opt := range opts {
		switch t := opt.(type) {
		case *ecdh.PrivateKey:
			c.priv = t
		case []byte:
			c.nonce.local = t
		case *pem.Block:
			remote = t
		}
	}
	if c.priv == nil {
		c.priv, err = ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			return nil, egress.Mark(err)
		}
	}
	if c.nonce.local == nil {
		c.nonce.local = NewNonce()
		if n, err := rand.Read(c.nonce.local); err != nil {
			return nil, egress.Mark(err)
		} else if n != NonceSize {
			return nil, egress.Mark(ErrIncomplete)
		}
	}
	der, err := x509.MarshalPKIXPublicKey(c.priv.PublicKey())
	if err != nil {
		return nil, egress.Mark(err)
	}
	c.PublicKey.Local = &pem.Block{
		Type: "PUBLIC KEY",
		Headers: map[string]string{
			"nonce": hex.EncodeToString(c.nonce.local),
		},
		Bytes: der,
	}
	c.nonce.remote = NewNonce()
	if remote != nil {
		if err := c.ECDH(remote); err != nil {
			return nil, egress.Mark(err)
		}
	}
	return c, nil
}

// Clone local key and nonce to peer with new remote.
func (c *Cipher) Clone(remote *pem.Block) (*Cipher, error) {
	clone := new(Cipher)
	clone.PublicKey.Local = c.PublicKey.Local
	clone.nonce.local = c.nonce.local
	clone.priv = c.priv
	return clone, clone.ECDH(remote)
}

// Calculate key and nonce shared with remote.
func (c *Cipher) ECDH(remote *pem.Block) error {
	var err error
	if s := remote.Headers["nonce"]; len(s) == 0 {
		return egress.Mark(ErrIncomplete)
	} else if c.nonce.remote, err = hex.DecodeString(s); err != nil {
		return egress.Mark(err)
	} else if len(c.nonce.remote) != NonceSize {
		return egress.Mark(ErrIncomplete)
	}
	rk, err := x509.ParsePKIXPublicKey(remote.Bytes)
	if err != nil {
		return egress.Mark(err)
	}
	shared, err := c.priv.ECDH(rk.(*ecdh.PublicKey))
	if err != nil {
		return egress.Mark(err)
	}
	cipherblock, err := aes.NewCipher(shared)
	if err != nil {
		return egress.Mark(err)
	}
	if c.gcm, err = cipher.NewGCM(cipherblock); err != nil {
		return egress.Mark(err)
	}
	if c.nonce.shared == nil || cap(c.nonce.shared) < NonceSize {
		c.nonce.shared = NewNonce()
	} else if len(c.nonce.shared) != NonceSize {
		c.nonce.shared = c.nonce.shared[:NonceSize]
	}
	copy(c.nonce.shared, c.nonce.local)
	for i, b := range c.nonce.remote {
		c.nonce.shared[i] ^= b
	}
	c.PublicKey.Remote = remote
	return nil
}
