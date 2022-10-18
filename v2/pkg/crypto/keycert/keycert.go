// Copyright © 2022 Platina Systems, Inc. All rights reserved.
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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/os/host"
)

const (
	BlockTypeECDSA   = "ECDSA PRIVATE KEY"
	BlockTypeED25519 = "ED25519 PRIVATE KEY"
	BlockTypeRSA     = "RSA PRIVATE KEY"
	KeyFileFlags     = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

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
	kbt := "PRIVATE KEY"
	switch alg {
	case x509.PureEd25519:
		_, k, err = ed25519.GenerateKey(random)
		kbt = BlockTypeED25519
	case x509.ECDSAWithSHA256:
		k, err = ecdsa.GenerateKey(elliptic.P256(), random)
		kbt = BlockTypeECDSA
	case x509.ECDSAWithSHA384:
		k, err = ecdsa.GenerateKey(elliptic.P384(), random)
		kbt = BlockTypeECDSA
	case x509.ECDSAWithSHA512:
		k, err = ecdsa.GenerateKey(elliptic.P521(), random)
		kbt = BlockTypeECDSA
	case x509.SHA256WithRSA:
		k, err = rsa.GenerateKey(random, 256)
		kbt = BlockTypeRSA
	case x509.SHA384WithRSA:
		k, err = rsa.GenerateKey(random, 384)
		kbt = BlockTypeRSA
	case x509.SHA512WithRSA:
		k, err = rsa.GenerateKey(random, 512)
		kbt = BlockTypeRSA
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
		Type:    kbt,
		Headers: map[string]string{},
		Bytes:   pkcs,
	}
	return
}

func NewX509Certificate(k PrivateKey, temp *x509.Certificate) (
	cert *x509.Certificate, block *pem.Block, err error,
) {
	random := rand.Reader
	hn, err := host.Name.ValErr()
	if err != nil {
		return
	}
	if temp.SerialNumber == nil {
		max := big.NewInt(math.MaxInt64)
		temp.SerialNumber, err = rand.Int(random, max)
		if err != nil {
			return
		}
	}
	if len(temp.DNSNames) == 0 {
		temp.DNSNames = []string{hn}
	}
	if len(temp.Subject.CommonName) == 0 {
		temp.Subject.CommonName = temp.DNSNames[0]
	}
	if temp.NotBefore.IsZero() {
		temp.NotBefore = time.Now()
	}
	if temp.NotAfter.IsZero() || temp.NotAfter.Before(temp.NotBefore) {
		temp.NotAfter = temp.NotBefore.Add(10 * 365 * 24 * time.Hour)
	}
	der, err := x509.CreateCertificate(random, temp, temp, k.Public(), k)
	if err != nil {
		return
	}
	if cert, err = x509.ParseCertificate(der); err != nil {
		return
	}
	block = &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	return
}

func NewTLSCertificate(cert, key *pem.Block) (tls.Certificate, error) {
	certEnc := pem.EncodeToMemory(cert)
	keyEnc := pem.EncodeToMemory(key)
	tlscert, err := tls.X509KeyPair(certEnc, keyEnc)
	if err == nil && tlscert.Leaf == nil {
		tlscert.Leaf, err =
			x509.ParseCertificate(tlscert.Certificate[0])
	}
	return tlscert, err
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

func ReadTLSCertificate(certFn, keyFn string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err == nil && cert.Leaf == nil {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	return cert, err
}

func DecodeFile(name string) (blocks []*pem.Block, err error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return
	}
	for i := 0; len(b) > 0; i++ {
		var block *pem.Block
		if block, b = pem.Decode(b); block != nil {
			blocks = append(blocks, block)
		} else {
			break
		}
	}
	return
}

func ParseX509Certificates(blocks []*pem.Block) (
	certs []*x509.Certificate, err error,
) {
	certs = make([]*x509.Certificate, len(blocks))
	for i, block := range blocks {
		certs[i], err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			err = fmt.Errorf("[%d]: %w", i, err)
			break
		}
	}
	return
}

func AppendX509CertificatesFile(
	name string,
	headers map[string]string,
	x509c *x509.Certificate,
) error {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = pem.Encode(f, &pem.Block{
		Headers: headers,
		Bytes:   x509c.Raw,
	}); err != nil {
		return err
	}
	return nil
}
