// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neighbor

const Man = `
DESCRIPTION
	The ip neigh command manipulates neighbor objects that establish bind‐
	ings between protocol addresses and link layer addresses for hosts
	sharing the same link.  Neighbour entries are organized into tables.
	The IPv4 neighbor table is also known by another name - the ARP table.

	The corresponding commands display neighbor bindings and their proper‐
	ties, add new neighbor entries and delete old ones.

	ip neighbor add
		add a new neighbor entry

	ip neighbor change
		change an existing entry

	ip neighbor replace
		add a new entry or change an existing one

		These commands create new neighbor records or update existing
		ones.

		to ADDRESS (default)
			the protocol address of the neighbor. It is either an
			IPv4 or IPv6 address.

		dev NAME
			the interface to which this neighbor is attached.

		lladdr LLADDRESS
			the link layer address of the neighbor.  LLADDRESS can
			also be null.

		nud STATE
			the state of the neighbor entry.  nud is an
			abbreviation for 'Neighbour Unreachability Detection'.
			The state can take one of the following values:

			permanent
				the neighbor entry is valid forever and can be
				only be removed administratively.

			noarp	the neighbor entry is valid.
				No attempts to validate this entry will be made
				but it can be removed when its lifetime
				expires.

			reachable
				the neighbor entry is valid until the
				reachability timeout expires.

			stale	the neighbor entry is valid but suspicious.
				This option to ip neigh does not change the
				neighbor state if it was valid and the address
				is not changed by this command.

			none	this is a pseudo state used when initially
				creating a neighbor entry or after trying to
				remove it before it becomes free to do so.

			incomplete
				the neighbor entry has not (yet) been vali‐
				dated/resolved.

			delay	neighbor entry validation is currently delayed.

			probe	neighbor is being probed.

			failed	max number of probes exceeded without success,
				neighbor validation has ultimately failed.

	ip neighbor delete
		delete a neighbor entry

		The arguments are the same as with ip neigh add, except that
		lladdr and nud are ignored.

		Warning: Attempts to delete or manually change a noarp entry
		created by the kernel may result in unpredictable behaviour.
		Particularly, the kernel may try to resolve this address even
		on a NOARP interface or if the address is multicast or
		broadcast.

	ip neighbor show
		list neighbor entries

		to ADDRESS (default)
			the prefix selecting the neighbors to list.

		dev NAME
			only list the neighbors attached to this device.

		vrf NAME
			only list the neighbors for given VRF.

		proxy  list neighbor proxies.

		unused only list neighbors which are not currently in use.

		nud STATE
			only list neighbor entries in this state.
			NUD_STATE takes values listed below or the special
			value all which means all states. This option may occur
			more than once.  If this option is absent, ip lists all
			entries except for none and noarp.

	ip neighbor flush
		flush neighbor entries
		This command has the same arguments as show.  The differences
		are that it does not run when no arguments are given, and that
		the default neighbor states to be flushed do not include
		permanent and noarp.

		With the -statistics option, the command becomes verbose. It
		prints out the number of deleted neighbors and the number of
		rounds made to flush the neighbor table. If the option is
		given twice, ip neigh flush also dumps all the deleted
		neighbors.

EXAMPLES
	ip neighbor
		Shows the current neighbor table in kernel.

	ip neighbor flush dev eth0
		Removes entries in the neighbor table on device eth0.

SEE ALSO
	man ip || ip -man

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
