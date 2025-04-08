// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"io"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/gcm"
	"github.com/platinasystems/goes/v2/pkg/nonce"
)

// Simulate a secure message from one host to another through an exchange.
func TestBox(t *testing.T) {
	const (
		hello   = "hello"
		bonjour = "bonjour"
	)
	const xId, h1Id, h2Id Id = 0, 1, 2

	assert := func(err error) {
		t.Helper()
		if err != nil {
			if err != context.Canceled {
				t.Fatal(err)
			} else {
				t.SkipNow()
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

	bx := New()
	_, err = io.WriteString(bx, hello)
	assert(err)
	bx.From(h1Id)
	bx.To(h2Id)
	bx.Via(xId)
	bx.CloseWith(h1h2)
	bx.SealWith(h1x)

	if from := bx.FromWhom(); from != h1Id {
		t.Fatal(from, "!=", h1Id)
	}
	if via := bx.ViaWhom(); via != xId {
		t.Fatal(via, "!=", xId)
	}

	assert(bx.UnsealWith(xh1))
	bx.SealWith(xh2)
	assert(bx.UnsealWith(h2x))
	if to := bx.ToWhom(); to != h2Id {
		t.Fatal(to, "!=", h2Id)
	}

	assert(bx.OpenWith(h2h1))
	if contents := string(bx.Contents); contents != hello {
		t.Fatal(contents, "!=", hello)
	}

	bx.Empty()
	_, err = io.WriteString(bx, bonjour)
	assert(err)
	bx.From(h2Id)
	bx.To(h1Id)
	bx.Via(xId)
	bx.CloseWith(h2h1)
	bx.SealWith(h2x)

	if from := bx.FromWhom(); from != h2Id {
		t.Fatal(from, "!=", h2Id)
	}
	if via := bx.ViaWhom(); via != xId {
		t.Fatal(via, "!=", xId)
	}

	assert(bx.UnsealWith(xh2))
	bx.SealWith(xh1)
	assert(bx.UnsealWith(h1x))
	if to := bx.ToWhom(); to != h1Id {
		t.Fatal(to, "!=", h2Id)
	}

	assert(bx.OpenWith(h1h2))
	if contents := string(bx.Contents); contents != bonjour {
		t.Fatal(contents, "!=", bonjour)
	}
}
