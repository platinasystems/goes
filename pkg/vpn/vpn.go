// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"log"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const DefaultRegistry = "https://127.0.0.1:8003"

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			"certificate": NewCertificate,
			"ed25519":     NewEd25519,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			string(Admins):      Admins.show,
			"certificate":       ShowCertificate,
			"signature":         ShowSignature,
			string(Pending):     Pending.show,
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

var Flags struct {
	// If len(Flags.FN.*) > 0, add respective [flag.StringVar] to
	// [flag.CommandLine].
	FN struct {
		Cfg,
		Crt,
		Key,
		Subscriptions string
	}
	// If len(Flags.Reg.String) > 0, add  [flag.StringVar] to
	// [flag.CommandLine] and parse [Flags.Reg.URL] after
	// [flag.CommandLine.Parse].
	Reg struct {
		String string
		URL    *url.URL
	}
	// If valid, and [flag.TextVar] to [flag.CommandLine].
	Svc netip.AddrPort
}

func AddAndParseFlags(ctx context.Context, args []string) error {
	q := flag.Bool("q", false, "Quiet logging.")
	v := flag.Bool("v", false, "Verbose logging.")

	if len(Flags.FN.Cfg) > 0 {
		flag.StringVar(&Flags.FN.Cfg, "c", Flags.FN.Cfg,
			"Configuration file name.")
	}
	if len(Flags.FN.Crt) > 0 {
		flag.StringVar(&Flags.FN.Crt, "x", Flags.FN.Crt,
			"X509 certificate file name, “-” for stdio.")
	}
	if len(Flags.FN.Key) > 0 {
		flag.StringVar(&Flags.FN.Key, "k", Flags.FN.Key,
			"Signature key file name, “-” for stdio.")
	}
	if len(Flags.FN.Subscriptions) > 0 {
		flag.StringVar(&Flags.FN.Subscriptions, "R",
			Flags.FN.Subscriptions, "Registry certificate(s).")
	}
	if len(Flags.Reg.String) > 0 {
		flag.StringVar(&Flags.Reg.String, "r", Flags.Reg.String,
			"Registry URL.")
	}
	if Flags.Svc.Addr().IsValid() {
		flag.TextVar(&Flags.Svc, "ap", Flags.Svc,
			`Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], use the first ip or ipv6 from the
address lookup of the certificate's primary DNS name.  If “port”
is 0, allocate from system.`)
	}

	err := flag.CommandLine.Parse(args)
	if err == nil {
		if *q {
			errata = xlog.Mute(errata)
		} else if *v {
			verbose = xlog.Unmute(verbose)
		}
		if len(Flags.Reg.String) > 0 {
			Flags.Reg.URL, err = url.Parse(Flags.Reg.String)
		}
		if Flags.Svc.Addr().IsUnspecified() {
			err = SvcLookup(ctx)
		}
	}
	return err

}

// DefaultCfg is [xos.ConfigHome] + "/vpn.yaml"
func DefaultCfg() string {
	return filepath.Join(xos.ConfigHome(), "vpn.yaml")
}

// DefaultCrt is [xos.ConfigHome] + "/vpn.pem"
func DefaultCrt() string {
	return filepath.Join(xos.ConfigHome(), "vpn.pem")
}

// DefaultKey is [xos.ConfigHome] + "/.vpn"
func DefaultKey() string {
	return filepath.Join(xos.ConfigHome(), ".vpn")
}

// DefaultSubscriptions is [xos.ConfigHome] + "/vpn-subscriptions.pem"
func DefaultSubscriptions() string {
	return filepath.Join(xos.ConfigHome(), "vpn-subscriptions.pem")
}

// DefaultUDPService is 0.0.0.0:0
var DefaultUDPService = func() netip.AddrPort {
	return netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
}

func SvcLookup(ctx context.Context) error {
	c, err := Crt()
	if err != nil {
		return err
	}
	first := c.First()
	if first == nil {
		return xerrors.Invalid(c.String())
	} else if len(first.DNSNames) == 0 {
		return xerrors.Invalid(c.String(), "dns")
	}
	if len(first.IPAddresses) > 0 {
		ip0 := first.IPAddresses[0]
		if a, ok := netip.AddrFromSlice(ip0); ok {
			Flags.Svc = netip.AddrPortFrom(a, Flags.Svc.Port())
			return nil
		}
		return xerrors.Invalid(c.String(), "ip")
	}
	network := "ip4"
	if Flags.Svc.Addr().Is6() {
		network = "ip6"
	}
	dns0 := first.DNSNames[0]
	ips, err := PatientLookupIP(ctx, network, dns0, 30*time.Second)
	if err != nil {
		return err
	} else if len(ips) == 0 {
		return xerrors.Incomplete(c.String(), "ip")
	}
	for _, ip := range ips {
		a, ok := netip.AddrFromSlice(ip)
		if !ok {
			verbose.Println("invalid", ip)
			continue
		}
		a = a.Unmap()
		if Flags.Svc.Addr().Is4() {
			if a.Is6() {
				verbose.Println("skipped v6 address", a)
				continue
			}
		} else if a.Is4() {
			verbose.Println("skipped v4 address", a)
			continue
		}
		Flags.Svc = netip.AddrPortFrom(a, Flags.Svc.Port())
		return nil
	}
	return xerrors.Invalid(c.String(), "has no valid IPs")
}
