// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"testing"
)

func assert(b *testing.B, v any) {
	b.Helper()
	if err, ok := v.(error); ok {
		if err != nil {
			if errors.Is(err, context.Canceled) {
				b.SkipNow()
			} else {
				b.Fatal(err)
			}
		}
	} else if t, ok := v.(bool); ok {
		if !t {
			b.Fatal("not!")
		}
	}
}

func BenchmarkCipherX25519(b *testing.B) {
	h1Priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	assert(b, err)
	h2Priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	assert(b, err)

	h1h2SharedKey, err := h1Priv.ECDH(h2Priv.PublicKey())
	assert(b, err)
	h2h1SharedKey, err := h2Priv.ECDH(h1Priv.PublicKey())
	assert(b, err)

	benchmarkCipher(b, h1h2SharedKey, h2h1SharedKey)
}

func BenchmarkCipherMLKEM(b *testing.B) {
	h1Priv, err := mlkem.GenerateKey768()
	assert(b, err)
	h1Pub := h1Priv.EncapsulationKey()

	h2h1Encap, err := mlkem.NewEncapsulationKey768(h1Pub.Bytes())
	assert(b, err)
	h2h1SharedKey, invite := h2h1Encap.Encapsulate()

	h1h2SharedKey, err := h1Priv.Decapsulate(invite)
	assert(b, err)

	benchmarkCipher(b, h1h2SharedKey, h2h1SharedKey)
}

func benchmarkCipher(b *testing.B, h1h2SharedKey, h2h1SharedKey []byte) {
	h1h2label := MakeLabel(1, 2)
	// h2h1label := MakeLabel(2, 1)
	h1h2block, err := aes.NewCipher(h1h2SharedKey)
	assert(b, err)
	h2h1block, err := aes.NewCipher(h2h1SharedKey)
	assert(b, err)
	h1h2gcm, err := cipher.NewGCMWithRandomNonce(h1h2block)
	assert(b, err)
	h2h1gcm, err := cipher.NewGCMWithRandomNonce(h2h1block)
	assert(b, err)

	mtu := UDP6MTU - h1h2gcm.Overhead() - SizeofLabel

	buf := make([]byte, mtu, UDP6MTU)
	n, err := rand.Read(buf)
	assert(b, err)
	b.SetBytes(int64(n))

	want := sha512.Sum512(buf)

	for b.Loop() {
		buf = h1h2gcm.Seal(buf[:0], nil, buf, h1h2label)
		h1h2block.Encrypt(buf[:aes.BlockSize], buf[:aes.BlockSize])
		buf = append(buf, h1h2label...)

		if to, from := ScanLabel(buf); to != 2 {
			b.Fatal("to:", to, "!=", 2)
		} else if from != 1 {
			b.Fatal("from:", from, "!=", 1)
		}

		h2h1block.Decrypt(buf[:aes.BlockSize], buf[:aes.BlockSize])
		i := len(buf) - SizeofLabel
		buf, err = h2h1gcm.Open(buf[:0], nil, buf[:i], buf[i:])
		if err != nil {
			b.Fatal(err)
		}
	}

	got := sha512.Sum512(buf)
	assert(b, bytes.Compare(want[:], got[:]) == 0)
}
