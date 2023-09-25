// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package box

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/log/style"
)

func TestBox(t *testing.T) {
	const (
		hello   = "hello"
		bonjour = "bonjour"
	)
	style.Test(t)
	defer style.Recovery(context.Canceled)

	h1, err := NewCipher()
	if err != nil {
		panic(err)
	}
	h2, err := NewCipher(h1.PublicKey.Local)
	if err != nil {
		panic(err)
	}
	if err = h1.ECDH(h2.PublicKey.Local); err != nil {
		panic(err)
	}

	bx := New()
	defer bx.Recycle()

	bx = bx.Empty().Append(hello).From(1).To(2).Close(h1).Seal(h1)
	if len(bx) != BeginContent+len(hello)+CipherOverhead {
		panic(ErrOverrun)
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from == 1 || to == 2 {
		panic(fmt.Errorf("unsealed address: %d, %d", from, to))
	}
	if got := string(bx.Contents()); strings.Index(got, hello) >= 0 {
		panic("unsealed contents")
	}
	if bx, err = bx.Unseal(h2); err != nil {
		panic(err)
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from != 1 || to != 2 {
		panic(fmt.Errorf("misaddressed: %d, %d", from, to))
	}
	if bx, err = bx.Open(h2); err != nil {
		panic(err)
	}
	if got := string(bx.Contents()); got != hello {
		style.Errorf("fail\t%q != %q", got, hello)
	} else {
		style.Notef("ok\t%q", got)
	}
	bx = bx.Empty().Append(bonjour).From(2).To(1).Close(h2).Seal(h2)
	if from, to := bx.FromWhom(), bx.ToWhom(); from == 2 || to == 1 {
		panic(fmt.Errorf("unsealed address: %d, %d", from, to))
	}
	if got := string(bx.Contents()); strings.Index(got, bonjour) >= 0 {
		panic("unsealed contents")
	}
	if len(bx) != BeginContent+len(bonjour)+CipherOverhead {
		panic(ErrOverrun)
	}
	if bx, err = bx.Unseal(h1); err != nil {
		panic(err)
	}
	if from, to := bx.FromWhom(), bx.ToWhom(); from != 2 || to != 1 {
		panic(fmt.Errorf("misaddressed: %d, %d", from, to))
	}
	if bx, err = bx.Open(h1); err != nil {
		panic(err)
	}
	if got := string(bx.Contents()); got != bonjour {
		style.Errorf("fail\t%q != %q", got, bonjour)
	} else {
		style.Notef("ok\t%q", got)
	}
}
