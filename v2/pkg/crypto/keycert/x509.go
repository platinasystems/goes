// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package keycert

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

type Headers = map[string]string

type X509 struct {
	Headers
	*x509.Certificate
}

func (x *X509) AppendFile(name string) error {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = pem.Encode(f, &pem.Block{
		Headers: x.Headers,
		Bytes:   x.Certificate.Raw,
	}); err != nil {
		return err
	}
	return nil
}

// Format as yaml like sequence to writer.
func (x *X509) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w, "- name:", x.Name())
	fmt.Fprintln(w, "  subject_key_id:", x.SKI())
	fmt.Fprintln(w, "  serial_number:", x.Certificate.SerialNumber)
	fmt.Fprintln(w, "  not_before:", x.Certificate.NotBefore)
	fmt.Fprintln(w, "  not_after:", x.Certificate.NotAfter)
	fmt.Fprintln(w, "  subject:", x.Certificate.Subject)
	if len(x.EmailAddresses) > 0 {
		fmt.Fprintln(w, "  email_addresses:")
		for _, email := range x.Certificate.EmailAddresses {
			fmt.Fprintln(w, "    -", email)
		}
	}
	if len(x.Certificate.DNSNames) > 0 {
		fmt.Fprintln(w, "  dns_names:")
		for _, dns := range x.Certificate.DNSNames {
			fmt.Fprintln(w, "    -", dns)
		}
	}
	if len(x.Certificate.IPAddresses) > 0 {
		fmt.Fprintln(w, "ip_addresses:")
		for _, ip := range x.Certificate.IPAddresses {
			fmt.Fprintln(w, "    -", ip)
		}
	}
	if len(x.Certificate.URIs) > 0 {
		fmt.Fprintln(w, "uris:")
		for _, uri := range x.Certificate.URIs {
			fmt.Fprintln(w, "    -", uri)
		}
	}
	fmt.Fprintln(w, "  public_key_algorithm:",
		x.Certificate.PublicKeyAlgorithm)
	fmt.Fprintln(w, "  signature_algorithm:",
		x.Certificate.SignatureAlgorithm)
	fmt.Fprintln(w, "  key_usage:")
	if x.Certificate.KeyUsage == 0 {
		fmt.Fprintln(w, "    - none")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageDigitalSignature) != 0 {
		fmt.Fprintln(w, "    -", "digital_signature")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageContentCommitment) != 0 {
		fmt.Fprintln(w, "    -", "content_commitment")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageKeyEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "key_encipherment")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageDataEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "data_encipherment")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageKeyAgreement) != 0 {
		fmt.Fprintln(w, "    -", "key_agreement")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageCertSign) != 0 {
		fmt.Fprintln(w, "    -", "cert_sign")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageCRLSign) != 0 {
		fmt.Fprintln(w, "    -", "CRL_sign")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageEncipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "encipher_only")
	}
	if (x.Certificate.KeyUsage & x509.KeyUsageDecipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "decipher_only")
	}
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(x.Certificate)
	fmt.Fprint(w, "  signature: ")
	if _, err := x.Certificate.Verify(opts); err == nil {
		fmt.Fprintln(w, "ok")
	} else {
		fmt.Fprintln(w, err)
	}
	for k, v := range x.Headers {
		fmt.Fprint(w, "  ", k, ": ", v, "\n")
	}
	fmt.Fprintln(w, "  version:", x.Certificate.Version)
	if data, err := x.MarshalPEM(); err == nil {
		fmt.Fprint(w, "  pem: |")
		DataBlock(w, data, 4)
	}
}

func (x *X509) MarshalPEM() (data []byte, err error) {
	data = pem.EncodeToMemory(&pem.Block{
		Type:    "CERTIFICATE",
		Headers: x.Headers,
		Bytes:   x.Certificate.Raw,
	})
	return
}

func (x *X509) Name() (s string) {
	if x.Certificate != nil && len(x.Certificate.DNSNames) > 0 {
		s = x.Certificate.DNSNames[0]
	}
	return
}

func (x *X509) SKI() (s string) {
	if x.Certificate != nil {
		s = hex.EncodeToString(x.Certificate.SubjectKeyId)
	}
	return
}

func (x *X509) UnmarshalPEM(blk *pem.Block) error {
	c, err := x509.ParseCertificate(blk.Bytes)
	if err == nil {
		x.Headers = blk.Headers
		x.Certificate = c
	}
	return err
}

func DataBlock(w io.Writer, data []byte, i int) {
	const nls = "\n                "
	w.Write([]byte(nls[:1+i]))
	buf := bytes.Replace(data, []byte(nls[:1]), []byte(nls[:1+i]), -1)
	w.Write(buf[:len(buf)-i])
}

func NewX509Certificate(k PrivateKey, temp *x509.Certificate) (
	cert *x509.Certificate, block *pem.Block, err error,
) {
	random := rand.Reader
	hn, err := host.Name.ValErr()
	if err = egress.Marked(err); err != nil {
		return
	}
	if temp.SerialNumber == nil {
		max := big.NewInt(math.MaxInt64)
		temp.SerialNumber, err = rand.Int(random, max)
		if err = egress.Marked(err); err != nil {
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
	if err = egress.Marked(err); err != nil {
		return
	}
	cert, err = x509.ParseCertificate(der)
	if err = egress.Marked(err); err != nil {
		return
	}
	block = &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	return
}
