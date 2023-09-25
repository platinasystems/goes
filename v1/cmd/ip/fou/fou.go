// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fou

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/ip/fou/add"
	"github.com/platinasystems/goes/cmd/ip/fou/delete"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME:  "fou",
	USAGE: "ip foo COMMAND [ OPTION... ]",
	APROPOS: lang.Alt{
		lang.EnUS: "Foo-over-UDP receive port configuration",
	},
	MAN: lang.Alt{
		lang.EnUS: `
COMMANDS
	add
	del[ete]

EXAMPLES
	Configure a FOU receive port for GRE bound to 7777
		# ip fou add port 7777 ipproto 47

	Configure a FOU receive port for IPIP bound to 8888
		# ip fou add port 8888 ipproto 4

	Configure a GUE receive port bound to 9999
		# ip fou add port 9999 gue

	Delete the GUE receive port bound to 9999
		# ip fou del port 9999

SEE ALSO
	ip fou man COMMAND || ip fou COMMAND -man
	man ip || ip -man`,
	},
	ByName: map[string]cmd.Cmd{
		"add":    add.Command{},
		"delete": delete.Command{},
	},
}
