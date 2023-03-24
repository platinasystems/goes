// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keycert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const KeyFileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

type PrivateKey interface {
	Public() crypto.PublicKey
}

var ErrUnsupportedSignatureAlgorithm = errors.New(
	"unsupported signature algorithm")

// type defined kind:
//
//	elliptic.Curve - ECDSA
//	int - RSA with given number of bits
//	nil - ED25519
func NewPrivateKey(alg x509.SignatureAlgorithm) (
	k PrivateKey, block *pem.Block, err error,
) {
	random := rand.Reader
	switch alg {
	case x509.PureEd25519:
		_, k, err = ed25519.GenerateKey(random)
	case x509.ECDSAWithSHA256:
		k, err = ecdsa.GenerateKey(elliptic.P256(), random)
	case x509.ECDSAWithSHA384:
		k, err = ecdsa.GenerateKey(elliptic.P384(), random)
	case x509.ECDSAWithSHA512:
		k, err = ecdsa.GenerateKey(elliptic.P521(), random)
	case x509.SHA256WithRSA:
		k, err = rsa.GenerateKey(random, 256)
	case x509.SHA384WithRSA:
		k, err = rsa.GenerateKey(random, 384)
	case x509.SHA512WithRSA:
		k, err = rsa.GenerateKey(random, 512)
	default:
		err = ErrUnsupportedSignatureAlgorithm
	}
	if err != nil {
		return
	}
	pkcs, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return
	}
	block = &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   pkcs,
	}
	return
}

func ReadPrivateKeyFile(name string) (PrivateKey, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(data)
	if blk == nil {
		return nil, fmt.Errorf("%s: isn't PEM", name)
	}
	pkcs, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	if k, ok := pkcs.(PrivateKey); ok {
		return k, nil
	}
	return nil, fmt.Errorf("%s: unsupported block type: %s", name, blk.Type)
}
