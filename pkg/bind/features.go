// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"github.com/platinasystems/goes/v2/pkg/bind/dig"
	"github.com/platinasystems/goes/v2/pkg/bind/host"
	"github.com/platinasystems/goes/v2/pkg/bind/named"
	checkzone "github.com/platinasystems/goes/v2/pkg/bind/named-checkzone"
)

// Features map by name these emulated BIND9 tools.
var Features = map[string]any{
	"dig":   dig.DiG,
	"host":  host.Host,
	"named": named.Named,

	"named-checkzone": checkzone.NamedCheckZone,
}
