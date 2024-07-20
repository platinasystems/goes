// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Vpn is a [goes] app providing daemons and utilities to implement and manage
// a secure, Virtual Private Network.
//
// [goes]: github.com/platinasystems/goes/v2/pkg/goes
package vpn

import (
	"context"
	_ "embed"
	"flag"
	"log"
	"net/netip"
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

const RegistryURL = "https://<regisitry>[:<port>][/<vpn>]"

//go:embed overview.txt
var overview string

var Features = map[string]any{
	"new": map[string]any{
		"vpn": map[string]any{
			"certificate": newX509Certificate,
			"ed25519-key": newEd25519Key,
		},
	},
	"show": map[string]any{
		"vpn": map[string]any{
			"admins":      restShow,
			"certificate": showX509Certificate,
			"ed25519-key": showEd25519Key,
			"pending":     restShow,
			"overview":    overview,
			"subscribers": restShow,
		},
	},
	"vpn": map[string]any{
		"approve":     restAdmin,
		"certify":     restCertify,
		"deny":        restAdmin,
		"disable":     restAdmin,
		"enable":      restAdmin,
		"exchange":    exchangeDaemon,
		"guest":       guestDaemon,
		"ping":        restPing,
		"registry":    registryDaemon,
		"subscribe":   restSubscribe,
		"unsubscriba": restAdmin,
	},
}

var opts struct {
	crt,
	key *string
	svc netip.AddrPort
}

func parseOpts(ctx context.Context, args []string) error {
	q := flag.Bool("q", false,
		"Quiet logging.")
	v := flag.Bool("v", false,
		"Verbose logging.")
	opts.crt = flag.String("certificate", defaultCrt(), "File name.")
	opts.key = flag.String("key", defaultKey(), "File name.")

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
		if opts.svc.Addr().IsUnspecified() {
			err = optsvc(ctx)
		}
	}
	return err

}

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

func defaultCrt() string {
	return filepath.Join(xos.ConfigHome(), "vpn.crt")
}

func defaultKey() string {
	return filepath.Join(xos.ConfigHome(), ".vpn.key")
}

var defaultUDPService = func() netip.AddrPort {
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
