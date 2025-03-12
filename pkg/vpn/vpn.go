// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"crypto/rand"
	"errors"
	"flag"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	DefaultCertFile   = "cert.pem"
	DefaultConfigFile = "config.yaml"
	DefaultRegFile    = "registry.pem"
	DefaultSigFile    = "sig.pk8"
)

const (
	NameCertFlag      = "cert"
	NameConfigFlag    = "config"
	NameConfigDirFlag = "config-dir"
	NameNatFlag       = "nat"
	NameQuietFlag     = "q"
	NameRegFlag       = "reg"
	NameSigFlag       = "sig"
	NameStateDirFlag  = "state-dir"
	NameTunnelFlag    = "t"
	NameVerboseFlag   = "v"
	NameVpnFlag       = "vpn"
)

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

var ErrLLAddrUnderrun = errors.New("link-local address underrun")

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

func RandLinkLocalAddr() (lladdr netip.Addr, err error) {
	var a [netph.IPv6len]byte
	a[0] = 0xfe
	a[1] = 0x80
	n, err := rand.Read(a[8:])
	if err != nil {
	} else if n != len(a[8:]) {
		err = ErrLLAddrUnderrun
	} else {
		lladdr = netip.AddrFrom16(a)
		verbose.Println("link-local address:", lladdr)
	}
	return
}

func ValueOfStringFlag(name string) (s string) {
	if f := flag.Lookup(name); f != nil {
		s = f.Value.String()
	}
	return
}

const (
	oAppend = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	oCreate = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)

	errata  = xlog.Unmute(mutable)
	verbose = xlog.Mute(mutable)

	udpRxTrace = xlog.Mute(mutable)
	udpTxTrace = xlog.Mute(mutable)

	goRoutineTrace = xlog.Mute(mutable)
)

func defineAndParseFlags(args []string) error {
	qFlag := flag.Bool(NameQuietFlag, false, "Quiet logging.")
	vFlag := flag.Bool(NameVerboseFlag, false, "Verbose logging.")

	flag.String(NameCertFlag, DefaultCertFile,
		"Certificate file name w/in config-dir.")
	flag.String(NameConfigDirFlag, DefaultConfigDir(),
		"Configuration directory.")
	flag.String(NameRegFlag, DefaultRegFile,
		"Registry certificate file name w/in config-dir.")
	flag.String(NameSigFlag, DefaultSigFile,
		"Signature file name w/in config-dir.")
	flag.String(NameVpnFlag, "", "Named VPN. (default unnamed)")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if *qFlag {
		errata = xlog.Mute(errata)
	} else if *vFlag {
		verbose = xlog.Unmute(verbose)
	}
	return nil
}

func pathCertFile() string {
	return filepath.Join(pathConfigDir(), ValueOfStringFlag(NameCertFlag))
}

func pathConfigFile() string {
	return filepath.Join(pathConfigDir(), ValueOfStringFlag(NameConfigFlag))
}

func pathConfigDir() string {
	return ValueOfStringFlag(NameConfigDirFlag)
}

func pathRegFile() string {
	return filepath.Join(pathConfigDir(), ValueOfStringFlag(NameRegFlag))
}

func pathSigFile() string {
	return filepath.Join(pathConfigDir(), ValueOfStringFlag(NameSigFlag))
}
