// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
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
	NameCertFlag      = "cert"
	NameConfigFlag    = "config"
	NameConfigDirFlag = "config-dir"
	NameListenFlag    = "listen"
	NamePublicFlag    = "public"
	NameQuietFlag     = "q"
	NameRegFlag       = "reg"
	NameSigFlag       = "sig"
	NameStateDirFlag  = "state-dir"
	NameTraceFlag     = "trace"
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
		xlog.MuteErrata()
	} else if *vFlag {
		xlog.UnmuteInfo()
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
