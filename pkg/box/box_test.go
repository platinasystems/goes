// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/gcm"
	"github.com/platinasystems/goes/v2/pkg/nonce"
)

// Simulate a secure message from one host to another through an exchange.
func BenchmarkBox(b *testing.B) {
	const xId, h1Id, h2Id Id = 0, 1, 2

	assert := func(err error) {
		b.Helper()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				b.SkipNow()
			} else {
				b.Fatal(err)
			}
		}
	}

	xKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	assert(err)
	h1Key, err := ecdh.X25519().GenerateKey(rand.Reader)
	assert(err)
	h2Key, err := ecdh.X25519().GenerateKey(rand.Reader)
	assert(err)

	xNonce := make([]byte, nonce.Size)
	h1Nonce := make([]byte, nonce.Size)
	h2Nonce := make([]byte, nonce.Size)

	_, err = rand.Read(xNonce)
	assert(err)
	_, err = rand.Read(h1Nonce)
	assert(err)
	_, err = rand.Read(h2Nonce)
	assert(err)

	xh1, err := gcm.New(xKey, h1Key.PublicKey(), xNonce, h1Nonce)
	assert(err)
	xh2, err := gcm.New(xKey, h2Key.PublicKey(), xNonce, h2Nonce)
	assert(err)

	h1x, err := gcm.New(h1Key, xKey.PublicKey(), h1Nonce, xNonce)
	assert(err)
	h2x, err := gcm.New(h2Key, xKey.PublicKey(), h2Nonce, xNonce)
	assert(err)

	h1h2, err := gcm.New(h1Key, h2Key.PublicKey(), h1Nonce, h2Nonce)
	assert(err)

	h2h1, err := gcm.New(h2Key, h1Key.PublicKey(), h2Nonce, h1Nonce)
	assert(err)

	bx, err := NewReadContents(rand.Reader)
	assert(err)
	defer bx.Return()
	expectSum := sha512.Sum512(bx.Contents)

	bx.From(h1Id)
	bx.To(h2Id)
	bx.Via(xId)

	for b.Loop() {
		// Simulate packaging from h1 to h2
		bx.CloseWith(h1h2)
		bx.SealWith(h1x)
		if from := bx.FromWhom(); from != h1Id {
			b.Fatal(from, "!=", h1Id)
		}
		if via := bx.ViaWhom(); via != xId {
			b.Fatal(via, "!=", xId)
		}

		// Simulate x1 reading then resealing label
		assert(bx.UnsealWith(xh1))
		bx.SealWith(xh2)

		// Simulate receipt by h2
		assert(bx.UnsealWith(h2x))
		if to := bx.ToWhom(); to != h2Id {
			b.Fatal(to, "!=", h2Id)
		}
		assert(bx.OpenWith(h2h1))
	}

	gotSum := sha512.Sum512(bx.Contents)
	if bytes.Compare(expectSum[:], gotSum[:]) != 0 {
		b.Fatal("mismatch")
	}

	bps := float64(ContentMTU*b.N) / b.Elapsed().Seconds()
	var scale string
	if bps > 1e9 {
		bps /= 1e9
		scale = "G"
	} else if bps > 1e6 {
		bps /= 1e6
		scale = "M"
	} else if bps > 1e3 {
		bps /= 1e3
		scale = "K"
	}
	fmt.Printf("%.1f %sB/s\n", bps, scale)
}
