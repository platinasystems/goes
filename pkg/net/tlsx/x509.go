// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

type X509 struct {
	*pem.Block
	*x509.Certificate
	Next *X509
}

func NewX509(data []byte) (*X509, error) {
	x := new(X509)
	err := x.UnmarshalText(data)
	return x, err
}

func (x *X509) Append(a *X509) {
	var p *X509
	for p = x; p.Next != nil; p = p.Next {
	}
	p.Next = a
}

// Format as yaml like sequence to writer.
func (x *X509) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w, "- name:", x.DNS0())
	fmt.Fprintln(w, "  subject_key_id:", x.SKI())
	fmt.Fprintln(w, "  serial_number:", x.SerialNumber)
	fmt.Fprintln(w, "  not_before:", x.NotBefore)
	fmt.Fprintln(w, "  not_after:", x.NotAfter)
	fmt.Fprintln(w, "  subject:", x.Subject)
	if len(x.EmailAddresses) > 0 {
		fmt.Fprintln(w, "  email_addresses:")
		for _, email := range x.EmailAddresses {
			fmt.Fprintln(w, "    -", email)
		}
	}
	if len(x.DNSNames) > 0 {
		fmt.Fprintln(w, "  dns_names:")
		for _, dns := range x.DNSNames {
			fmt.Fprintln(w, "    -", dns)
		}
	}
	if len(x.IPAddresses) > 0 {
		fmt.Fprintln(w, "ip_addresses:")
		for _, ip := range x.IPAddresses {
			fmt.Fprintln(w, "    -", ip)
		}
	}
	if len(x.URIs) > 0 {
		fmt.Fprintln(w, "uris:")
		for _, uri := range x.URIs {
			fmt.Fprintln(w, "    -", uri)
		}
	}
	fmt.Fprintln(w, "  public_key_algorithm:", x.PublicKeyAlgorithm)
	fmt.Fprintln(w, "  signature_algorithm:", x.SignatureAlgorithm)
	fmt.Fprintln(w, "  key_usage:")
	if x.KeyUsage == 0 {
		fmt.Fprintln(w, "    - none")
	}
	if (x.KeyUsage & x509.KeyUsageDigitalSignature) != 0 {
		fmt.Fprintln(w, "    -", "digital_signature")
	}
	if (x.KeyUsage & x509.KeyUsageContentCommitment) != 0 {
		fmt.Fprintln(w, "    -", "content_commitment")
	}
	if (x.KeyUsage & x509.KeyUsageKeyEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "key_encipherment")
	}
	if (x.KeyUsage & x509.KeyUsageDataEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "data_encipherment")
	}
	if (x.KeyUsage & x509.KeyUsageKeyAgreement) != 0 {
		fmt.Fprintln(w, "    -", "key_agreement")
	}
	if (x.KeyUsage & x509.KeyUsageCertSign) != 0 {
		fmt.Fprintln(w, "    -", "cert_sign")
	}
	if (x.KeyUsage & x509.KeyUsageCRLSign) != 0 {
		fmt.Fprintln(w, "    -", "CRL_sign")
	}
	if (x.KeyUsage & x509.KeyUsageEncipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "encipher_only")
	}
	if (x.KeyUsage & x509.KeyUsageDecipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "decipher_only")
	}
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(x.Certificate)
	fmt.Fprint(w, "  signature: ")
	if _, err := x.Verify(opts); err == nil {
		fmt.Fprintln(w, "ok")
	} else {
		fmt.Fprintln(w, err)
	}
	for k, v := range x.Block.Headers {
		fmt.Fprint(w, "  ", k, ": ", v, "\n")
	}
	fmt.Fprintln(w, "  version:", x.Version)
	data := pem.EncodeToMemory(&pem.Block{
		Type:    "CERTIFICATE",
		Headers: x.Block.Headers,
		Bytes:   x.Raw,
	})
	const (
		indent = 4
		nls    = "\n                "
	)
	fmt.Fprint(w, "  pem: |")
	w.Write([]byte(nls[:1+indent]))
	buf := bytes.Replace(data, []byte(nls[:1]), []byte(nls[:1+indent]), -1)
	w.Write(buf[:len(buf)-indent])
}

func (x *X509) DNS0() string {
	name := "anonymous"
	if x.Certificate != nil && len(x.DNSNames) > 0 {
		name = x.DNSNames[0]
	}
	return name
}

func (x *X509) Index(nameOrSki string) int {
	for i, p := 0, x; p != nil; i, p = i+1, p.Next {
		if p.IsMatch(nameOrSki) {
			return i
		}
	}
	return -1
}

func (x *X509) IsMatch(nameOrSKI string) bool {
	if x.Certificate == nil {
		return false
	}
	if nameOrSKI == x.SKI() {
		return true
	}
	for _, name := range x.DNSNames {
		if nameOrSKI == name {
			return true
		}
	}
	return false
}

func (x *X509) MarshalPEM() ([]byte, error) {
	if x == nil || x.Block == nil {
		return nil, ErrNilLeaf
	}
	return pem.EncodeToMemory(x.Block), nil
}

func (x *X509) Names() (names []string) {
	for p := x; p != nil; p = p.Next {
		for _, name := range x.DNSNames {
			names = append(names, name)
		}
	}
	return
}

func (x *X509) Range(f func(*X509) bool) {
	for p := x; p != nil; p = p.Next {
		if !f(p) {
			break
		}
	}
}

func (x *X509) ReadFrom(r io.Reader) (n int64, err error) {
	data, err := io.ReadAll(r)
	if err == nil {
		err = x.UnmarshalText(data)
	}
	return int64(len(data)), err
}

func (x *X509) SKI() (s string) {
	if x.Certificate != nil {
		s = hex.EncodeToString(x.Certificate.SubjectKeyId)
	}
	return
}

func (x *X509) SKIs() (skis []string) {
	for p := x; p != nil; p = p.Next {
		skis = append(skis, p.SKI())
	}
	return
}

func (x *X509) UnmarshalText(data []byte) error {
	if x.Block, data = pem.Decode(data); x.Block == nil {
		return egress.Mark(ErrInvalid)
	}
	if x.Block.Type != "CERTIFICATE" {
		x.Block = nil
		if err := x.UnmarshalText(data); err != nil {
			return egress.Mark(err)
		} else if x.Block == nil {
			return nil
		}
	}
	if err := x.UnmarshalDER(x.Block.Bytes); err != nil {
		return egress.Mark(err)
	}
	if len(data) > 0 {
		if next, err := NewX509(data); err == nil {
			if next.Block != nil {
				x.Next = next
			}
		}
	}
	return nil
}

func (x *X509) UnmarshalDER(der []byte) (err error) {
	x.Certificate, err = x509.ParseCertificate(der)
	return err
}

func (x *X509) WhoIs(nameOrSKI string) *X509 {
	for p := x; p != nil; p = p.Next {
		if p.IsMatch(nameOrSKI) {
			return p
		}
	}
	return nil
}
