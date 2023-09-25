// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

var (
	DefaultStateDir = sync.OnceValue(func() string {
		const tlsx = "tlsx"
		home := xdg.StateHome()
		if base := program.Base(); base != tlsx {
			home = filepath.Join(home, base)
		}
		return filepath.Join(home, tlsx)
	})
	StateDir = DefaultStateDir
)

func MkStateDir() error {
	return xdg.MkPath(StateDir())
}

var CertFileName = sync.OnceValue(func() string {
	return stateFileName("TLSX_CERT_FILE", "cert.pem")
})

var KeyFileName = sync.OnceValue(func() string {
	return stateFileName("TLSX_KEY_FILE", "key.pem")
})

var SubscribersFileName = sync.OnceValue(func() string {
	return stateFileName("TLSX_SUBSCRIBERS_FILE", "subscribers.pem")
})

var SubscriptionsFileName = sync.OnceValue(func() string {
	return stateFileName("TLSX_SUBSCRIPTIONS_FILE", "subscriptions.pem")
})

func Filenames() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, CertFileName())
	fmt.Fprintln(&sb, KeyFileName())
	fmt.Fprintln(&sb, SubscribersFileName())
	fmt.Fprintln(&sb, SubscriptionsFileName())
	return sb.String()
}

func stateFileName(env, base string) (fn string) {
	if fn = os.Getenv(env); len(fn) == 0 {
		fn = filepath.Join(StateDir(), base)
	}
	return
}
