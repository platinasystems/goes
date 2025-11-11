// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/sig"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"github.com/platinasystems/goes/v2/pkg/xsync"
)

var Features = map[string]any{
	"show": map[string]any{
		"vpn": ShowFeatures,
	},
	"vpn": MainFeatures,
}

var MainFeatures = map[string]any{
	"approve":     RestAdminReq,
	"certify":     RestCertifyReq,
	"deny":        RestAdminReq,
	"exchange":    Exchange,
	"get":         RestGetReq,
	"guest":       Guest,
	"im":          InstantMessaging,
	"lookup":      RestLookupReq,
	"reload":      RestReloadReq,
	"registry":    Registry,
	"subscribe":   RestSubscribeReq,
	"unsubscribe": RestAdminReq,
	"update":      RestUpdateReq,
}

var ShowFeatures = map[string]any{
	"address":    RestShowReq,
	"admins":     RestShowReq,
	"domain":     RestShowReq,
	"exchanges":  RestShowReq,
	"pending":    RestShowReq,
	"prefix":     RestShowReq,
	"start":      RestShowReq,
	"status":     RestShowReq,
	"subscriber": RestShowReq,
	"vcs":        RestShowReq,
}

const (
	FromTunCap = xnet.BatchCap
	FromVpnCap = xnet.BatchCap

	ToTunCap = xnet.BatchCap + 4
	ToVpnCap = xnet.BatchCap

	RestartExitCode = xos.EX_TEMPFAIL
)

const DefaultExchangePort = 8003

var (
	errEOC = errors.New("end of channel")

	errNoService = errors.New("no service addr:port")

	errUnaddressed = errors.New("unaddressed")

	errUnassigned = errors.New("unanassigned")

	errUnestablished = errors.New("unestablished")

	errUnidentified = errors.New("unidentified")

	mp = xnet.NewMsgPool(netph.ETHMTU)

	// Unix Micro start of registry
	RegistryStart int64

	wg xsync.WaitGroup
)

type empty = struct{}

var novalue, done empty
var prefix netip.Prefix

func noServiceError(args ...any) error {
	return xerrors.Label(errNoService, args...)
}

func unaddressedError(args ...any) error {
	return xerrors.Label(errUnaddressed, args...)
}

func unassignedError(args ...any) error {
	return xerrors.Label(errUnassigned, args...)
}

func unestablishedError(args ...any) error {
	return xerrors.Label(errUnestablished, args...)
}

func NewGreeting(now int64) *xnet.Msg {
	var err error

	m := mp.Get()
	m.Data = m.Data[:0]

	if now == 0 {
		now = time.Now().UnixMicro()
	}

	m.Data, err = xnet.ByteOrderAppend(m.Data, RegistryStart)
	if err != nil {
		xlog.Errata.Print(err)
		mp.Put(m)
		return nil
	}
	m.Data, err = xnet.ByteOrderAppend(m.Data, now)
	if err != nil {
		xlog.Errata.Print(err)
		mp.Put(m)
		return nil
	}
	sig := sig.Sign(m.Data)
	m.Data = append(m.Data, sig...)
	m.Data = append(m.Data, MyLabel...)
	return m
}

func unmap4in6(ap netip.AddrPort) netip.AddrPort {
	if addr := ap.Addr(); addr.Is4In6() {
		ap = netip.AddrPortFrom(addr.Unmap(), ap.Port())
	}
	return ap
}
