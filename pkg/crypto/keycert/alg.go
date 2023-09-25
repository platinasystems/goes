// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keycert

import (
	"crypto/x509"
	"fmt"
)

var Algs = []string{
	"ed25519",
	"ecdsa256",
	"ecdsa384",
	"ecdsa512",
	"rsa256",
	"rsa384",
	"rsa512",
}

type Alg x509.SignatureAlgorithm

const (
	PureEd25519     = Alg(x509.PureEd25519)
	ECDSAWithSHA256 = Alg(x509.ECDSAWithSHA256)
	ECDSAWithSHA384 = Alg(x509.ECDSAWithSHA384)
	ECDSAWithSHA512 = Alg(x509.ECDSAWithSHA512)
	SHA256WithRSA   = Alg(x509.SHA256WithRSA)
	SHA384WithRSA   = Alg(x509.SHA384WithRSA)
	SHA512WithRSA   = Alg(x509.SHA512WithRSA)
)

func (alg *Alg) String() string {
	return map[Alg]string{
		PureEd25519:     "ed25519",
		ECDSAWithSHA256: "ecdsa256",
		ECDSAWithSHA384: "ecdsa384",
		ECDSAWithSHA512: "ecdsa512",
		SHA256WithRSA:   "rsa256",
		SHA384WithRSA:   "rsa384",
		SHA512WithRSA:   "rsa512",
	}[*alg]
}

func (alg *Alg) Set(s string) error {
	v, ok := map[string]Alg{
		"ed25519":  PureEd25519,
		"ecdsa256": ECDSAWithSHA256,
		"ecdsa384": ECDSAWithSHA384,
		"ecdsa512": ECDSAWithSHA512,
		"rsa256":   SHA256WithRSA,
		"rsa384":   SHA384WithRSA,
		"rsa512":   SHA512WithRSA,
	}[s]
	if !ok {
		return fmt.Errorf("%q: invalid", s)
	}
	*alg = Alg(v)
	return nil
}

func (alg *Alg) Value() x509.SignatureAlgorithm {
	return x509.SignatureAlgorithm(*alg)
}
