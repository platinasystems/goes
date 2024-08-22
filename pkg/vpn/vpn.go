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

// [xdg.ConfigHome]/GOES/vpn
var ConfigHome = sync.OnceValue(func() string {
	return filepath.Join(xdg.ConfigHome(), xprogram.MainName(), "vpn")
})

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			string(ExchangeCF): ExchangeCF.Create,
			string(GuestCF):    GuestCF.Create,
			string(RegistryCF): RegistryCF.Create,
			"signature":        NewEd25519,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			"admins":           RestShow,
			string(ExchangeCF): ExchangeCF.Show,
			string(GuestCF):    GuestCF.Show,
			"pending":          RestShow,
			string(RegistryCF): RegistryCF.Show,
			"signature":        ShowSignature,
			"subscribers":      RestShow,
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
		dfn = filepath.Join(ConfigHome(), s+".pem")
		if _, err := os.Stat(dfn); err == nil {
			break
		}
	}
	return flag.String("a", dfn, "Certificate file.")
}

func ConfigFlag() *string {
	dfn := filepath.Join(ConfigHome(), "config.yaml")
	return flag.String("c", dfn, "Configuration file name.")
}

func ExchangeFlag() *string {
	dfn := filepath.Join(ConfigHome(), "exchange.pem")
	return flag.String("e", dfn,
		"Exchange certificate file name, “-” for stdio.")
}

func GuestFlag() *string {
	dfn := filepath.Join(ConfigHome(), "guest.pem")
	return flag.String("g", dfn,
		"Guest certificate file name, “-” for stdio.")
}

func DefaultKey() string {
	return filepath.Join(ConfigHome(), ".key")
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
	dfn := filepath.Join(ConfigHome(), "registry.pem")
	return flag.String("r", dfn,
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
