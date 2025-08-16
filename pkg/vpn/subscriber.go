// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/x509"
	"fmt"
	"net/netip"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type Subscriber struct {
	Id   Id
	Addr netip.Addr
	Port uint16

	// Most to least Cert.Subject.CommonName
	ExchangePrecedence []string

	// Complete ASN.1 DER content
	// (CSR, signature algorithm and signature).
	CertDER []byte

	// [mlkem.DecapsulationKey768].EncapsulationKey.Bytes
	EncapKey []byte

	sharedKey []byte `json:"-"`

	cert  *x509.Certificate `json:"-"`
	cb    cipher.Block      `json:"-"`
	gcm   cipher.AEAD       `json:"-"`
	label struct {
		fromMe, toMe []byte
	} `json:"-"`
	via netip.AddrPort `json:"-"`
}

func NewSubscriber(c *x509.Certificate) *Subscriber {
	return &Subscriber{
		CertDER: c.Raw,
		cert:    c,
	}
}

func (sub *Subscriber) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, sub.name(), ",", sub.Id)
	if sub.Addr.IsValid() {
		fmt.Fprint(w, ",", sub.Addr)
	}
	if sub.via.IsValid() {
		fmt.Fprint(w, ",", sub.via)
	}
}

func (sub *Subscriber) helloIsOK(m *xnet.Msg) bool {
	m.Data = TruncLabel(m.Data)
	if len(m.Data) < 2*8 {
		xlog.Errata.Println("incomplete hello from", sub)
	} else if !sub.verify(m.Data) {
		xlog.Errata.Println("invalid hello from", sub)
	} else {
		xlog.Trace.Println("hello from", sub)
		return true
	}
	return false
}

func (sub *Subscriber) name() string {
	return sub.cert.Subject.CommonName
}

func (sub *Subscriber) resolve(ctx context.Context) {
	var addr netip.Addr
	var ok bool
	var err error

	name := sub.name()
	if sub.via.IsValid() {
		sync.OnceFunc(func() {
			xlog.Errata.Println(name, "already", sub.via)
		})()
		return
	}
	sub.via = netip.AddrPort{}
	port := sub.Port
	switch {
	case len(sub.cert.IPAddresses) > 0:
		addr, ok = netip.AddrFromSlice(sub.cert.IPAddresses[0])
		if !ok {
			xlog.Errata.Println(name, "IPAddresses[0] invalid")
			return
		}
	case len(sub.cert.URIs) > 0:
		uri := sub.cert.URIs[0]
		if s := uri.Port(); len(s) > 0 && port == 0 {
			if _, err = fmt.Sscan(s, &port); err != nil {
				port = 0
			}
		}
		if addr, err = resolve(ctx, uri.Hostname()); err != nil {
			xlog.Errata.Println(name, err)
			return
		}
	case len(sub.cert.DNSNames) > 0:
		if addr, err = resolve(ctx, sub.cert.DNSNames[0]); err != nil {
			xlog.Errata.Println(name, err)
			return
		}
	default:
		if addr, err = resolve(ctx, name); err != nil {
			xlog.Errata.Println(name, err)
			return
		}
	}
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	if port == 0 {
		port = vpnExchangePort
	}
	sub.via = netip.AddrPortFrom(addr, port)
	xlog.Trace.Println("resolved", sub)
}

func (sub *Subscriber) setVia(ap netip.AddrPort) {
	if addr := ap.Addr(); addr.Is4In6() {
		sub.via = netip.AddrPortFrom(addr.Unmap(), ap.Port())
	} else {
		sub.via = ap
	}
}

func (sub *Subscriber) stateFileName() string {
	return filepath.Join(vpnStateDir, fmt.Sprint(sub.name(), ".pem"))
}

// Parse [Subscriber.Cert] from [Subscriber.CertDER] and validate signature.
func (sub *Subscriber) validate() error {
	var err error
	sub.cert, err = xerrors.MarkResult(x509.ParseCertificate(sub.CertDER))
	if err != nil {
		return err
	}
	switch sub.cert.SignatureAlgorithm {
	case x509.UnknownSignatureAlgorithm:
		err = xerrors.Invalid(sub.name(), "signature")
	case x509.PureEd25519:
		if _, ok := sub.cert.PublicKey.(ed25519.PublicKey); !ok {
			err = fmt.Errorf("%q: %w signature (%T)",
				sub.name(), xerrors.ErrInvalid,
				sub.cert.PublicKey)
		}
	default:
		err = xerrors.Unsupported(sub.name(), "signature")
	}
	return err
}

// Verify data with appended signature.
func (sub *Subscriber) verify(data []byte) bool {
	switch sub.cert.SignatureAlgorithm {
	case x509.PureEd25519:
		if n := len(data); n >= ed25519.SignatureSize {
			i := n - ed25519.SignatureSize
			pub := sub.cert.PublicKey.(ed25519.PublicKey)
			return ed25519.Verify(pub, data[:i], data[i:])
		}
	}
	return false
}
