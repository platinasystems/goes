// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"strings"
	"testing"
)

func TestBox(t *testing.T) {
	const (
		hello   = "hello"
		bonjour = "bonjour"
	)

	h1, err := NewCipher()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := NewCipher(h1.PublicKey.Local)
	if err != nil {
		t.Fatal(err)
	}
	if err = h1.ECDH(h2.PublicKey.Local); err != nil {
		t.Fatal(err)
	}

	bx := New()
	defer bx.Recycle()

	bx = bx.Empty().Append(hello).From(1).To(2).Close(h1).Seal(h1)
	if len(bx) != BeginContent+len(hello)+CipherOverhead {
		t.Fatal(ErrOverrun)
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from == 1 || to == 2 {
		t.Fatalf("unsealed address: %d, %d", from, to)
	}
	if got := string(bx.Contents()); strings.Index(got, hello) >= 0 {
		t.Fatal("unsealed contents")
	}
	if bx, err = bx.Unseal(h2); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
		return
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from != 1 || to != 2 {
		t.Fatalf("misaddressed: %d, %d", from, to)
	}
	if bx, err = bx.Open(h2); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
		return
	}
	if got := string(bx.Contents()); got != hello {
		t.Errorf("fail\t%q != %q", got, hello)
	} else {
		t.Logf("ok\t%q", got)
	}
	bx = bx.Empty().Append(bonjour).From(2).To(1).Close(h2).Seal(h2)
	if from, to := bx.FromWhom(), bx.ToWhom(); from == 2 || to == 1 {
		t.Fatalf("unsealed address: %d, %d", from, to)
	}
	if got := string(bx.Contents()); strings.Index(got, bonjour) >= 0 {
		t.Fatal("unsealed contents")
	}
	if len(bx) != BeginContent+len(bonjour)+CipherOverhead {
		t.Fatal(ErrOverrun)
	}
	if bx, err = bx.Unseal(h1); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
		return
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from != 2 || to != 1 {
		t.Errorf("misaddressed: %d, %d", from, to)
	}
	if bx, err = bx.Open(h1); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
		return
	}
	if got := string(bx.Contents()); got != bonjour {
		t.Errorf("fail\t%q != %q", got, bonjour)
	} else {
		t.Logf("ok\t%q", got)
	}
}
