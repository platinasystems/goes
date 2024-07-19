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
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/x509certs"
	"github.com/platinasystems/goes/v2/pkg/x509keys"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const (
	AdminsFileName      = "vpn.admins"
	HostsFileName       = "vpn.hosts"
	PrefixFileName      = "vpn.prefix"
	SubscribersFileName = "vpn.subscribers"
)

const RegistryURL = "https://<regisitry>[:<port>][/<vpn>]"

const DefaultPort = 8003

var DefaultService = func() netip.AddrPort {
	return netip.AddrPortFrom(netip.IPv4Unspecified(), DefaultPort)
}

var Features = map[string]any{
	"generate": map[string]any{
		"certificate": generateX509Certificate,
		"key":         generateEd25519Key,
	},
	"show": map[string]any{
		"certificate": showX509Certificate,
		"key":         showX509Certificate,
		"vpn": map[string]any{
			"admins":      restShow,
			"pending":     restShow,
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

var errata = xlog.Unmute(log.New(os.Stdout, "", log.Lshortfile))
var verbose = xlog.Mute(log.New(os.Stdout, "", log.Lshortfile))

var opts struct {
	crt,
	key,
	subscriptions *string
	svc netip.AddrPort
}

func parseOpts(ctx context.Context, args []string) error {
	q := flag.Bool("q", false,
		"Quiet logging.")
	v := flag.Bool("v", false,
		"Verbose logging.")
	opts.crt = flag.String("certificate",
		filepath.Join(xos.ConfigHome(), "vpn.crt"),
		"File name.")
	opts.key = flag.String("key",
		filepath.Join(xos.ConfigHome(), ".vpn.key"),
		"File name.")
	opts.subscriptions = flag.String("subscriptions",
		filepath.Join(xos.StateHome(), "vpn.subscriptions"),
		"File name.")
	if opts.svc.IsValid() {
		flag.TextVar(&opts.svc, "service", opts.svc, `<addr>:<port>
If <addr> is 0.0.0.0 or [::], use the first ip address of
the certificate's primary DNS name.`)
	}
	err := flag.CommandLine.Parse(args)
	if err == nil {
		if *q {
			errata = xlog.Mute(errata)
		} else if *v {
			verbose = xlog.Unmute(verbose)
		}
		if opts.svc.Addr().IsUnspecified() {
			err = crtsvc(ctx)
		}
	}
	return err

}

var crtFile = sync.OnceValues(func() (*x509certs.File, error) {
	return x509certs.NewFile(crtPath())
})

var crtPath = sync.OnceValue(func() string {
	return filepath.Join(xos.ConfigHome(), "vpn.crt")
})

var keyFile = sync.OnceValues(func() (*x509keys.File, error) {
	return x509keys.NewFile(keyPath())
})

var keyPath = sync.OnceValue(func() string {
	return filepath.Join(xos.ConfigHome(), ".vpn.key")
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

func crtsvc(ctx context.Context) error {
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

func vpnName(dir string) string {
	const pathSeparatorString = string(filepath.Separator)
	name := strings.TrimPrefix(dir, xos.ConfigHome())
	return strings.TrimLeft(name, pathSeparatorString)
}

func vpnPrefix(path string) (netip.Prefix, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return netip.Prefix{}, err
	}
	prefix, err := netip.ParsePrefix(strings.TrimSpace(string(data)))
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("%s: %w", path, err)
	}
	return prefix, nil
}
