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
	"github.com/platinasystems/goes/v2/pkg/xos"
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
	"im":          InstantMessaging,
	"nc":          NetCat,
	"ping":        Ping,
	"reload":      RestReload,
	"registry":    Registry,
	"subscribe":   RestSubscribe,
	"unsubscribe": RestAdmin,
	"update":      RestUpdate,
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

	UpgradeExitCode = xos.EX_TEMPFAIL
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

	VpnConfigDir, VpnDataDir, VpnStateDir string

	VpnAdminsFile    = "admins"
	VpnCertFile      = "cert.pem"
	VpnExchangesFile = "exchanges"
	VpnHostsFile     = "hosts"
	VpnRegistryFile  = defaultRegistryFile
	VpnSigFile       = "sig.pk8"

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

// BestDir returns the primary directory name if it exists,
// or the first existing alternate directory,
// or the primary if none exist.
func BestDir(primary string, alternates ...string) string {
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

// Default:
//   - Value of existing [MainEnv]("ADMINS")
//   - “admins” if that exist in current or [VpnConfigDir]
func DefineAdminsFlag() {
	if s, ok := MainEnv("ADMINS"); ok {
		VpnAdminsFile = s
	}
	xflag.Define(&VpnAdminsFile, "admins", `
An optional file w/in current or config directory containing a newline
separated list of certificate common names that may administer subscriptions.
`[1:])
}

// Default:
//   - Value of existing [MainEnv]("CERT")
//   - “cert.pem” if that exist in current or [VpnConfigDir]
//   - otherwise, [Host] + ".pem“
func DefineCertFlag() {
	if s, ok := MainEnv("CERT"); ok {
		VpnCertFile = s
	} else if _, err := os.Stat(VpnCertFile); err == nil {
	} else if _, err = os.Stat(filepath.
		Join(VpnConfigDir, VpnCertFile)); err == nil {
	} else if s, err = Host(); err == nil {
		VpnCertFile = fmt.Sprint(s, ".pem")
	}
	xflag.Define(&VpnCertFile, "cert",
		"Certificate file w/in current or config directory.")
}

var FhsConfigMainDir = func() string {
	return filepath.Join(fhs.Config(), MainSubDir())
}

var XdgConfigMainDir = func() string {
	return filepath.Join(xdg.ConfigHome(), MainSubDir())
}

var DefineConfigFlag = sync.OnceFunc(func() {
	if s, ok := MainEnv("CONFIG"); ok {
		VpnConfigDir = s
	} else if xprogram.IsKoApp() {
		VpnConfigDir = FhsConfigMainDir()
	} else if os.Geteuid() == 0 {
		VpnConfigDir = BestDir(FhsConfigMainDir(), XdgConfigMainDir())
	} else {
		VpnConfigDir = BestDir(XdgConfigMainDir(), FhsConfigMainDir())
	}
	xflag.Define(&VpnConfigDir, "config", "Configuration directory.")
})

func DefineCountryFlag() {
	xflag.Define(&vpnCountry, "country",
		"New certificate's country code.")
}

var FhsDataMainDir = func() string {
	return filepath.Join(fhs.Data(), MainSubDir())
}

var XdgDataMainDir = func() string {
	return filepath.Join(xdg.DataHome(), MainSubDir())
}

var DefineDataFlag = sync.OnceFunc(func() {
	if s, ok := os.LookupEnv("KO_DATA_PATH"); ok {
		VpnDataDir = s
	} else if s, ok = MainEnv("DATA"); ok {
		VpnDataDir = s
	} else if os.Geteuid() == 0 {
		VpnDataDir = BestDir(FhsDataMainDir(), XdgDataMainDir())
	} else {
		VpnDataDir = BestDir(XdgDataMainDir(), FhsDataMainDir())
	}
	xflag.Define(&VpnDataDir, "data",
		"Registry directory containing alternate platforms.")
})

func DefineDNSFlag() {
	if s, err := HostFQDN(); err == nil && len(s) > 0 {
		vpnDNS = s
	}
	xflag.Define(&vpnDNS, "dns",
		"New certificate's comma separated domain names.")
}

func DefineDomainFlag() {
	xflag.Define(&vpnDomain, "domain", "Search domain suffix.")
}

func DefineDurationFlag() {
	xflag.Define(&vpnDuration, "duration",
		"New certificate's life span, e.g. 360s, 60m, or 1h.")
}

func DefineEmailFlag() {
	xflag.Define(&vpnEmail, "email",
		"New certificate's comma separated addresses.")
}

var DefineExchangesFlag = func() {
	if s, ok := MainEnv("EXCHANGES"); ok {
		VpnExchangesFile = s
	}
	xflag.Define(&VpnExchangesFile, "exchanges", `
An optional file w/in current or config directory containing a newline
separated list of guest exhange assignment and exchange port numbers.
`[1:])
}

// Default:
//   - Value of existing [MainEnv]("HOSTS")
//   - othersise, "hosts" in current or [VpnConfigDir]
var DefineHostsFlag = func() {
	if s, ok := MainEnv("HOSTS"); ok {
		VpnHostsFile = s
	}
	xflag.Define(&VpnHostsFile, "hosts", `
An optional file w/in current or config directory containing a newline
separated list of static address assignments in /ets/hosts format.
`[1:])
}

func DefineLocalityFlag() {
	xflag.Define(&vpnLocality, "locality",
		"New certificate's city.")
}

func DefineNameFlag() {
	if s, err := Host(); err == nil {
		vpnName = s
	}
	xflag.Define(&vpnName, "name", "VPN identfier.")
}

func DefineOrganizationFlag() {
	xflag.Define(&vpnOrganization, "organization",
		"New certificate's company")
}

func DefineOrganizationalUnitFlag() {
	xflag.Define(&vpnOrganizationalUnit, "organizational-unit",
		"New certificate's department.")
}

func DefinePortFlag() {
	xflag.Define(&vpnPort, "port", "REST listener.")
}

func DefinePostalCodeFlag() {
	xflag.Define(&vpnPostalCode, "postal-code",
		"New certificate's zip code.")
}

func DefinePrefixFlag() {
	vpnPrefix = netip.MustParsePrefix("fc00:1234::/64")
	xflag.Define(&vpnPrefix, "prefix", "Network prefix.")
}

func DefineProvinceFlag() {
	xflag.Define(&vpnProvince, "province",
		"New certificate's state.")
}

func DefineQuietFlag() {
	xflag.Enable("q", "Quiet logging.", func() error {
		xlog.MuteErrata()
		return nil
	})
}

var DefineRegistryFlag = func() {
	if s, ok := MainEnv("REGISTRY"); ok {
		VpnRegistryFile = s
	}
	xflag.Define(&VpnRegistryFile, "registry",
		"Registry certificate file w/in current or config directory.")
}

func DefineSerialNumberFlag() {
	xflag.Define(&vpnSerialNumber, "serial-number",
		"New certificate's identifier, random if zero.")
}

var DefineSigFlag = func() {
	if s, ok := MainEnv("SIG"); ok {
		VpnSigFile = s
	}
	xflag.Define(&VpnSigFile, "sig",
		"Signature file w/in current or config directory.")
}

var FhsStateMainDir = func() string {
	return filepath.Join(fhs.State(), MainSubDir())
}

var XdgStateMainDir = func() string {
	return filepath.Join(xdg.StateHome(), MainSubDir())
}

var DefineStateFlag = func() {
	if s, ok := MainEnv("STATE"); ok {
		VpnStateDir = s
	} else if xprogram.IsKoApp() {
		VpnStateDir = FhsStateMainDir()
	} else if os.Geteuid() == 0 {
		VpnStateDir = BestDir(FhsStateMainDir(), XdgStateMainDir())
	} else {
		VpnStateDir = BestDir(XdgStateMainDir(), FhsStateMainDir())
	}
	xflag.Define(&VpnStateDir, "state",
		"Registry directory to save approved subscriber certificates.")
}

func DefineStreetFlag() {
	xflag.Define(&vpnStreet, "street", "")
}

func DefineTunnelFlag() {
	xflag.Define(&vpnTunnel, "t", "Tunnel unit number.")
}

func DefineURIFlag() {
	xflag.Define(&vpnURI, "uri", "Comma separated URLs.")
}

func DefineTraceFlag() {
	xflag.Enable("trace", "Log packet forwarding.", func() error {
		xlog.UnmuteTrace()
		return nil
	})
}

func DefineVerboseFlag() {
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

// MainEnv returns the [os.LookupEnv] of [MainEnvPrefix] with the given suffix.
func MainEnv(suffix string) (string, bool) {
	return os.LookupEnv(fmt.Sprint(MainEnvPrefix(), suffix))
}

// MainEnvPrefix returns the [xprogram.MainName] as an uppercase, environment
// keyword prefix.  If not [MainHasVpn], this appends that keyword with "_VPN".
//
// So, both “goes” and “goes-vpn” mains return “GOES_VPN_” prefix.
var MainEnvPrefix = sync.OnceValue(func() string {
	var sb strings.Builder
	sb.WriteString(ToUpperEnvName(xprogram.MainName()))
	if !MainHasVpn() {
		sb.WriteString("_VPN")
	}
	sb.WriteRune('_')
	return sb.String()
})

func ToUpperEnvName(s string) string {
	return strings.Replace(strings.ToUpper(s), "-", "_", -1)
}

// MainHasVpn returns true if [xprogram.MainName] has "-vpn".
var MainHasVpn = sync.OnceValue(func() bool {
	return strings.Index(xprogram.MainName(), "-vpn") > 0
})

// MainSubDir returns [xprogram.MainName] if it has "-vpn";
// otherwise, it filepath joins that with "vpn".
//
// So, a “goes” main returns “goes/vpn”
// and the “goes-vpn” main returns “goes-vpn”.
var MainSubDir = sync.OnceValue(func() string {
	s := xprogram.MainName()
	if !MainHasVpn() {
		s = filepath.Join(s, "vpn")
	}
	return s
})

func NewGreeting(now int64) *xnet.Msg {
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
			s = filepath.Join(VpnConfigDir, s)
		}
	}
	return s
}

var vpnAdminsPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnAdminsFile)
})

var vpnCertPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnCertFile)
})

var vpnExchangesPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnExchangesFile)
})

var vpnHostsPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnHostsFile)
})

var vpnRegistryPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnRegistryFile)
})

var vpnSigPath = sync.OnceValue(func() string {
	return vpnConfigFile(VpnSigFile)
})
