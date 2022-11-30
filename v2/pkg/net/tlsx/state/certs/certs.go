// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package certs

import (
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type File struct {
	Name func() string
}

var (
	SubscribersFile   = File{filename.Subscribers}
	SubscriptionsFile = File{filename.Subscriptions}
)

func Unsubscribed(ex string) error {
	return fmt.Errorf("%q: not found in %s", ex, SubscriptionsFile.Name())
}

type Headers = map[string]string

type Entry struct {
	Headers
	*x509.Certificate
	Name, SKI string
}

func NewEntry(c *x509.Certificate) Entry {
	return Entry{
		Headers:     make(Headers),
		Certificate: c,
		Name:        c.DNSNames[0],
		SKI:         hex.EncodeToString(c.SubjectKeyId),
	}
}

type Cache struct {
	FileName string
	Entries  []Entry
	Pool     *x509.CertPool
}

type Certs struct {
	Cache *cache.Cache[Cache]
}

var (
	Subscribers   = Certs{cache.New[Cache](SubscribersFile.Load)}
	Subscriptions = Certs{cache.New[Cache](SubscriptionsFile.Load)}
)

// Format entry as yaml like sequence to writer.
func (entry Entry) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w, "- name:", entry.Name)
	fmt.Fprintln(w, "  subject_key_id:", entry.SKI)
	fmt.Fprintln(w, "  serial_number:", entry.Certificate.SerialNumber)
	fmt.Fprintln(w, "  not_before:", entry.Certificate.NotBefore)
	fmt.Fprintln(w, "  not_after:", entry.Certificate.NotAfter)
	fmt.Fprintln(w, "  subject:", entry.Certificate.Subject)
	if len(entry.EmailAddresses) > 0 {
		fmt.Fprintln(w, "email_addresses:")
		for _, email := range entry.Certificate.EmailAddresses {
			fmt.Fprintln(w, "    -", email)
		}
	}
	if len(entry.Certificate.DNSNames) > 0 {
		fmt.Fprintln(w, "  dns_names:")
		for _, dns := range entry.Certificate.DNSNames {
			fmt.Fprintln(w, "    -", dns)
		}
	}
	if len(entry.Certificate.IPAddresses) > 0 {
		fmt.Fprintln(w, "ip_addresses:")
		for _, ip := range entry.Certificate.IPAddresses {
			fmt.Fprintln(w, "    -", ip)
		}
	}
	if len(entry.Certificate.URIs) > 0 {
		fmt.Fprintln(w, "uris:")
		for _, uri := range entry.Certificate.URIs {
			fmt.Fprintln(w, "    -", uri)
		}
	}
	fmt.Fprintln(w, "  public_key_algorithm:", entry.Certificate.PublicKeyAlgorithm)
	fmt.Fprintln(w, "  signature_algorithm:", entry.Certificate.SignatureAlgorithm)
	fmt.Fprintln(w, "  key_usage:")
	if entry.Certificate.KeyUsage == 0 {
		fmt.Fprintln(w, "    - none")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageDigitalSignature) != 0 {
		fmt.Fprintln(w, "    -", "digital_signature")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageContentCommitment) != 0 {
		fmt.Fprintln(w, "    -", "content_commitment")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageKeyEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "key_encipherment")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageDataEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "data_encipherment")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageKeyAgreement) != 0 {
		fmt.Fprintln(w, "    -", "key_agreement")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageCertSign) != 0 {
		fmt.Fprintln(w, "    -", "cert_sign")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageCRLSign) != 0 {
		fmt.Fprintln(w, "    -", "CRL_sign")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageEncipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "encipher_only")
	}
	if (entry.Certificate.KeyUsage & x509.KeyUsageDecipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "decipher_only")
	}
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(entry.Certificate)
	fmt.Fprint(w, "  signature: ")
	if _, err := entry.Certificate.Verify(opts); err == nil {
		fmt.Fprintln(w, "ok")
	} else {
		fmt.Fprintln(w, err)
	}
	for k, v := range entry.Headers {
		fmt.Fprint(w, "  ", k, ": ", v, "\n")
	}
	fmt.Fprintln(w, "  version:", entry.Certificate.Version)
}

func (certs Certs) Add(h Headers, c *x509.Certificate) error {
	return certs.Cache.Ref(func(p *Cache) error {
		(*p).Entries = append((*p).Entries, Entry{
			Headers:     h,
			Certificate: c,
			Name:        c.DNSNames[0],
			SKI:         hex.EncodeToString(c.SubjectKeyId),
		})
		(*p).Pool.AddCert(c)
		if len(p.FileName) == 0 {
			return nil
		}
		return keycert.AppendX509CertificatesFile(p.FileName, h, c)
	})
}

func (certs Certs) Pool() *x509.CertPool {
	return certs.Cache.Value().Pool
}

func (certs Certs) Range(f func(Entry) bool) {
	certs.Cache.Ref(func(p *Cache) error {
		for _, entry := range (*p).Entries {
			if !f(entry) {
				break
			}
		}
		return nil
	})
}

func (certs Certs) Names() (names []string) {
	certs.Range(func(entry Entry) bool {
		names = append(names, entry.Name)
		return true
	})
	return
}

func (certs Certs) SKIs() (skis []string) {
	certs.Range(func(entry Entry) bool {
		skis = append(skis, entry.SKI)
		return true
	})
	return
}

// Preload test certificates instead of parsing file.
func (certs Certs) TestLoad(cs ...*x509.Certificate) {
	certs.Cache.Preload(func(p *Cache) {
		(*p).Entries = make([]Entry, len(cs))
		(*p).Pool = x509.NewCertPool()
		c, err := cert.ValErr()
		if err != nil {
			panic(err)
		}
		(*p).Pool.AddCert(c.Leaf)
		for i, c := range cs {
			(*p).Entries[i].Headers = make(Headers)
			(*p).Entries[i].Certificate = c
			(*p).Entries[i].SKI = hex.EncodeToString(c.SubjectKeyId)
			(*p).Pool.AddCert(c)
		}
	})
}

func (file File) Load(c *Cache) (err error) {
	c.FileName = file.Name()
	if tlsc, err := cert.ValErr(); err == nil {
		c.Pool = x509.NewCertPool()
		c.Pool.AddCert(tlsc.Leaf)
	}
	blocks, err := keycert.DecodeFile(c.FileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return
	}
	c.Entries = make([]Entry, len(blocks))
	cs, err := keycert.ParseX509Certificates(blocks)
	if err == nil {
		for i, block := range blocks {
			c.Entries[i].Headers = block.Headers
			c.Entries[i].Certificate = cs[i]
			c.Entries[i].SKI =
				hex.EncodeToString(cs[i].SubjectKeyId)
			c.Pool.AddCert(cs[i])
		}
	}
	return
}
