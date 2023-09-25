// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
)

var Self = sync.OnceValue(func() *keycert.TLS {
	self := new(keycert.TLS)
	err := self.LoadX509KeyPair(CertFileName(), KeyFileName())
	if err != nil {
		panic(err)
	}
	return self
})

type selfie struct{}

var Selfie selfie

func (selfie) MarshalPEM() ([]byte, error) {
	return Self().MarshalPEM()
}

func (selfie) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	self := Self()
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
