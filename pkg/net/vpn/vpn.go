// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509certs"
	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509keys"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/log/mute"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const pathSeparatorString = string(filepath.Separator)

const (
	vpnAdminsFileName        = "vpn.admins"
	vpnCrtFileName           = "vpn.crt"
	vpnHostsFileName         = "vpn.hosts"
	vpnKeyFileName           = ".vpn.key"
	vpnMirrorsFileName       = "vpn.mirrors"
	vpnPrefixFileName        = "vpn.prefix"
	vpnRegistrySyntax        = "https://<regsitry>[:<port>][/<vpn>]"
	vpnSubscribersFileName   = "vpn.subscribers"
	vpnSubscriptionsFileName = "vpn.subscriptions"
)

var (
	Commands = map[string]any{
		"approve":     httpAdmin,
		"certify":     httpCertify,
		"deny":        httpAdmin,
		"disable":     httpAdmin,
		"enable":      httpAdmin,
		"generate":    generate,
		"ping":        httpPing,
		"subscribe":   httpSubscribe,
		"unsubscribe": httpAdmin,
	}
	Daemons = map[string]any{
		"exchange": new(exchange).daemon,
		"guest":    new(guest).daemon,
		"registry": new(registry).daemon,
	}
	Shows = map[string]any{
		"active":      httpShow,
		"admins":      httpShow,
		"certificate": showCertificateFile,
		"key":         showKeyFile,
		"mirrors":     httpShow,
		"pending":     httpShow,
	}
)

var errata = mute.Off(log.New(os.Stdout, "", log.Lshortfile))
var verbose = mute.On(log.New(os.Stdout, "", log.Lshortfile))

var vpnCrtFile = sync.OnceValues(func() (*x509certs.File, error) {
	return x509certs.NewFile(vpnCrtPath())
})

var vpnCrtPath = sync.OnceValue(func() string {
	return filepath.Join(program.ConfigHome(), vpnCrtFileName)
})

var vpnKeyFile = sync.OnceValues(func() (*x509keys.File, error) {
	return x509keys.NewFile(vpnKeyPath())
})

var vpnKeyPath = sync.OnceValue(func() string {
	return filepath.Join(program.ConfigHome(), vpnKeyFileName)
})

var vpnSubscriptionsFile = sync.OnceValues(func() (*x509certs.File, error) {
	return vpnCertsFile(vpnSubscriptionsFileName)
})

func vpnCertsFile(subpath string) (*x509certs.File, error) {
	config := filepath.Join(program.ConfigHome(), subpath)
	state := filepath.Join(program.StateHome(), subpath)
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

func vpnDaemonFlags(
	ctx context.Context, usage string, args []string,
) (
	netip.AddrPort, error,
) {
	flags := goes.ContextFlags(ctx)
	qFlag := flags.Bool("q", false, "Silence most logs.")
	vFlag := flags.Bool("v", false, "Log everything.")
	svc := netip.AddrPortFrom(netip.IPv4Unspecified(), 8003)
	flags.TextVar(&svc, "service", svc, `
If <addr> of <addr>:<port> is 0.0.0.0 or [::],
lookup first ipv4 or ipv6 address of certificate's
primary DNS name.`[1:])

	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return svc, err
	}
	if goes.ContextHelp(ctx) {
		branch := goes.ContextBranch(ctx)
		if len(branch) > 1 && branch[1] == "daemon" {
			branch[1] = "start"
		}
		return svc, goes.Usage(ctx, usage)
	}

	if *vFlag {
		verbose = mute.Off(verbose)
	} else if *qFlag {
		errata = mute.On(errata)
	}

	if !svc.Addr().IsUnspecified() {
		return svc, nil
	}

	c, err := vpnCrtFile()
	if err != nil {
		return svc, err
	}
	first := c.First()
	if first == nil {
		return svc, fmt.Errorf("%s: %w", c.Path, ErrNoCertificates)
	} else if len(first.DNSNames) == 0 {
		return svc, fmt.
			Errorf("%s: %w, has no DNS name", c.Path, ErrInvalid)
	}
	if len(first.IPAddresses) > 0 {
		ip0 := first.IPAddresses[0]
		if a, ok := netip.AddrFromSlice(ip0); ok {
			return netip.AddrPortFrom(a, svc.Port()), nil
		}
		return svc, fmt.Errorf("%s: invalid IP: %v", c.Path, ip0)
	}
	network := "ip4"
	if svc.Addr().Is6() {
		network = "ip6"
	}
	dns0 := first.DNSNames[0]
	ips, err := PatientLookupIP(ctx, network, dns0, 30*time.Second)
	if err != nil {
		return svc, err
	} else if len(ips) == 0 {
		return svc, ErrNoServiceIP
	}
	for _, ip := range ips {
		a, ok := netip.AddrFromSlice(ip)
		if !ok {
			verbose.Println("invalid", ip)
			continue
		}
		a = a.Unmap()
		if svc.Addr().Is4() {
			if a.Is6() {
				verbose.Println("skipped v6 address", a)
				continue
			}
		} else if a.Is4() {
			verbose.Println("skipped v4 address", a)
			continue
		}
		verbose.Println("selected address", a)
		return netip.AddrPortFrom(a, svc.Port()), nil
	}
	return svc, fmt.Errorf("%s: no valid IP", c.Path)
}

func vpnName(dir string) string {
	name := strings.TrimPrefix(dir, program.ConfigHome())
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

func showCertificateFile(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<file>]
Print parsed certificate file with default "{{file}}" and "-" for stdin.`
	var (
		err  error
		path string
		rc   io.ReadCloser
	)
	if goes.ContextComplete(ctx) {
		return complete.Last(args, "vpn.*")
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["file"] = vpnCrtPath
		return goes.Usage(ctx, usage)
	}
	if len(args) > 0 {
		path = args[0]
	} else {
		path = vpnCrtPath()
	}
	if path == "-" {
		rc = goes.ContextStdin(ctx)
	} else if rc, err = os.Open(path); err != nil {
		return err
	} else {
		defer rc.Close()
	}
	crt := &x509certs.File{Path: path}
	if _, err = crt.ReadFrom(rc); err == nil {
		err = crt.Show(goes.ContextStdout(ctx))
	}
	return err
}

func showKeyFile(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<file>]
Print parsed key file with default "{{file}}" and "-" for stdin.`
	var (
		err  error
		path string
		rc   io.ReadCloser
	)
	if goes.ContextComplete(ctx) {
		return complete.Last(args, ".vpn.*", "vpn.*", "*.key")
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["file"] = vpnKeyPath
		return goes.Usage(ctx, usage)
	}
	if len(args) > 0 {
		path = args[0]
	} else {
		path = vpnKeyPath()
	}
	if path == "-" {
		rc = goes.ContextStdin(ctx)
	} else if rc, err = os.Open(path); err != nil {
		return err
	} else {
		defer rc.Close()
	}
	key := &x509keys.File{Path: path}
	if _, err = key.ReadFrom(rc); err == nil {
		err = key.Show(goes.ContextStdout(ctx))
	}
	return err
}
