// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"flag"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

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

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			"exchange":  CreateCertificate,
			"guest":     CreateCertificate,
			"registry":  CreateCertificate,
			"signature": NewEd25519,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			"admins":      RestShow,
			"exchange":    ShowCertificate,
			"guest":       ShowCertificate,
			"pending":     RestShow,
			"registry":    ShowCertificate,
			"signature":   ShowSignature,
			"subscribers": RestShow,
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

func AdminFlag() *string {
	var dfn string
	for _, s := range []string{"guest", "exchange", "registry"} {
		dfn = filepath.Join(ConfigDir(), s+".pem")
		if _, err := os.Stat(dfn); err == nil {
			break
		}
	}
	return flag.String("a", dfn, "Certificate file.")
}

func ConfigFlag() *string {
	return flag.String("c", filepath.Join(ConfigDir(), "config.yaml"),
		"Configuration file name.")
}

func ExchangeFlag() *string {
	return flag.String("e", filepath.Join(ConfigDir(), "exchange.pem"),
		"Exchange certificate file name, “-” for stdio.")
}

func GuestFlag() *string {
	return flag.String("g", filepath.Join(ConfigDir(), "guest.pem"),
		"Guest certificate file name, “-” for stdio.")
}

func DefaultKey() string {
	return filepath.Join(ConfigDir(), ".key")
}

func KeyFlag() *string {
	return flag.String("k", DefaultKey(), "Key file name, “-” for stdio.")
}

func NatFlag() *netip.AddrPort {
	ap := new(netip.AddrPort)
	dap := netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	flag.TextVar(ap, "n", dap, `NAT'd service {addr}:{port}.
Ignored if 0.0.0.0:0.`)
	return ap
}

func RegistryFlag() *string {
	return flag.String("r", filepath.Join(ConfigDir(), "registry.pem"),
		"Registry certificate file name, “-” for stdio.")
}

func ServiceFlag(defport uint16) *netip.AddrPort {
	ap := new(netip.AddrPort)
	dap := netip.AddrPortFrom(netip.IPv4Unspecified(), defport)
	flag.TextVar(ap, "s", dap, `Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], listen on all ipv4 or ipv6
interface addresses.  If “port” is 0, allocate from system.`)
	return ap
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
