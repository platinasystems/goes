// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type Headers = map[string]string

type Certs struct {
	Cache *cache.Cache[Cache]
}

type Cache struct {
	FileName string
	Headers  []Headers
	Peers    []*x509.Certificate
	Pool     *x509.CertPool
}

type File struct {
	Name filename.Type
}

var (
	Clients   = Certs{cache.New[Cache](File{filename.Clients}.Load)}
	Exchanges = Certs{cache.New[Cache](File{filename.Exchanges}.Load)}
)

// Format yaml like sequence entry to writer.
func Fsequent(w io.Writer, headers Headers, cert *x509.Certificate) {
	ski := hex.EncodeToString(cert.SubjectKeyId)
	fmt.Fprintln(w, "- subject_key_id:", ski)
	fmt.Fprintln(w, "  serial_number:", cert.SerialNumber)
	fmt.Fprintln(w, "  not_before:", cert.NotBefore)
	fmt.Fprintln(w, "  not_after:", cert.NotAfter)
	fmt.Fprintln(w, "  subject:", cert.Subject)
	if len(cert.EmailAddresses) > 0 {
		fmt.Fprintln(w, "email_addresses:")
		for _, email := range cert.EmailAddresses {
			fmt.Fprintln(w, "    -", email)
		}
	}
	if len(cert.DNSNames) > 0 {
		fmt.Fprintln(w, "  dns_names:")
		for _, dns := range cert.DNSNames {
			fmt.Fprintln(w, "    -", dns)
		}
	}
	if len(cert.IPAddresses) > 0 {
		fmt.Fprintln(w, "ip_addresses:")
		for _, ip := range cert.IPAddresses {
			fmt.Fprintln(w, "    -", ip)
		}
	}
	if len(cert.URIs) > 0 {
		fmt.Fprintln(w, "uris:")
		for _, uri := range cert.URIs {
			fmt.Fprintln(w, "    -", uri)
		}
	}
	fmt.Fprintln(w, "  public_key_algorithm:", cert.PublicKeyAlgorithm)
	fmt.Fprintln(w, "  signature_algorithm:", cert.SignatureAlgorithm)
	fmt.Fprintln(w, "  key_usage:")
	if cert.KeyUsage == 0 {
		fmt.Fprintln(w, "    - none")
	}
	if (cert.KeyUsage & x509.KeyUsageDigitalSignature) != 0 {
		fmt.Fprintln(w, "    -", "digital_signature")
	}
	if (cert.KeyUsage & x509.KeyUsageContentCommitment) != 0 {
		fmt.Fprintln(w, "    -", "content_commitment")
	}
	if (cert.KeyUsage & x509.KeyUsageKeyEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "key_encipherment")
	}
	if (cert.KeyUsage & x509.KeyUsageDataEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "data_encipherment")
	}
	if (cert.KeyUsage & x509.KeyUsageKeyAgreement) != 0 {
		fmt.Fprintln(w, "    -", "key_agreement")
	}
	if (cert.KeyUsage & x509.KeyUsageCertSign) != 0 {
		fmt.Fprintln(w, "    -", "cert_sign")
	}
	if (cert.KeyUsage & x509.KeyUsageCRLSign) != 0 {
		fmt.Fprintln(w, "    -", "CRL_sign")
	}
	if (cert.KeyUsage & x509.KeyUsageEncipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "encipher_only")
	}
	if (cert.KeyUsage & x509.KeyUsageDecipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "decipher_only")
	}
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(cert)
	fmt.Fprint(w, "  signature: ")
	if _, err := cert.Verify(opts); err == nil {
		fmt.Fprintln(w, "ok")
	} else {
		fmt.Fprintln(w, err)
	}
	for k, v := range headers {
		fmt.Fprint(w, "  ", k, ": ", v, "\n")
	}
	fmt.Fprintln(w, "  version:", cert.Version)
}

func (certs Certs) Add(
	headers Headers,
	peer *x509.Certificate,
) error {
	return certs.Cache.Ref(func(p *Cache) error {
		(*p).Headers = append((*p).Headers, headers)
		(*p).Peers = append((*p).Peers, peer)
		(*p).Pool.AddCert(peer)
		if len(p.FileName) == 0 {
			return nil
		}
		return keycert.AppendX509CertificatesFile(p.FileName,
			headers, peer)
	})
}

func (certs Certs) NamePort(nameOrSKI string) (name, port string) {
	if len(nameOrSKI) == 0 {
		return
	}
	certs.Range(func(
		headers Headers,
		cert *x509.Certificate,
	) bool {
		if cert == nil || len(cert.DNSNames) == 0 {
			return true
		}
		if nameOrSKI == hex.EncodeToString(cert.SubjectKeyId) {
			name, port = cert.DNSNames[0], headers["port"]
			return false
		}
		for _, s := range cert.DNSNames {
			if s == nameOrSKI {
				name, port = s, headers["port"]
				return false
			}
		}
		return true
	})
	return
}

func (certs Certs) FirstDNS() (s string) {
	certs.Cache.Ref(func(p *Cache) error {
		if len((*p).Peers) > 0 && (*p).Peers[0] != nil &&
			len((*p).Peers[0].DNSNames) > 0 {
			s = (*p).Peers[0].DNSNames[0]
		}
		return nil
	})
	return
}

func (certs Certs) Pool() (pool *x509.CertPool, err error) {
	c, err := certs.Cache.Value()
	if err == nil {
		pool = c.Pool
	}
	return
}

func (certs Certs) Range(f func(
	headers Headers,
	cert *x509.Certificate,
) bool) {
	certs.Cache.Ref(func(p *Cache) error {
		for i, peer := range (*p).Peers {
			if !f((*p).Headers[i], peer) {
				break
			}
		}
		return nil
	})
}

func (certs Certs) SKIs() (skis []string) {
	certs.Range(func(
		headers Headers,
		cert *x509.Certificate,
	) bool {
		skis = append(skis, hex.EncodeToString(cert.SubjectKeyId))
		return true
	})
	return
}

// Preload test peers instead of parsing file.
func (certs Certs) TestLoad(peers ...*x509.Certificate) {
	certs.Cache.Preload(func(p *Cache) {
		(*p).Headers = make([]Headers, len(peers))
		(*p).Peers = make([]*x509.Certificate, len(peers))
		(*p).Pool = x509.NewCertPool()
		c, err := cert.Value()
		if err != nil {
			panic(err)
		}
		(*p).Pool.AddCert(c.Leaf)
		for i, peer := range peers {
			(*p).Headers[i] = make(Headers)
			(*p).Peers[i] = peer
			(*p).Pool.AddCert(peer)
		}
	})
}

func (file File) Load(c *Cache) (err error) {
	c.FileName, err = file.Name.Value()
	if err != nil {
		return
	}
	tlsc, err := cert.Value()
	if err != nil {
		return
	}
	c.Pool = x509.NewCertPool()
	c.Pool.AddCert(tlsc.Leaf)
	blocks, err := keycert.DecodeFile(c.FileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return
	}
	c.Headers = make([]Headers, len(blocks))
	c.Peers, err = keycert.ParseX509Certificates(blocks)
	if err == nil {
		for i, block := range blocks {
			c.Headers[i] = block.Headers
			c.Pool.AddCert(c.Peers[i])
		}
	}
	return
}
