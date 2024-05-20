// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

var DefaultSignatureFileName = sync.OnceValue(func() string {
	return filepath.Join(xdg.StateHome(), program.Base(),
		"tlsx", "sig.pem")
})

var OptionalSignatureFileName *string

var Signature struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

var InitSignature = sync.OnceValue(func() error {
	data, err := os.ReadFile(*OptionalSignatureFileName)
	if err != nil || len(data) == 0 {
		return err
	}
	var block *pem.Block
	for {
		var rest []byte
		if block, rest = pem.Decode(data); block == nil {
			return nil
		}
		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			break
		}
		data = rest
	}
	pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	ed25519pk, ok := pk.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("%w: %T", ErrKeyType, pk)
	}
	Signature.Private = ed25519pk
	Signature.Public = Signature.Private.Public().(ed25519.PublicKey)
	return nil
})

func GenerateSignature(ctx context.Context, args []string) error {
	fn := *OptionalSignatureFileName
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["filename"] = fn
		return goes.Usage(ctx, `
usage: {{branch .}}
Write a PEM encoded private ED25519 signature to {{filename}}.`)
	}
	random := rand.Reader
	var err error
	Signature.Public, Signature.Private, err = ed25519.GenerateKey(random)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(Signature.Private)
	if err != nil {
		return err
	}
	w, err := goes.ContextOutput(ctx, fn, 0600)
	if err != nil {
		return err
	}
	return pem.Encode(w, &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   der,
	})
}

func ShowSignature(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["filename"] = func() string {
			return *OptionalSignatureFileName
		}
		return goes.Usage(ctx, `
usage: {{branch .}}
Print public key of {{filename}}.`)
	}
	err := InitSignature()
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKIXPublicKey(Signature.Public)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    "PUBLIC KEY",
		Headers: map[string]string{},
		Bytes:   der,
	}
	w := goes.ContextStdout(ctx)
	return pem.Encode(w, blk)
}
