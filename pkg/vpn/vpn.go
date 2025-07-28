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
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
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
			"exchanges":   RestShow,
			"guests":      RestShow,
			"hosts":       RestShow,
			"pending":     RestShow,
			"prefix":      RestShow,
			"signature":   ShowSignature,
			"start":       RestShow,
			"status":      RestShow,
			"subscriber":  RestShow,
			"vcs":         RestShow,
		},
	},
	"vpn": map[string]any{
		"approve":     RestAdmin,
		"certify":     RestCertify,
		"deny":        RestAdmin,
		"exchange":    Exchange,
		"get":         RestGet,
		"guest":       Guest,
		"nc":          NetCat,
		"reload":      RestReload,
		"registry":    Registry,
		"subscribe":   RestSubscribe,
		"unsubscribe": RestAdmin,
	},
}

const (
	SizeofFromTunC = 1
	SizeofFromVpnC = 1

	SizeofToTunC = 1
	SizeofToVpnC = 1
)

const (
	defaultExchangePort = 8003
	defaultRegistryPort = 8003

	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

	year    = 365 * 24 * time.Hour
	longest = 10 * year
)

var Start time.Time // Registry start time

var (
	errEOC = errors.New("end of channel")

	errNoService = errors.New("no service addr:port")

	errUnaddressed = errors.New("unaddressed")

	errUnassigned = errors.New("unanassigned")

	errUnestablished = errors.New("unestablished")

	errUnidentified = errors.New("unidentified")

	mp = xnet.NewMsgPool(netph.ETHMTU)

	vpnDomain       = ".example.platina.io."
	vpnDuration     = year
	vpnSerialNumber = int64(1)
	vpnPrefix       netip.Prefix

	vpnExchangePort = uint16(defaultExchangePort)
	vpnRegistryPort = uint16(defaultRegistryPort)

	vpnAdminsFile,
	vpnConfigDir,
	vpnCertFile,
	vpnCountry,
	vpnDataDir,
	vpnEmail,
	vpnDNS,
	vpnHostsFile,
	vpnLocality,
	vpnName,
	vpnOrganization,
	vpnOrganizationalUnit,
	vpnPostalCode,
	vpnProvince,
	vpnRegistryFile,
	vpnSigFile,
	vpnStateDir,
	vpnStreet,
	vpnURI,
	vpnViaFileName,
	vpnVPN string

	vpnTunnel = -1

	// Unix Micro start of registry
	vpnStart int64

	wg xsync.WaitGroup
)

type empty = struct{}

var novalue, done empty

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

func bestDir(primary string, alternates ...string) string {
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

func defineAdmins() {
	vpnAdminsFile = filepath.Join(vpnConfigDir, "admins")
	xflag.Define(&vpnAdminsFile, "admins", `
An optional file containing a newline separated list of certificate
common names that may administer subscriptions.
`[1:])
}

func defineCert() {
	fn := "cert.pem"
	if s, err := Host(); err == nil {
		fn = fmt.Sprint(s, ".pem")
	}
	vpnCertFile = filepath.Join(vpnConfigDir, fn)
	xflag.Define(&vpnCertFile, "cert", "Certificate file.")
}

// Default { [xdg.ConfigHome] or [fhs.Config] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
func defineConfig() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.Config(), subDir)
	if xprogram.IsKoApp() {
		vpnConfigDir = fhsDir
	} else {
		xdgDir := filepath.Join(xdg.ConfigHome(), subDir)
		if os.Geteuid() == 0 {
			vpnConfigDir = bestDir(fhsDir, xdgDir)
		} else {
			vpnConfigDir = bestDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnConfigDir, "config", "Configuration directory.")
}

func defineCountry() {
	xflag.Define(&vpnCountry, "country", "")
}

// Default $KO_DATA_PATH; or
// { [xdg.DataHome] or [fhs.Data] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
func defineData() {
	if s, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		vpnDataDir = s
	} else {
		subDir := mainSubDir()
		fhsDir := filepath.Join(fhs.Data(), subDir)
		xdgDir := filepath.Join(xdg.DataHome(), subDir)
		if os.Geteuid() == 0 {
			vpnDataDir = bestDir(fhsDir, xdgDir)
		} else {
			vpnDataDir = bestDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnDataDir, "data", "Registry service directory.")
}

func defineDNS() {
	if s, err := HostFQDN(); err == nil && len(s) > 0 {
		vpnDNS = s
	}
	xflag.Define(&vpnDNS, "dns", "Comma separated domain names.")
}

func defineDomain() {
	xflag.Define(&vpnDomain, "domain", "Search domain suffix.")
}

func defineDuration() {
	xflag.Define(&vpnDuration, "duration", "e.g. 360s, 60m, or 1h.")
}

func defineEmail() {
	xflag.Define(&vpnEmail, "email", "Comma separated addresses.")
}

func defineHosts() {
	vpnHostsFile = filepath.Join(vpnConfigDir, "hosts")
	xflag.Define(&vpnHostsFile, "hosts",
		"Static address assignments in /ets/hosts format.")
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

func defineExchangePort() {
	xflag.Define(&vpnExchangePort, "exchange-port", "Packet forwarding.")
}

func defineRegistryPort() {
	xflag.Define(&vpnRegistryPort, "registry-port", "REST.")
}

func definePrefix() {
	vpnPrefix = netip.MustParsePrefix("fc00:1234::/64")
	xflag.Define(&vpnPrefix, "prefix", "Network prefix.")
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
	vpnRegistryFile = filepath.Join(vpnConfigDir, "registry.pem")
	xflag.Define(&vpnRegistryFile, "registry",
		"Registry certificate file.")
}

func defineSerialNumber() {
	xflag.Define(&vpnSerialNumber, "serial-number", "Random if zero.")
}

func defineSig() {
	vpnSigFile = filepath.Join(vpnConfigDir, "sig.pk8")
	xflag.Define(&vpnSigFile, "sig", "Signature file.")
}

// Default { [xdg.StateHome] or [fhs.State] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
func defineState() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.State(), subDir)
	if xprogram.IsKoApp() {
		vpnStateDir = fhsDir
	} else {
		xdgDir := filepath.Join(xdg.StateHome(), subDir)
		if os.Geteuid() == 0 {
			vpnStateDir = bestDir(fhsDir, xdgDir)
		} else {
			vpnStateDir = bestDir(xdgDir, fhsDir)
		}
	}
	xflag.Define(&vpnStateDir, "state",
		"State directory to save approved client certificates.")
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

func defineVia() {
	vpnViaFileName = filepath.Join(vpnConfigDir, "via")
	xflag.Define(&vpnViaFileName, "via",
		"Subscriber exchange precedence file.")
}

func defineVPN() {
	xflag.Define(&vpnVPN, "vpn", "Named VPN. (default unnamed)")
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

func newGreeting(now int64) *xnet.Msg {
	var err error

	m := mp.Get()
	m.Data = m.Data[:0]

	if now == 0 {
		now = time.Now().UnixMicro()
	}

	m.Data, err = xnet.Attach(m.Data, vpnStart)
	if err != nil {
		xlog.Errata.Print(err)
		mp.Put(m)
		return nil
	}
	m.Data, err = xnet.Attach(m.Data, now)
	if err != nil {
		xlog.Errata.Print(err)
		mp.Put(m)
		return nil
	}
	sig := sign(m.Data)
	m.Data = append(m.Data, sig...)
	m.Data = append(m.Data, MyLabel...)
	return m
}
