// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package add

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	addtype "github.com/platinasystems/goes/cmd/ip/link/add/type"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "add",
	USAGE: `
ip link add type TYPE [[ name ] NAME ] [ OPTION ]... [ ARGS ]...`,
	APROPOS: lang.Alt{
		lang.EnUS: "add virtual link",
	},
	MAN: lang.Alt{
		lang.EnUS: `
OPTIONS
	address LLADDRESS
	broadcast LLADDRESS
		initial the station and broadcast address of the interface.

	index NUMBER
		the desired index of the new virtual device.
		The link creation fails, if the index is busy.

	link DEVICE
		physical device associated with new virtual link

	mtu NUMBER
		initial maximum transmission unit of the device.

	[ name ] NAME
		name of new virtual link
		The kernel will allocate if not given and will fail if
		already in use.

	numrxqueues QUEUE_COUNT
		initial number of receive queues for new device.

	numtxqueues QUEUE_COUNT
		initial number of transmit queues for new device.

	{ txqueuelen | txqlen } PACKETS
		initial transmit queue depth

SEE ALSO
	ip link add man type || ip link add type TYPE -man
	ip link add type man TYPE || ip link add type TYPE -man
	man ip || ip -man`,
	},
	ByName: map[string]cmd.Cmd{
		"type": addtype.Goes,
	},
}
