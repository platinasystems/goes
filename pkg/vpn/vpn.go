// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"crypto/rand"
	"errors"
	"flag"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)

	errata  = xlog.Unmute(mutable)
	verbose = xlog.Mute(mutable)

	udpRxTrace = xlog.Mute(mutable)
	udpTxTrace = xlog.Mute(mutable)

	goRoutineTrace = xlog.Mute(mutable)
)

var ErrLLAddrUnderrun = errors.New("link-local address underrun")

// [xdg.ConfigHome] or [fhs.Config] + GOES/vpn
var ConfigDir = sync.OnceValue(func() string {
	mn := xprogram.MainName()
	sys := filepath.Join(fhs.Config(), mn, "vpn")
	if s := xdg.ConfigHome(); len(s) > 0 {
		s = filepath.Join(s, mn, "vpn")
		if fi, err := os.Stat(s); err == nil && fi.IsDir() {
			return s
		} else if fi, err = os.Stat(sys); err == nil && fi.IsDir() {
			return sys
		} else if os.Geteuid() != 0 {
			return s
		}
	}
	return sys
})

// [xdg.StateHome] or [fhs.State] + GOES/vpn
var StateDir = sync.OnceValue(func() string {
	mn := xprogram.MainName()
	sys := filepath.Join(fhs.State(), mn, "vpn")
	if s := xdg.StateHome(); len(s) > 0 {
		s = filepath.Join(s, mn, "vpn")
		if fi, err := os.Stat(s); err == nil && fi.IsDir() {
			return s
		} else if fi, err = os.Stat(sys); err == nil && fi.IsDir() {
			return sys
		} else if os.Geteuid() != 0 {
			return s
		}
	}
	return sys
})

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			"certificate": CreateCertificate,
			"signature":   NewEd25519,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			"address":     RestShow,
			"admins":      RestShow,
			"certificate": ShowCertificate,
			"pending":     RestShow,
			"signature":   ShowSignature,
			"subscriber":  RestShow,
			"tenant":      RestShow,
		},
	},
	"vpn": map[string]any{
		"approve":     RestAdmin,
		"certify":     RestCertify,
		"deny":        RestAdmin,
		"exchange":    Exchange,
		"guest":       Guest,
		"ping":        RestPing,
		"registry":    Registry,
		"subscribe":   RestSubscribe,
		"unsubscribe": RestAdmin,
	},
}

func CertFlag() *string {
	return flag.String("cert",
		filepath.Join(ConfigDir(), "cert.pem"),
		"Certificate file name.")
}

func ConfigFlag() *string {
	return flag.String("config",
		filepath.Join(ConfigDir(), "config.yaml"),
		"Configuration file name.")
}

func NatFlag() *netip.AddrPort {
	ap := new(netip.AddrPort)
	dap := netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	flag.TextVar(ap, "nat", dap, `NAT'd service {addr}:{port}.
Ignored if 0.0.0.0:0.`)
	return ap
}

func RandLinkLocalAddr() (lladdr netip.Addr, err error) {
	var a [netph.IPv6len]byte
	a[0] = 0xfe
	a[1] = 0x80
	n, err := rand.Read(a[8:])
	if err != nil {
	} else if n != len(a[8:]) {
		err = ErrLLAddrUnderrun
	} else {
		lladdr = netip.AddrFrom16(a)
		verbose.Println("link-local address:", lladdr)
	}
	return
}

func RegFlag() *string {
	return flag.String("reg",
		filepath.Join(ConfigDir(), "registry.pem"),
		"Registry certificate file name.")
}

func ServiceFlag(defport uint16) *netip.AddrPort {
	ap := new(netip.AddrPort)
	dap := netip.AddrPortFrom(netip.IPv4Unspecified(), defport)
	flag.TextVar(ap, "s", dap, `Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], listen on all ipv4 or ipv6
interface addresses.  If “port” is 0, allocate from system.`)
	return ap
}

func SigFlag() *string {
	return flag.String("sig",
		filepath.Join(ConfigDir(), "sig.pk8"),
		"Signature file name.")
}

func TunnelFlag() *uint {
	return flag.Uint("t", 0, "Tunnel unit number.")
}

func VpnFlag() *string {
	return flag.String("vpn", "",
		"Named VPN, default unnamed.")
}

func qvFlags(args []string) error {
	qFlag := flag.Bool("q", false, "Quiet logging.")
	vFlag := flag.Bool("v", false, "Verbose logging.")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if *qFlag {
		errata = xlog.Mute(errata)
	} else if *vFlag {
		verbose = xlog.Unmute(verbose)
	}
	return nil
}
