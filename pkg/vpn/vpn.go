// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"fmt"
	"io/fs"
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
		"vpn": NewFeatures,
	},
	"show": map[string]any{
		"vpn": ShowFeatures,
	},
	"vpn": MainFeatures,
}

var MainFeatures = map[string]any{
	"approve":     RestAdmin,
	"certify":     RestCertify,
	"deny":        RestAdmin,
	"exchange":    Exchange,
	"get":         RestGet,
	"guest":       Guest,
	"nc":          NetCat,
	"ping":        Ping,
	"reload":      RestReload,
	"registry":    Registry,
	"subscribe":   RestSubscribe,
	"unsubscribe": RestAdmin,
}

var NewFeatures = map[string]any{
	"certificate": CreateCertificate,
	"signature":   NewEd25519,
}

var ShowFeatures = map[string]any{
	"address":     RestShow,
	"admins":      RestShow,
	"certificate": ShowCertificate,
	"exchanges":   RestShow,
	"guests":      RestShow,
	"hosts":       RestShow,
	"pending":     RestShow,
	"prefix":      RestShow,
	"registry": map[string]any{
		"vcs": RestShow,
	},
	"signature":  ShowSignature,
	"start":      RestShow,
	"status":     RestShow,
	"subscriber": RestShow,
}

const (
	FromTunCap = xnet.BatchCap
	FromVpnCap = xnet.BatchCap

	ToTunCap = xnet.BatchCap + 4
	ToVpnCap = xnet.BatchCap

	UpgradeExitCode = 127
)

const (
	defaultExchangePort = 8003
	defaultRegistryPort = 8003

	defaultRegistryFile = "registry.pem"

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

	vpnPort = uint16(defaultRegistryPort)

	vpnConfigDir, vpnDataDir, vpnStateDir string

	vpnAdminsFile    = "admins"
	vpnCertFile      = "cert.pem"
	vpnExchangesFile = "exchanges"
	vpnHostsFile     = "hosts"
	vpnRegistryFile  = defaultRegistryFile
	vpnSigFile       = "sig.pk8"

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
	vpnURI string

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

func defineAdminsFlag() {
	if s, ok := mainEnv("ADMINS"); ok {
		vpnAdminsFile = s
	}
	xflag.Define(&vpnAdminsFile, "admins", `
An optional file w/in current or config directory containing a newline
separated list of certificate common names that may administer subscriptions.
`[1:])
}

func defineCertFlag() {
	if s, ok := mainEnv("CERT"); ok {
		vpnCertFile = s
	} else if _, err := os.Stat(vpnCertFile); err == nil {
	} else if _, err = os.Stat(filepath.
		Join(vpnConfigDir, vpnCertFile)); err == nil {
	} else if s, err = Host(); err == nil {
		vpnCertFile = fmt.Sprint(s, ".pem")
	}
	xflag.Define(&vpnCertFile, "cert",
		"Certificate file w/in current or config directory.")
}

// Default { [xdg.ConfigHome] or [fhs.Config] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
var defineConfigFlag = sync.OnceFunc(func() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.Config(), subDir)
	xdgDir := filepath.Join(xdg.ConfigHome(), subDir)
	if s, ok := mainEnv("CONFIG"); ok {
		vpnConfigDir = s
	} else if xprogram.IsKoApp() {
		vpnConfigDir = fhsDir
	} else if os.Geteuid() == 0 {
		vpnConfigDir = bestDir(fhsDir, xdgDir)
	} else {
		vpnConfigDir = bestDir(xdgDir, fhsDir)
	}
	xflag.Define(&vpnConfigDir, "config", "Configuration directory.")
})

func defineCountryFlag() {
	xflag.Define(&vpnCountry, "country",
		"New certificate's country code.")
}

// Default $KO_DATA_PATH; or
// { [xdg.DataHome] or [fhs.Data] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
var defineDataFlag = sync.OnceFunc(func() {
	if s, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		vpnDataDir = s
	} else if s, ok = mainEnv("DATA"); ok {
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
	xflag.Define(&vpnDataDir, "data",
		"Registry directory containing alternate platforms.")
})

func defineDNSFlag() {
	if s, err := HostFQDN(); err == nil && len(s) > 0 {
		vpnDNS = s
	}
	xflag.Define(&vpnDNS, "dns",
		"New certificate's comma separated domain names.")
}

func defineDomainFlag() {
	xflag.Define(&vpnDomain, "domain", "Search domain suffix.")
}

func defineDurationFlag() {
	xflag.Define(&vpnDuration, "duration",
		"New certificate's life span, e.g. 360s, 60m, or 1h.")
}

func defineEmailFlag() {
	xflag.Define(&vpnEmail, "email",
		"New certificate's comma separated addresses.")
}

func defineExchangesFlag() {
	if s, ok := mainEnv("EXCHANGES"); ok {
		vpnExchangesFile = s
	}
	xflag.Define(&vpnExchangesFile, "exchanges", `
An optional file w/in current or config directory containing a newline
separated list of guest exhange assignment and exchange port numbers.
`[1:])
}

