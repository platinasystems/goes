// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

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
			string(Admins):      Admins.show,
			string(ExchangeCF):  ExchangeCF.Show,
			string(GuestCF):     GuestCF.Show,
			string(Pending):     Pending.show,
			string(RegistryCF):  RegistryCF.Show,
			"signature":         ShowSignature,
			string(Subscribers): Subscribers.show,
		},
	},
	"vpn": map[string]any{
		string(Approve):     Approve.admin,
		string(Certify):     Certify.certify,
		string(Deny):        Deny.admin,
		"exchange":          Exchange,
		"guest":             Guest,
		string(Ping):        Ping.ping,
		"registry":          Registry,
		string(Subscribe):   Subscribe.subscribe,
		string(Unsubscribe): Unsubscribe.admin,
	},
}

var local struct {
	crt *Certificates
	sig *Signatures
	svc netip.AddrPort
}

var remote struct {
	crt *Certificates
	url *url.URL
}

var transport *http.Transport

func qvflags(args []string) error {
	qflag := flag.Bool("q", false, "Quiet logging.")
	vflag := flag.Bool("v", false, "Verbose logging.")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if *qflag {
		errata = xlog.Mute(errata)
	} else if *vflag {
		verbose = xlog.Unmute(verbose)
	}
	return nil
}

func mkTransport() {
	cfg := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: local.crt.DERs(),
				PrivateKey:  local.sig.First(),
			},
		},
	}

	if rcas, err := x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	} else {
		cfg.RootCAs = rcas
	}

	local.crt.Join(cfg.RootCAs)
	if remote.crt != nil {
		remote.crt.Join(cfg.RootCAs)
	}

	transport = http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = cfg
}

func AdminFlag() *string {
	dfn := filepath.Join(ConfigHome(), "registry.pem")
	return flag.String("a", dfn, "Admin ertificate.")
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

func PortFlag() *uint {
	return flag.Uint("p", 8003, "Port number of registry.")
}

func RegistryFlag() *string {
	dfn := filepath.Join(ConfigHome(), "registry.pem")
	return flag.String("r", dfn,
		"Registry certificate file name, “-” for stdio.")
}

func ServiceFlag() *netip.AddrPort {
	ap := new(netip.AddrPort)
	dap := netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	flag.TextVar(ap, "s", dap, `Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], use the first ip or ipv6 from the
address lookup of the certificate's primary DNS name.  If “port”
is 0, allocate from system.`)
	return ap
}

func TunnelFlag() *uint {
	return flag.Uint("t", 0, "Tunnel unit number.")
}

func UrlFlag() *string {
	return flag.String("u", "https://127.0.0.1:8003",
		"Registry URL. (https://<host>[:port][/<vpn>])")
}

func SvcLookup(ctx context.Context) error {
	first := local.crt.First()
	if first == nil {
		return xerrors.Invalid(local.crt.String())
	} else if len(first.DNSNames) == 0 {
		return xerrors.Invalid(local.crt.String(), "dns")
	}
	if len(first.IPAddresses) > 0 {
		ip0 := first.IPAddresses[0]
		if a, ok := netip.AddrFromSlice(ip0); ok {
			local.svc = netip.AddrPortFrom(a, local.svc.Port())
			return nil
		}
		return xerrors.Invalid(local.svc.String(), "ip")
	}
	network := "ip4"
	if local.svc.Addr().Is6() {
		network = "ip6"
	}
	dns0 := first.DNSNames[0]
	ips, err := PatientLookupIP(ctx, network, dns0, 30*time.Second)
	if err != nil {
		return err
	} else if len(ips) == 0 {
		return xerrors.Incomplete(local.crt.String(), "ip")
	}
	for _, ip := range ips {
		a, ok := netip.AddrFromSlice(ip)
		if !ok {
			verbose.Println("invalid", ip)
			continue
		}
		a = a.Unmap()
		if local.svc.Addr().Is4() {
			if a.Is6() {
				verbose.Println("skipped v6 address", a)
				continue
			}
		} else if a.Is4() {
			verbose.Println("skipped v4 address", a)
			continue
		}
		local.svc = netip.AddrPortFrom(a, local.svc.Port())
		return nil
	}
	return xerrors.Invalid(local.crt.String(), "has no valid IPs")
}
