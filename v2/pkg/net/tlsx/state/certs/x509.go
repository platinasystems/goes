// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

type Headers = map[string]string

type X509 struct {
	Headers
	*x509.Certificate
	Name, SKI string
}

func NewX509(h Headers, c *x509.Certificate) *X509 {
	x := new(X509)
	x.Set(h, c)
	return x
}

func (x *X509) Set(h Headers, c *x509.Certificate) {
	x.Headers = h
	x.Certificate = c
	x.Name = c.DNSNames[0]
	x.SKI = hex.EncodeToString(c.SubjectKeyId)
}

// Format as yaml like sequence to writer.
func (x X509) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w, "- name:", x.Name)
	fmt.Fprintln(w, "  subject_key_id:", x.SKI)
	fmt.Fprintln(w, "  serial_number:", x.Certificate.SerialNumber)
	fmt.Fprintln(w, "  not_before:", x.Certificate.NotBefore)
	fmt.Fprintln(w, "  not_after:", x.Certificate.NotAfter)
	fmt.Fprintln(w, "  subject:", x.Certificate.Subject)
	if len(x.EmailAddresses) > 0 {
		fmt.Fprintln(w, "email_addresses:")
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
}
