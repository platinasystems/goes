// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/gcm"
	"github.com/platinasystems/goes/v2/pkg/nonce"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

// common to guest and exchange
type client struct {
	name,
	udpv string
	rest
	priv   *ecdh.PrivateKey
	pub    *ecdh.PublicKey
	pubder []byte
	nonce  [nonce.Size]byte
	id     box.Id
	addr   netip.Addr
	hostPrefix,
	vpnPrefix netip.Prefix
	addressed map[netip.Addr]box.Id
	gcm       map[int]*gcm.Cipher
	service   map[int]netip.AddrPort
	ver       map[int]uint8
	via       map[int]box.Id
}

func defineClientFlags(listen netip.AddrPort) {
	vpnListen = listen
	xflag.Define(&vpnListen, "listen", `
Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], listen on all ipv4 or ipv6
interface addresses.  If “port” is 0, allocate from system.`[1:])
	xflag.Define(&vpnTrace, "trace", "Log packet forwarding.")
}

func (cl *client) setup() error {
	cl.udpv = "udp"
	if a := vpnListen.Addr(); a.Is4() {
		cl.udpv = "udp4"
	} else if a.Is6() {
		cl.udpv = "udp6"
	}
	return cl.rest.setup()
}

func (cl *client) register(
	ctx context.Context,
	optsvc netip.AddrPort,
) error {
	err := cl.waitForDNS(ctx)
	if err != nil {
		return err
	}

	cl.name = cl.crt.Subject.CommonName
	cl.addressed = make(map[netip.Addr]box.Id)
	cl.gcm = make(map[int]*gcm.Cipher)
	cl.service = make(map[int]netip.AddrPort)
	cl.via = make(map[int]box.Id)
	cl.ver = make(map[int]uint8)

	cl.priv, err = ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return xerrors.Label(err, "GenerateX25519Key")
	}

	cl.pub = cl.priv.PublicKey()
	cl.pubder, err = x509.MarshalPKIXPublicKey(cl.pub)
	if err != nil {
		return xerrors.Label(err, "MarshalPublicKey")
	}

	n, err := rand.Read(cl.nonce[:])
	if err != nil {
		return xerrors.Label(err, "ReadNonce")
	} else if n != nonce.Size {
		return xerrors.Invalid("LocalNonce")
	}

	var via box.Id
	cl.id, via, cl.addr, cl.vpnPrefix, err =
		cl.checkin(ctx, cl.pubder, cl.nonce[:], optsvc)
	if err != nil {
		return err
	}
	idi := cl.id.Index()
	if via != box.InvalidId {
		cl.via[idi] = via
	}
	cl.service[idi] = optsvc

	if cl.addr.Is4() {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 32)
	} else {
		cl.hostPrefix = netip.PrefixFrom(cl.addr, 128)
	}

	xlog.Info.Printf("assigned id %d, ver %d, at %v, via %v, vpn %v",
		cl.id.Index(), cl.id.Version(), cl.hostPrefix,
		via, cl.vpnPrefix)

	return nil
}

func (cl *client) peer(blk *pem.Block) error {
	pub, err := ecdhPublicKey(blk)
	if err != nil {
		return xerrors.Label(err, "PeerPubKey")
	}

	remNonce, err := nonceHeader(blk)
	if err != nil {
		return xerrors.Label(err, "PeerNonce")
	} else if len(remNonce) != nonce.Size {
		return xerrors.Invalid("PeerNonce")
	}

	id, err := idHeader(blk)
	if err != nil {
		return xerrors.Label(err, "AssignedId")
	}

	idi := id.Index()
	cl.ver[idi] = id.Version()

	addr, err := addressHeader(blk)
	if err != nil {
		return xerrors.Label(err, "PeerAddress")
	}
	cl.addressed[addr] = id

	if _, ok := blk.Headers["service"]; ok {
		cl.service[idi], err = serviceHeader(blk)
		if err != nil {
			return xerrors.Label(err, "PeerService")
		}
	} else if via, err := viaHeader(blk); err != nil {
		return xerrors.Label(err, "PeerVia")
	} else {
		cl.via[idi] = via
	}

	cl.gcm[idi], err = gcm.New(cl.priv, pub, cl.nonce[:], remNonce)
	return err
}

func (cl *client) waitForDNS(ctx context.Context) error {
	const timeout = 60 * time.Second
	hostname := cl.url.Hostname()
	ips, err := PatientLookupIP(ctx, "ip", hostname, timeout)
	if err == nil {
		xlog.Info.Println(hostname, ips)
	}
	return err
}
