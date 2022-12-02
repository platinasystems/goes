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

type File struct{ Name func() string }

var (
	SubscribersFile   = File{filename.Subscribers}
	SubscriptionsFile = File{filename.Subscriptions}
)

func Unsubscribed(ex string) error {
	return fmt.Errorf("%q: not found in %s", ex, SubscriptionsFile.Name())
}

type Headers = map[string]string

type Cert struct {
	Headers
	*x509.Certificate
	Name, SKI string
}

func NewCert(c *x509.Certificate) Cert {
	return Cert{
		Headers:     make(Headers),
		Certificate: c,
		Name:        c.DNSNames[0],
		SKI:         hex.EncodeToString(c.SubjectKeyId),
	}
}

type certs struct {
	fn string
	l  []Cert
}

type Certs struct{ cache *cache.Cache[certs] }

var Subscribers = Certs{cache.New[certs](SubscribersFile.Load)}
var Subscriptions = Certs{cache.New[certs](SubscriptionsFile.Load)}

type Pool struct{ cache *cache.Cache[*x509.CertPool] }

// ClientCAs includes self plus all Subscribers and Subscriptions.
var ClientCAs = Pool{cache.New[*x509.CertPool](func(p **x509.CertPool) error {
	*p = x509.NewCertPool()
	if tlsc, err := cert.ValErr(); err != nil {
		return err
	} else {
		(*p).AddCert(tlsc.Leaf)
	}
	Subscriptions.Range(func(c Cert) bool {
		(*p).AddCert(c.Certificate)
		return true
	})
	Subscribers.Range(func(c Cert) bool {
		(*p).AddCert(c.Certificate)
		return true
	})
	return nil
})}

// RootCAs includes self plus all Subscriptions.
var RootCAs = Pool{cache.New[*x509.CertPool](func(p **x509.CertPool) error {
	*p = x509.NewCertPool()
	if tlsc, err := cert.ValErr(); err != nil {
		return err
	} else {
		(*p).AddCert(tlsc.Leaf)
	}
	Subscriptions.Range(func(c Cert) bool {
		(*p).AddCert(c.Certificate)
		return true
	})
	return nil
})}

func (pool Pool) Add(c *x509.Certificate) {
	pool.cache.Ref(func(p **x509.CertPool) error {
		(*p).AddCert(c)
		return nil
	})
}

func (pool Pool) Clone() (cas *x509.CertPool) {
	pool.cache.Ref(func(p **x509.CertPool) error {
		cas = (*p).Clone()
		return nil
	})
	return
}

