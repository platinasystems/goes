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
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/x509certs"
	"github.com/platinasystems/goes/v2/pkg/x509keys"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const DefaultRegistry = "https://127.0.0.1:8003"

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			"certificate": Certificate,
			"ed25519":     Ed25519,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			string(Admins):      Admins.show,
			"certificate":       Certificaté,
			"signature":         Signature,
			string(Pending):     Pending.show,
			string(Subscribers): Subscribers.show,
			"subscriptions":     Subscriptions,
		},
	},
	"vpn": map[string]any{
		string(Approve):     Approve.admin,
		string(Certify):     Certify.certify,
		string(Deny):        Deny.admin,
		string(Disable):     Disable.admin,
		string(Enable):      Enable.admin,
		"exchange":          Exchange,
		"guest":             Guest,
		string(Ping):        Ping.ping,
		"registry":          Registry,
		string(Subscribe):   Subscribe.subscribe,
		string(Unsubscribe): Unsubscribe.admin,
	},
}

var opts struct {
	crt,
	key *string
	reg    string
	regurl *url.URL
	svc    netip.AddrPort
}

func parseOpts(ctx context.Context, args []string) error {
	q := flag.Bool("q", false,
		"Quiet logging.")
	v := flag.Bool("v", false,
		"Verbose logging.")

	opts.crt = flag.String("certificate", DefaultCrt(), "File name.")
	opts.key = flag.String("key", DefaultKey(), "File name.")

	if len(opts.reg) > 0 {
		flag.StringVar(&opts.reg, "r", opts.reg, "Registry.")
	}

	if opts.svc.Addr().IsValid() {
		flag.TextVar(&opts.svc, "service", opts.svc, `{addr}:{port}
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
		if len(opts.reg) > 0 {
			opts.regurl, err = url.Parse(opts.reg)
		}
		if opts.svc.Addr().IsUnspecified() {
			err = optsvc(ctx)
		}
	}
	return err

}

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

// DefaultCrt is [xos.ConfigHome] + "/vpn.crt"
func DefaultCrt() string {
	return filepath.Join(xos.ConfigHome(), "vpn.crt")
}

// DefaultCrt is [xos.ConfigHome] + ".vpn.key"
func DefaultKey() string {
	return filepath.Join(xos.ConfigHome(), ".vpn.key")
}

// DefaultSubscriptions is [xos.StateHome] + "vpn.subscriptions"
func DefaultSubscriptions() string {
	return filepath.Join(xos.StateHome(), "vpn.subscriptions")
}

// DefaultUDPService is 0.0.0.0:0
var DefaultUDPService = func() netip.AddrPort {
	return netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
}

var crtFile = sync.OnceValues(func() (*x509certs.File, error) {
	return x509certs.NewFile(*opts.crt)
})

var keyFile = sync.OnceValues(func() (*x509keys.File, error) {
	return x509keys.NewFile(*opts.key)
})

var subscriptionsFile = sync.OnceValues(func() (*x509certs.File, error) {
	return certsFile("vpn.subscriptions")
})

func certsFile(subpath string) (*x509certs.File, error) {
	config := filepath.Join(xos.ConfigHome(), subpath)
	state := filepath.Join(xos.StateHome(), subpath)
	if _, err := os.Stat(state); err == nil {
		return x509certs.NewFile(state)
	}
	if _, err := os.Stat(config); err == nil {
		c, err := x509certs.NewFile(config)
		if err == nil {
			c.Path = state
		}
		return c, err
	}
	return &x509certs.File{Path: state}, nil
}

func optsvc(ctx context.Context) error {
	c, err := crtFile()
	if err != nil {
		return err
	}
	first := c.First()
	if first == nil {
		return xerrors.Invalid(c.Path)
	} else if len(first.DNSNames) == 0 {
		return xerrors.Invalid(c.Path, "dns")
	}
	if len(first.IPAddresses) > 0 {
		ip0 := first.IPAddresses[0]
		if a, ok := netip.AddrFromSlice(ip0); ok {
			opts.svc = netip.AddrPortFrom(a, opts.svc.Port())
			return nil
		}
		return xerrors.Invalid(c.Path, "ip")
	}
	network := "ip4"
	if opts.svc.Addr().Is6() {
		network = "ip6"
	}
	dns0 := first.DNSNames[0]
	ips, err := PatientLookupIP(ctx, network, dns0, 30*time.Second)
	if err != nil {
		return err
	} else if len(ips) == 0 {
		return xerrors.Incomplete(c.Path, "ip")
	}
	for _, ip := range ips {
		a, ok := netip.AddrFromSlice(ip)
		if !ok {
			verbose.Println("invalid", ip)
			continue
		}
		a = a.Unmap()
		if opts.svc.Addr().Is4() {
			if a.Is6() {
				verbose.Println("skipped v6 address", a)
				continue
			}
		} else if a.Is4() {
			verbose.Println("skipped v4 address", a)
			continue
		}
		opts.svc = netip.AddrPortFrom(a, opts.svc.Port())
		return nil
	}
	return xerrors.Invalid(c.Path, "has no valid IPs")
}
