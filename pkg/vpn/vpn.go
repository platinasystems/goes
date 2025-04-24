// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	DefaultCertFile   = "cert.pem"
	DefaultConfigFile = "config.yaml"
	DefaultRegFile    = "registry.pem"
	DefaultSigFile    = "sig.pk8"
)

const (
	CertFlag xflag.KeyUsage[string] = `cert
Certificate file name w/in config-dir.`

	ConfigFlag xflag.KeyUsage[string] = `config
Configuration file name w/in config-dir.`

	ConfigDirFlag xflag.KeyUsage[string] = `config-dir
Configuration directory.`

	CountryFlag xflag.KeyUsage[string] = `country`

	EmailFlag xflag.KeyUsage[string] = `email Comma separated addresses.`

	DNSFlag xflag.KeyUsage[string] = `dns
Comma separated domain names.`

	DurationFlag xflag.KeyUsage[time.Duration] = `duration
e.g. 360s, 60m, or 1h.`

	ListenFlag xflag.KeyUsage[netip.AddrPort] = `listen
Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], listen on all ipv4 or ipv6
interface addresses.  If “port” is 0, allocate from system.`

	LocalityFlag xflag.KeyUsage[string] = `locality aka. city.`

	NameFlag xflag.KeyUsage[string] = `name VPN identfier.`

	OrganizationFlag xflag.KeyUsage[string] = `organization aka. company`

	OrganizationalUnitFlag xflag.KeyUsage[string] = `organizational-unit
aka. department.`

	PostalCodeFlag xflag.KeyUsage[string] = `postal-code aka. zip.`

	ProvinceFlag xflag.KeyUsage[string] = `province aka. state.`

	PublicFlag xflag.KeyUsage[netip.AddrPort] = `public
NAT'd listen {addr}:{port}. (0.0.0.0:0 ignored)`

	QuietFlag xflag.KeyUsage[bool] = `q Quiet logging.`

	RegFlag xflag.KeyUsage[string] = `reg
Registry certificate file name w/in config-dir.`

	SerialNumberFlag xflag.KeyUsage[int64] = `serial-number`

	SigFlag xflag.KeyUsage[string] = `sig
Signature file name w/in config-dir.`

	StateDirFlag xflag.KeyUsage[string] = `state-dir
State directory to save approved client certificates.`

	StreetFlag xflag.KeyUsage[string] = `street address`

	TraceFlag xflag.KeyUsage[bool] = `trace Log packet forwarding.`

	TunnelFlag xflag.KeyUsage[uint] = `t Tunnel unit number.`

	URIFlag xflag.KeyUsage[string] = `uri Comma separated URLs.`

	VerboseFlag xflag.KeyUsage[bool] = `v Verbose logging.`

	VpnFlag xflag.KeyUsage[string] = `vpn Named VPN. (default unnamed)`
)

func ConfigDirFile(sflag xflag.KeyUsage[string]) string {
	return filepath.Join(ConfigDirFlag.Value(), sflag.Value())
}

// [xdg.ConfigHome] or [fhs.Config] + GOES/vpn
var DefaultConfigDir = sync.OnceValue(func() string {
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
var DefaultStateDir = sync.OnceValue(func() string {
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

func defineAndParseFlags(args []string) error {
	qFlag := QuietFlag.Define(false)
	vFlag := VerboseFlag.Define(false)

	CertFlag.Define(DefaultCertFile)
	ConfigDirFlag.Define(DefaultConfigDir())
	RegFlag.Define(DefaultRegFile)
	SigFlag.Define(DefaultSigFile)
	VpnFlag.Define("")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if *qFlag {
		xlog.MuteErrata()
	} else if *vFlag {
		xlog.UnmuteInfo()
	}
	return nil
}
