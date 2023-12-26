// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
)

var Self = sync.OnceValue(func() *xcert.TLS {
	self := new(xcert.TLS)
	data, err := os.ReadFile(KeyFileName())
	if err != nil {
		return nil
	}
	if err = self.X509.UnmarshalText(data); err != nil {
		return nil
	}
	if data, err = os.ReadFile(CertFileName()); err != nil {
		return nil
	}
	if err = self.UnmarshalText(data); err != nil {
		return nil
	}
	return self
})

type selfie struct{}

var Selfie selfie

func (selfie) MarshalPEM() ([]byte, error) {
	self := Self()
	if self == nil {
		return nil, ErrNotFound
	}
	return self.MarshalPEM()
}

func (selfie) MarshalText() ([]byte, error) {
	self := Self()
	if self == nil {
		return nil, ErrNotFound
	}
	buf := new(bytes.Buffer)
	algs := self.Certificate.SupportedSignatureAlgorithms
	if n := len(algs); n > 0 {
		fmt.Fprintln(buf, "supported_signature_algoritums:")
		for _, alg := range algs {
			fmt.Fprintln(buf, "  -", alg)
		}
	}
	ctss := self.Certificate.SignedCertificateTimestamps
	if n := len(ctss); n > 0 {
		fmt.Fprintln(buf, "signed_certificate_timestamps:", n)
	}
	fmt.Fprint(buf, &self.X509)
	return buf.Bytes(), nil
}

func (v selfie) Format(w fmt.State, verb rune) {
	if buf, err := v.MarshalText(); err == nil {
		w.Write(buf)
	} else {
		fmt.Fprint(w, err)
	}
}
