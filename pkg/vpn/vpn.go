// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net/netip"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsync"
)

const year = 365 * 24 * time.Hour

var (
	wg xsync.WaitGroup

	vpnCert         = "cert.pem"
	vpnConfig       = "config.yaml"
	vpnConfigDir    = "/etc/goes"
	vpnDuration     = 10 * year
	vpnRegistry     = "registry.pem"
	vpnSerialNumber = int64(1)
	vpnSig          = "sig.pk8"
	vpnStateDir     = "/var/run/goes"

	vpnQuiet,
	vpnTrace,
	vpnVerbose bool

	vpnListen,
	vpnPublic netip.AddrPort

	vpnCountry,
	vpnEmail,
	vpnDNS,
	vpnLocality,
	vpnName,
	vpnOrganization,
	vpnOrganizationalUnit,
	vpnPostalCode,
	vpnProvince,
	vpnStreet,
	vpnURI,
	vpnVPN string

	vpnTunnel uint
)

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
			"hosts":       RestShow,
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
		"reload":      RestReload,
		"registry":    Registry,
		"subscribe":   RestSubscribe,
		"unsubscribe": RestAdmin,
	},
}

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

func defineCommonFlags() {
	vpnConfigDir = defaultConfigDir()
	xflag.Define(&vpnQuiet, "q", "Quiet logging.")
	xflag.Define(&vpnVerbose, "v", "Verbose logging.")
	xflag.Define(&vpnCert, "cert", "Certificate file name w/in config-dir.")
	xflag.Define(&vpnConfigDir, "config-dir", "Configuration directory.")
	xflag.Define(&vpnRegistry, "registry",
		"Registry certificate file name w/in config-dir.")
	xflag.Define(&vpnSig, "sig", "Signature file name w/in config-dir.")
	xflag.Define(&vpnVPN, "vpn", "Named VPN. (default unnamed)")
}

// [xdg.ConfigHome] or [fhs.Config] + GOES/vpn
func defaultConfigDir() string {
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
}

// [xdg.StateHome] or [fhs.State] + GOES/vpn
func defaultStateDir() string {
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
}
