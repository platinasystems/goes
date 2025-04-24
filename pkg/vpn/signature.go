// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

type Privater interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// ShowSignature prints algorithm.
func ShowSignature(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Print algorithm.

{{flags .}}`)

	err := defineAndParseFlags(args)
	if err != nil {
		return err
	}

	sig, err := NewSignatures(ConfigDirFile(SigFlag))
	if err == nil {
		err = sig.Show(os.Stdout)
	}
	return err
}

// The embedding type or method must mutex Signatures.
type Signatures struct {
	fn  string
	BKs []BK
}

type BK struct {
	Block *pem.Block
	Key   Privater
}

func NewSignatures(fn string) (sigs *Signatures, err error) {
	var r io.Reader
	sigs = &Signatures{fn: fn}
	if fn == "-" {
		r = os.Stdin
	} else if rc, err := os.Open(fn); err != nil {
		return sigs, err
	} else {
		defer rc.Close()
		r = rc
	}
	data, err := io.ReadAll(r)
	if err == nil {
		err = sigs.parse(data)
	}
	return sigs, err
}

// Decode PEM then parse private keys from PKCS#8 DER.
func (sigs *Signatures) parse(data []byte) error {
	var blks []*pem.Block
	for blk, r := pem.Decode(data); blk != nil; blk, r = pem.Decode(r) {
		if strings.HasSuffix(blk.Type, "PRIVATE KEY") {
			blks = append(blks, blk)
		}
	}
	for i, blk := range blks {
		k, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
		if err != nil {
			return xerrors.Label(err, "block", fmt.Sprint(i))
		} else {
			sigs.BKs = append(sigs.BKs, BK{blk, k.(Privater)})
		}
	}
	return nil
}

func (sigs *Signatures) Add(blk *pem.Block, k Privater) error {
	sigs.BKs = append(sigs.BKs, BK{blk, k})
	if sigs.fn == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	w, err := os.OpenFile(sigs.fn, oAppend, 0600)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Thie returns <nil> if there are no keys.
func (sigs *Signatures) First() Privater {
	for _, bk := range sigs.BKs {
		if bk.Key != nil {
			return bk.Key
		}
	}
	return nil
}

func (sigs *Signatures) Show(w io.Writer) error {
	if len(sigs.BKs) == 0 {
		fmt.Fprintln(w, "# none")
		return nil
	}
	for _, bk := range sigs.BKs {
		fmt.Fprintf(w, "- %T\n", bk.Key)
	}
	return nil
}

func (sigs *Signatures) String() string {
	return sigs.fn
}