// Format as yaml like sequence to writer.
func (c Cert) Format(w fmt.State, verb rune) {
	fmt.Fprintln(w, "- name:", c.Name)
	fmt.Fprintln(w, "  subject_key_id:", c.SKI)
	fmt.Fprintln(w, "  serial_number:", c.Certificate.SerialNumber)
	fmt.Fprintln(w, "  not_before:", c.Certificate.NotBefore)
	fmt.Fprintln(w, "  not_after:", c.Certificate.NotAfter)
	fmt.Fprintln(w, "  subject:", c.Certificate.Subject)
	if len(c.EmailAddresses) > 0 {
		fmt.Fprintln(w, "email_addresses:")
		for _, email := range c.Certificate.EmailAddresses {
			fmt.Fprintln(w, "    -", email)
		}
	}
	if len(c.Certificate.DNSNames) > 0 {
		fmt.Fprintln(w, "  dns_names:")
		for _, dns := range c.Certificate.DNSNames {
			fmt.Fprintln(w, "    -", dns)
		}
	}
	if len(c.Certificate.IPAddresses) > 0 {
		fmt.Fprintln(w, "ip_addresses:")
		for _, ip := range c.Certificate.IPAddresses {
			fmt.Fprintln(w, "    -", ip)
		}
	}
	if len(c.Certificate.URIs) > 0 {
		fmt.Fprintln(w, "uris:")
		for _, uri := range c.Certificate.URIs {
			fmt.Fprintln(w, "    -", uri)
		}
	}
	fmt.Fprintln(w, "  public_key_algorithm:",
		c.Certificate.PublicKeyAlgorithm)
	fmt.Fprintln(w, "  signature_algorithm:",
		c.Certificate.SignatureAlgorithm)
	fmt.Fprintln(w, "  key_usage:")
	if c.Certificate.KeyUsage == 0 {
		fmt.Fprintln(w, "    - none")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageDigitalSignature) != 0 {
		fmt.Fprintln(w, "    -", "digital_signature")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageContentCommitment) != 0 {
		fmt.Fprintln(w, "    -", "content_commitment")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageKeyEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "key_encipherment")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageDataEncipherment) != 0 {
		fmt.Fprintln(w, "    -", "data_encipherment")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageKeyAgreement) != 0 {
		fmt.Fprintln(w, "    -", "key_agreement")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageCertSign) != 0 {
		fmt.Fprintln(w, "    -", "cert_sign")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageCRLSign) != 0 {
		fmt.Fprintln(w, "    -", "CRL_sign")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageEncipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "encipher_only")
	}
	if (c.Certificate.KeyUsage & x509.KeyUsageDecipherOnly) != 0 {
		fmt.Fprintln(w, "    -", "decipher_only")
	}
	opts := x509.VerifyOptions{
		Roots: x509.NewCertPool(),
	}
	opts.Roots.AddCert(c.Certificate)
	fmt.Fprint(w, "  signature: ")
	if _, err := c.Certificate.Verify(opts); err == nil {
		fmt.Fprintln(w, "ok")
	} else {
		fmt.Fprintln(w, err)
	}
	for k, v := range c.Headers {
		fmt.Fprint(w, "  ", k, ": ", v, "\n")
	}
	fmt.Fprintln(w, "  version:", c.Certificate.Version)
}

func (cached Certs) Add(h Headers, c *x509.Certificate) error {
	return cached.cache.Ref(func(p *certs) error {
		(*p).l = append((*p).l, Cert{
			Headers:     h,
			Certificate: c,
			Name:        c.DNSNames[0],
			SKI:         hex.EncodeToString(c.SubjectKeyId),
		})
		return keycert.AppendX509CertificatesFile(p.fn, h, c)
	})
}

func (cached Certs) Format(w fmt.State, verb rune) {
	cached.Range(func(c Cert) bool {
		fmt.Fprint(w, c)
		return true
	})
}

func (cached Certs) Lookup(ex string) (name, ski string, x *x509.Certificate) {
	cached.Range(func(c Cert) bool {
		if ex == c.SKI {
			ski = ex
			x = c.Certificate
			name = c.DNSNames[0]
			return false
		}
		for _, s := range c.Certificate.DNSNames {
			if ex == s {
				ski = c.SKI
				name = ex
				x = c.Certificate
				return false
			}
		}
		return true
	})
	return
}

func (cached Certs) Names() (names []string) {
	cached.Range(func(c Cert) bool {
		names = append(names, c.Name)
		return true
	})
	return
}

func (cached Certs) Range(f func(Cert) bool) {
	cached.cache.Ref(func(p *certs) error {
		for _, c := range (*p).l {
			if !f(c) {
				break
			}
		}
		return nil
	})
}

func (cached Certs) SKIs() (skis []string) {
	cached.Range(func(c Cert) bool {
		skis = append(skis, c.SKI)
		return true
	})
	return
}

func (file File) Load(c *certs) error {
	c.fn = file.Name()
	blocks, err := keycert.DecodeFile(c.fn)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return nil
	}
	c.l = make([]Cert, len(blocks))
	cs, err := keycert.ParseX509Certificates(blocks)
	if err == nil {
		for i, block := range blocks {
			c.l[i].Headers = block.Headers
			c.l[i].Certificate = cs[i]
			c.l[i].SKI = hex.
				EncodeToString(cs[i].SubjectKeyId)
		}
	}
	return err
}
