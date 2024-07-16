// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"strings"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/gcm"
)

// Simulate a secure message from one host to another through an exchange.
func TestBox(t *testing.T) {
	const (
		hello   = "hello"
		bonjour = "bonjour"
	)

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

	box := New()

	x, err := gcm.New()
	assert(err)
	h1, err := gcm.New()
	assert(err)
	h2, err := gcm.New()
	assert(err)
	h1h2, err := h1.Peer(h2.PublicKey.Local)
	assert(err)
	h2h1, err := h2.Peer(h1.PublicKey.Local)
	assert(err)
	h1x, err := h1.Peer(x.PublicKey.Local)
	assert(err)
	xh1, err := x.Peer(h1.PublicKey.Local)
	assert(err)
	h2x, err := h2.Peer(x.PublicKey.Local)
	assert(err)
	xh2, err := x.Peer(h2.PublicKey.Local)
	assert(err)

	box = box.Empty().Append(hello).From(1).To(2).Close(h1h2)
	box.Seal(h1x)
	if len(box) != len(hello)+Overhead {
		t.Fatal(ErrOverrun)
	}
	if from := box.FromWhom(); from != 1 {
		t.Fatalf("wrong from address: %dd", from)
	}
	if to := box.ToWhom(); to == 2 {
		t.Fatalf("unsealed to address: %d", to)
	}
	if got := string(box.Contents()); strings.Index(got, hello) >= 0 {
		t.Fatal("unsealed contents")
	}
	assert(box.Unseal(xh1))
	box.Seal(xh2)
	if from := box.FromWhom(); from != 1 {
		t.Fatalf("wrong from address: %d", from)
	}
	if to := box.ToWhom(); to == 2 {
		t.Fatalf("unsealed to address")
	}
	if got := string(box.Contents()); strings.Index(got, hello) >= 0 {
		t.Fatal("unsealed contents")
	}
	assert(box.Unseal(h2x))
	if to := box.ToWhom(); to != 2 {
		t.Fatalf("misaddressed: %d", to)
	}
	box, err = box.Open(h2h1)
	assert(err)
	if got := string(box.Contents()); got != hello {
		t.Errorf("%q", got)
	} else {
		t.Log(got)
	}
	box = box.Empty().Append(bonjour).From(2).To(1).Close(h2h1)
	box.Seal(h2x)
	if from := box.FromWhom(); from != 2 {
		t.Fatalf("wrong from address: %d", from)
	}
	if to := box.ToWhom(); to == 1 {
		t.Fatalf("unsealed to address")
	}
	if got := string(box.Contents()); strings.Index(got, bonjour) >= 0 {
		t.Fatal("unsealed contents")
	}
	if len(box) != len(bonjour)+Overhead {
		t.Fatal(ErrOverrun)
	}
	assert(box.Unseal(xh2))
	if to := box.ToWhom(); to != 1 {
		t.Errorf("misaddressed: %d", to)
	}
	box.Seal(xh1)
	assert(box.Unseal(h1x))
	box, err = box.Open(h1h2)
	assert(err)
	if got := string(box.Contents()); got != bonjour {
		t.Errorf("%q", got)
	} else {
		t.Log(got)
	}
}