func defineHostsFlag() {
	if s, ok := mainEnv("HOSTS"); ok {
		vpnHostsFile = s
	}
	xflag.Define(&vpnHostsFile, "hosts", `
An optional file w/in current or config directory containing a newline
separated list of static address assignments in /ets/hosts format.
`[1:])
}

func defineLocalityFlag() {
	xflag.Define(&vpnLocality, "locality",
		"New certificate's city.")
}

func defineNameFlag() {
	if s, err := Host(); err == nil {
		vpnName = s
	}
	xflag.Define(&vpnName, "name", "VPN identfier.")
}

func defineOrganizationFlag() {
	xflag.Define(&vpnOrganization, "organization",
		"New certificate's company")
}

func defineOrganizationalUnitFlag() {
	xflag.Define(&vpnOrganizationalUnit, "organizational-unit",
		"New certificate's department.")
}

func definePortFlag() {
	xflag.Define(&vpnPort, "port", "REST listener.")
}

func definePostalCodeFlag() {
	xflag.Define(&vpnPostalCode, "postal-code",
		"New certificate's zip code.")
}

func definePrefixFlag() {
	vpnPrefix = netip.MustParsePrefix("fc00:1234::/64")
	xflag.Define(&vpnPrefix, "prefix", "Network prefix.")
}

func defineProvinceFlag() {
	xflag.Define(&vpnProvince, "province",
		"New certificate's state.")
}

func defineRegistryFlag() {
	if s, ok := mainEnv("REGISTRY"); ok {
		vpnRegistryFile = s
	}
	xflag.Define(&vpnRegistryFile, "registry",
		"Registry certificate file w/in current or config directory.")
}

func defineSerialNumberFlag() {
	xflag.Define(&vpnSerialNumber, "serial-number",
		"New certificate's identifier, random if zero.")
}

func defineSigFlag() {
	if s, ok := mainEnv("SIG"); ok {
		vpnSigFile = s
	}
	xflag.Define(&vpnSigFile, "sig",
		"Signature file w/in current or config directory.")
}

// Default { [xdg.StateHome] or [fhs.State] } + MAIN,
// if MAIN as "-vpn", or MAIN/vpn
func defineStateFlag() {
	subDir := mainSubDir()
	fhsDir := filepath.Join(fhs.State(), subDir)
	xdgDir := filepath.Join(xdg.StateHome(), subDir)
	if s, ok := mainEnv("STATE"); ok {
		vpnStateDir = s
	} else if xprogram.IsKoApp() {
		vpnStateDir = fhsDir
	} else if os.Geteuid() == 0 {
		vpnStateDir = bestDir(fhsDir, xdgDir)
	} else {
		vpnStateDir = bestDir(xdgDir, fhsDir)
	}
	xflag.Define(&vpnStateDir, "state",
		"Registry directory to save approved subscriber certificates.")
}

func defineStreetFlag() {
	xflag.Define(&vpnStreet, "street", "")
}

func defineTunnelFlag() {
	xflag.Define(&vpnTunnel, "t", "Tunnel unit number.")
}

func defineURIFlag() {
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

func mainEnv(suffix string) (string, bool) {
	s := fmt.Sprint(mainEnvPrefix(), "_", suffix)
	return os.LookupEnv(s)
}

var mainEnvPrefix = sync.OnceValue(func() string {
	s := xprogram.MainName()
	if strings.Index(s, "-vpn") < 0 {
		s += "_VPN"
	}
	s = strings.ToUpper(mainSubDir())
	s = strings.Replace(s, "-", "_", -1)
	return s
})

var mainSubDir = sync.OnceValue(func() string {
	s := xprogram.MainName()
	if strings.Index(s, "-vpn") < 0 {
		s = filepath.Join(s, "vpn")
	}
	return s
})

func newGreeting(now int64) *xnet.Msg {
	var err error

	m := mp.Get()
	m.Data = m.Data[:0]

	if now == 0 {
		now = time.Now().UnixMicro()
	}

	m.Data, err = xnet.ByteOrderAppend(m.Data, vpnStart)
	if err != nil {
		xlog.Errata.Print(err)
		mp.Put(m)
		return nil
	}
	m.Data, err = xnet.ByteOrderAppend(m.Data, now)
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

func unmap4in6(ap netip.AddrPort) netip.AddrPort {
	if addr := ap.Addr(); addr.Is4In6() {
		ap = netip.AddrPortFrom(addr.Unmap(), ap.Port())
	}
	return ap
}

func vpnConfigFile(s string) string {
	if s != "-" && strings.IndexRune(s, filepath.Separator) < 0 {
		if _, err := os.Stat(s); errors.Is(err, fs.ErrNotExist) {
			s = filepath.Join(vpnConfigDir, s)
		}
	}
	return s
}

var vpnAdminsPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnAdminsFile)
})

var vpnCertPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnCertFile)
})

var vpnExchangesPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnExchangesFile)
})

var vpnHostsPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnHostsFile)
})

var vpnRegistryPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnRegistryFile)
})

var vpnSigPath = sync.OnceValue(func() string {
	return vpnConfigFile(vpnSigFile)
})
