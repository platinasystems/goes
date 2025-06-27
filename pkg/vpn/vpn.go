// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
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
			RestOpShowActive:     RestShow,
			RestOpShowAddress:    RestShow,
			RestOpShowAdmins:     RestShow,
			"certificate":        ShowCertificate,
			RestOpShowExchanges:  RestShow,
			RestOpShowHosts:      RestShow,
			RestOpShowPending:    RestShow,
			RestOpShowPrefix:     RestShow,
			"signature":          ShowSignature,
			RestOpShowStart:      RestShow,
			RestOpShowSubscriber: RestShow,
			RestOpShowTenant:     RestShow,
			RestOpShowVcs: map[string]any{
				RestOpShowVcsModified: RestShowVcs,
				RestOpShowVcsRevision: RestShowVcs,
			},
		},
	},
	"vpn": map[string]any{
		RestOpApprove:     RestAdmin,
		RestOpCertify:     RestCertify,
		RestOpDeny:        RestAdmin,
		"exchange":        Exchange,
		"get":             RestGet,
		"guest":           Guest,
		RestOpPing:        RestPing,
		RestOpReload:      RestReload,
		"registry":        Registry,
		RestOpSubscribe:   RestSubscribe,
		RestOpUnsubscribe: RestAdmin,
	},
}

const (
	defaultRegistryPort = 8003
	defaultServicePort  = 8003

	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	year    = 365 * 24 * time.Hour
	longest = 10 * year
)

var (
	errEOC = errors.New("end of channel")

	errNoService = errors.New("no service addr:port")

	errUnaddressed = errors.New("unaddressed")

	errUnassigned = errors.New("unanassigned")

	errUnestablished = errors.New("unestablished")

	errUnidentified = errors.New("unidentified")

	vpnCert = "cert.pem"

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

	wg xsync.WaitGroup
)

var DomainName = sync.OnceValues(func() (string, error) {
	s, err := HostFQDN()
	if err == nil {
		if i := strings.Index(s, "."); i > 0 {
			s = s[i+1:]
		} else {
			s = ""
		}
	}
	return s, err
})

var Host = sync.OnceValues(func() (string, error) {
	s, err := HostFQDN()
	if err == nil {
		if i := strings.Index(s, "."); i > 0 {
			s = s[:i]
		}
	}
	return s, err
})

var HostFQDN = sync.OnceValues(func() (string, error) {
	return os.Hostname()
})

func defineCert() {
	if s, err := Host(); err == nil {
		vpnCert = fmt.Sprint(s, ".pem")
	}
	xflag.Define(&vpnCert, "cert",
		"Certificate file name w/in config-dir.")
}

func defineConfig() {
	xflag.Define(&vpnConfig, "config",
		"Configuration file name w/in config-dir.")
}

// Default [xdg.ConfigHome] or [fhs.Config] +
// MAIN, if MAIN as "-vpn", or MAIN/vpn
func defineConfigDir() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.Config(), subDir)
	if xprogram.IsKoApp() {
		vpnConfigDir = fhsDir
	} else {
		xdgDir := filepath.Join(xdg.ConfigHome(), subDir)
		if os.Geteuid() == 0 {
			vpnConfigDir = prefDir(fhsDir, xdgDir)
		} else {
			vpnConfigDir = prefDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnConfigDir, "config-dir", "Configuration directory.")
}

func defineCountry() {
	xflag.Define(&vpnCountry, "country", "")
}

// Default $KO_DATA_PATH; or [xdg.DataHome] or [fhs.Data] +
// MAIN, if MAIN as "-vpn", or MAIN/vpn
func defineDataDir() {
	if s, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		vpnDataDir = s
	} else {
		subDir := mainSubDir()
		fhsDir := filepath.Join(fhs.Data(), subDir)
		xdgDir := filepath.Join(xdg.DataHome(), subDir)
		if os.Geteuid() == 0 {
			vpnDataDir = prefDir(fhsDir, xdgDir)
		} else {
			vpnDataDir = prefDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnDataDir, "data-dir", "Registry service directory.")
}

func defineDNS() {
	if s, err := HostFQDN(); err == nil && len(s) > 0 {
		vpnDNS = s
	}
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
	if s, err := Host(); err == nil {
		vpnName = s
	}
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

// Default [xdg.StateHome] or [fhs.State] + GOES/vpn
// MAIN, if MAIN as "-vpn", or MAIN/vpn
func defineStateDir() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.State(), subDir)
	if xprogram.IsKoApp() {
		vpnStateDir = fhsDir
	} else {
		xdgDir := filepath.Join(xdg.StateHome(), subDir)
		if os.Geteuid() == 0 {
			vpnStateDir = prefDir(fhsDir, xdgDir)
		} else {
			vpnStateDir = prefDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnStateDir, "state-dir",
		"State directory to save approved client certificates.")
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

func noServiceError(args ...any) error {
	return xerrors.Label(errNoService, args...)
}

func unaddressedError(args ...any) error {
	return xerrors.Label(errUnaddressed, args...)
}

func unassignedError(args ...any) error {
	return xerrors.Label(errUnassigned, args...)
}

func unestablishedError(args ...any) error {
	return xerrors.Label(errUnestablished, args...)
}

func mainSubDir() string {
	s := xprogram.MainName()
	if strings.Index(s, "-vpn") < 0 {
		s = filepath.Join(s, "vpn")
	}
	return s
}

func prefDir(primary string, alternates ...string) string {
	if fi, err := os.Stat(primary); err == nil && fi.IsDir() {
		return primary
	}
	for _, alt := range alternates {
		if fi, err := os.Stat(alt); err == nil && fi.IsDir() {
			return alt
		}
	}
	return primary
}
