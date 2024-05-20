// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/netip"
	"net/url"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/gcm"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/nonce"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

// common to guest and exchange
type client struct {
	name     string
	priv     *ecdh.PrivateKey
	pub      *ecdh.PublicKey
	pubder   []byte
	nonce    [nonce.Size]byte
	label    label.Label
	registry *url.URL
	addr     netip.Addr
	hostPrefix,
	vpnPrefix netip.Prefix
	addressed map[netip.Addr]label.Label
	gcm       map[label.Index]*gcm.Cipher
	service   map[label.Index]netip.AddrPort
	ver       map[label.Index]label.Version
	via       map[label.Index]label.Label
}

func (cl *client) register(
	ctx context.Context, registry string, optsvc ...netip.AddrPort,
) error {
	crt, err := vpnCrtFile()
	if err != nil {
		return err
	}
	cl.name = crt.First().Subject.CommonName

	if cl.registry, err = url.Parse(registry); err != nil {
		return err
	} else if len(cl.registry.Scheme) == 0 {
		cl.registry.Scheme = "https"
	}

	err = waitForDNS(ctx, "ip", cl.registry.Hostname())
	if err != nil {
		return err
	}

	cl.addressed = make(map[netip.Addr]label.Label)
	cl.gcm = make(map[label.Index]*gcm.Cipher)
	cl.service = make(map[label.Index]netip.AddrPort)
	cl.via = make(map[label.Index]label.Label)
	cl.ver = make(map[label.Index]label.Version)

	cl.priv, err = egress.MarkResult(ecdh.X25519().GenerateKey(rand.Reader))
	if err != nil {
		return err
	}

	cl.pub = cl.priv.PublicKey()
	cl.pubder, err = egress.MarkResult(x509.MarshalPKIXPublicKey(cl.pub))
	if err != nil {
		return err
	}

	n, err := egress.MarkResult(rand.Read(cl.nonce[:]))
	if err != nil {
		return err
	} else if n != nonce.Size {
		return fmt.Errorf("nonce: %w", ErrUnderrun)
	}

	var via label.Label
	cl.label, via, cl.addr, cl.vpnPrefix, err =
		httpCheckin(ctx, cl.registry, cl.pubder, cl.nonce[:], optsvc...)
	if err != nil {
		return err
	}
	if via != label.Mislabel {
		cl.via[cl.label.Index()] = via
	}
	if len(optsvc) > 0 {
		cl.service[cl.label.Index()] = optsvc[0]
	}

	verbose.Println("label", cl.label, "via", via, "addr", cl.addr,
		"prefix", cl.vpnPrefix)

	if cl.addr.Is4() {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 32)
	} else {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 128)
	}

	return nil
}

func (cl *client) peer(blk *pem.Block) error {
	pub, err := egress.MarkResult(ecdhPublicKey(blk))
	if err != nil {
		return err
	}

	remNonce, err := egress.MarkResult(nonceHeader(blk))
	if err != nil {
		return err
	} else if len(remNonce) != nonce.Size {
		return egress.Mark(ErrUnderrun)
	}

	lbl, err := egress.MarkResult(labelHeader(blk))
	if err != nil {
		return err
	}

	lbli := lbl.Index()
	cl.ver[lbli] = lbl.Version()

	addr, err := egress.MarkResult(addressHeader(blk))
	if err != nil {
		return err
	}
	cl.addressed[addr] = lbl

	if _, ok := blk.Headers["service"]; ok {
		cl.service[lbli], err = egress.MarkResult(serviceHeader(blk))
		if err != nil {
			return err
		}
	} else if via, err := egress.MarkResult(viaHeader(blk)); err != nil {
		return err
	} else {
		cl.via[lbli] = via
	}

	cl.gcm[lbli], err = gcm.New(cl.priv, pub, cl.nonce[:], remNonce)
	return err
}

func waitForDNS(ctx context.Context, network, hostname string) error {
	const timeout = 60 * time.Second
	switch network {
	case "udp":
		network = "ip"
	case "udp4":
		network = "ip4"
	case "udp6":
		network = "ip6"
	}
	ips, err := PatientLookupIP(ctx, network, hostname, timeout)
	if err == nil {
		verbose.Println(hostname, ips)
	}
	return err
}
