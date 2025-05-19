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
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsync"
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

			"vcs.modified": RestShow,
			"vcs.revision": RestShow,
		},
	},
	"vpn": map[string]any{
		"approve":     RestAdmin,
		"certify":     RestCertify,
		"deny":        RestAdmin,
		"exchange":    Exchange,
		"get":         RestGet,
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

	year    = 365 * 24 * time.Hour
	longest = 10 * year
)

var (
	wg xsync.WaitGroup

	vpnCert         = "cert.pem"
	vpnConfig       = "config.yaml"
	vpnConfigDir    = "/etc/goes"
	vpnDataDir      = "/usr/share/goes"
	vpnDuration     = year
	vpnRegistry     = "registry.pem"
	vpnSerialNumber = int64(1)
	vpnSig          = "sig.pk8"
	vpnStateDir     = "/var/run/goes"

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

func defineCert() {
	xflag.Define(&vpnCert, "cert",
		"Certificate file name w/in config-dir.")
}

func defineConfig() {
	xflag.Define(&vpnConfig, "config",
		"Configuration file name w/in config-dir.")
}

// Default [xdg.ConfigHome] or [fhs.Config] + GOES/vpn
func defineConfigDir() {
	mn := xprogram.MainName()
	vpnConfigDir = filepath.Join(fhs.Config(), mn, "vpn")
	if !xprogram.IsKoApp() && os.Geteuid() != 0 {
		fhsVpnInfo, err := os.Stat(vpnConfigDir)
		if err != nil {
			fhsVpnInfo = nil
		}
		if home := xdg.ConfigHome(); len(home) > 0 {
			homeVpn := filepath.Join(home, mn, "vpn")
			if homeVpnInfo, err := os.Stat(homeVpn); err == nil {
				if homeVpnInfo.IsDir() {
					vpnConfigDir = homeVpn
				}
			} else if fhsVpnInfo == nil || !fhsVpnInfo.IsDir() {
				vpnConfigDir = homeVpn
			}
		}
	}
	xflag.Define(&vpnConfigDir, "config-dir", "Configuration directory.")
}

func defineCountry() {
	xflag.Define(&vpnCountry, "country", "")
}

// Default $KO_DATA_PATH; or [xdg.DataHome] or [fhs.Data] + GOES/vpn
func defineDataDir() {
	if s, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		vpnDataDir = s
	} else {
		mn := xprogram.MainName()
		vpnDataDir = filepath.Join(fhs.Data(), mn, "vpn")
		if os.Geteuid() != 0 {
			fhsVpnInfo, err := os.Stat(vpnDataDir)
			if err != nil {
				fhsVpnInfo = nil
			}
			if home := xdg.DataHome(); len(home) > 0 {
				homeVpn := filepath.Join(home, mn, "vpn")
				homeVpnInfo, err := os.Stat(homeVpn)
				if err == nil {
					if homeVpnInfo.IsDir() {
						vpnDataDir = homeVpn
					}
				} else if fhsVpnInfo == nil ||
					!fhsVpnInfo.IsDir() {
					vpnDataDir = homeVpn
				}
			}
		}
	}
	xflag.Define(&vpnDataDir, "data-dir", "Registry service directory.")
}

func defineDNS() {
	xflag.Define(&vpnDNS, "dns", "Comma separated domain names.")
}

func defineDuration() {
	xflag.Define(&vpnDuration, "duration", "e.g. 360s, 60m, or 1h.")
}

func defineEmail() {
	xflag.Define(&vpnEmail, "email", "Comma separated addresses.")
}

func defineListen(port uint16) {
	vpnListen = netip.AddrPortFrom(netip.IPv4Unspecified(), port)
	xflag.Define(&vpnListen, "listen", `
Service {addr}:{port}.
If “addr” is 0.0.0.0 or [::], listen on all ipv4 or ipv6
interface addresses.  If “port” is 0, allocate from system.`[1:])
}

func defineLocality() {
	xflag.Define(&vpnLocality, "locality", "aka. city.")
}

func defineName() {
	xflag.Define(&vpnName, "name", "VPN identfier.")
}

func definePublic() {
	vpnPublic = netip.AddrPortFrom(netip.IPv4Unspecified(), 0)
	xflag.Define(&vpnPublic, "public",
		"NAT'd listen {addr}:{port}. (0.0.0.0:0 ignored)")
}

func defineOrganization() {
	xflag.Define(&vpnOrganization, "organization", "aka. company")
}

func defineOrganizationalUnit() {
	xflag.Define(&vpnOrganizationalUnit, "organizational-unit",
		"aka. department.")
}

func definePostalCode() {
	xflag.Define(&vpnPostalCode, "postal-code", "aka. zip.")
}

func defineProvince() {
	xflag.Define(&vpnProvince, "province", "aka. state.")
}

func defineRegistry() {
	xflag.Define(&vpnRegistry, "registry",
		"Registry certificate file name w/in config-dir.")
}

func defineSerialNumber() {
	xflag.Define(&vpnSerialNumber, "serial-number", "Random if zero.")
}

func defineSig() {
	xflag.Define(&vpnSig, "sig", "Signature file name w/in config-dir.")
}

func defineStreet() {
	xflag.Define(&vpnStreet, "street", "")
}

func defineTunnel() {
	xflag.Define(&vpnTunnel, "t", "Tunnel unit number.")
}

func defineURI() {
	xflag.Define(&vpnURI, "uri", "Comma separated URLs.")
}

func enableQuiet() {
	xflag.Enable("q", "Quiet logging.", func() error {
		xlog.MuteErrata()
		return nil
	})
}

func enableTrace() {
	xflag.Enable("trace", "Log packet forwarding.", func() error {
		xlog.UnmuteTrace()
		return nil
	})
}

func enableVerbose() {
	xflag.Enable("v", "Verbose logging.", func() error {
		xlog.UnmuteInfo()
		return nil
	})
}

func defineVPN() {
	xflag.Define(&vpnVPN, "vpn", "Named VPN. (default unnamed)")
}

// Default [xdg.StateHome] or [fhs.State] + GOES/vpn
func defineStateDir() {
	mn := xprogram.MainName()
	vpnStateDir = filepath.Join(fhs.State(), mn, "vpn")
	if !xprogram.IsKoApp() && os.Geteuid() != 0 {
		fhsVpnInfo, err := os.Stat(vpnStateDir)
		if err != nil {
			fhsVpnInfo = nil
		}
		if home := xdg.StateHome(); len(home) > 0 {
			homeVpn := filepath.Join(home, mn, "vpn")
			if homeVpnInfo, err := os.Stat(homeVpn); err == nil {
				if homeVpnInfo.IsDir() {
					vpnStateDir = homeVpn
				}
			} else if fhsVpnInfo == nil || !fhsVpnInfo.IsDir() {
				vpnStateDir = homeVpn
			}
		}
	}
	xflag.Define(&vpnStateDir, "state-dir",
		"State directory to save approved client certificates.")
}
