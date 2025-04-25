// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

type BindClassFlag = xflag.Description[xdnsmessage.Class]
type BindTypeFlag = xflag.Description[xdnsmessage.Type]

// Features map by name these emulated BIND9 tools.
var Features = map[string]any{
	"dig":   Dig,
	"host":  Host,
	"named": Named,

	"named-checkzone": NCZ,
}
