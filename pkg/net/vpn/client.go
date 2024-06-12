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

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/gcm"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/nonce"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

// common to guest and exchange
type client struct {
	name     string
	priv     *ecdh.PrivateKey
	pub      *ecdh.PublicKey
	pubder   []byte
	nonce    [nonce.Size]byte
	id       box.Id
	registry *url.URL
	addr     netip.Addr
	hostPrefix,
	vpnPrefix netip.Prefix
	addressed map[netip.Addr]box.Id
	gcm       map[int]*gcm.Cipher
	service   map[int]netip.AddrPort
	ver       map[int]uint8
	via       map[int]box.Id
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

	cl.addressed = make(map[netip.Addr]box.Id)
	cl.gcm = make(map[int]*gcm.Cipher)
	cl.service = make(map[int]netip.AddrPort)
	cl.via = make(map[int]box.Id)
	cl.ver = make(map[int]uint8)

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

	var via box.Id
	cl.id, via, cl.addr, cl.vpnPrefix, err =
		httpCheckin(ctx, cl.registry, cl.pubder, cl.nonce[:], optsvc...)
	if err != nil {
		return err
	}
	idi := IdIndex(cl.id)
	if via != InvalidId {
		cl.via[idi] = via
	}
	if len(optsvc) > 0 {
		cl.service[idi] = optsvc[0]
	}

	verbose.Printf("id %d %v via %v prefix %v",
		cl.id, cl.addr, via, cl.vpnPrefix)

	if cl.addr.Is4() {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 32)
	} else {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 128)
	}

	return nil
}

func (cl *client) hello(ch chan<- *box.Box, to box.Id, now time.Time) error {
	ifrom := IdIndex(cl.id)
	ito := IdIndex(to)
	cto, ok := cl.gcm[ito]
	if !ok {
		return fmt.Errorf("%d: no cipher", ito)
	}
	via, ok := cl.via[ito]
	if !ok {
		via = to
	}
	ivia := IdIndex(via)
	cvia, ok := cl.gcm[ivia]
	if !ok {
		return fmt.Errorf("%d: no exchange cipher", ivia)
	}
	svc, ok := cl.service[ivia]
	if !ok {
		return fmt.Errorf("%d: no exchange service", ivia)
	}
	bx := box.New()
	binph.TunPI{
		Proto: VPN_P_HELLO,
	}.WriteTo(bx)
	binint.BigEndianValue(now.UnixMicro()).WriteTo(bx)
	bx.AddrPort = svc
	bx.From(cl.id)
	bx.To(to)
	bx.Via(via)
	bx.CloseWith(cto)
	bx.SealWith(cvia)
	bx.NonBlockingPut(ch)
	verbose.Printf("hello %d from %d via %d", ito, ifrom, ivia)
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

	id, err := egress.MarkResult(idHeader(blk))
	if err != nil {
		return err
	}

	idi := IdIndex(id)
	cl.ver[idi] = IdVersion(id)

	addr, err := egress.MarkResult(addressHeader(blk))
	if err != nil {
		return err
	}
	cl.addressed[addr] = id

	if _, ok := blk.Headers["service"]; ok {
		cl.service[idi], err = egress.MarkResult(serviceHeader(blk))
		if err != nil {
			return err
		}
	} else if via, err := egress.MarkResult(viaHeader(blk)); err != nil {
		return err
	} else {
		cl.via[idi] = via
	}

	cl.gcm[idi], err = gcm.New(cl.priv, pub, cl.nonce[:], remNonce)
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
