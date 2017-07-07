// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ip

const Man = `
OPTIONS
	-human, -human-readable
		Output statistics with human readable suffix.

	-s, -stats, -statistics
		Output more information. If the option appears twice or more,
		the amount of information increases.  As a rule, the
		information is statistics or some time values.

	-d, -details
		Output more detailed information.

	-l <COUNT>
		Specify maximum number of loops the 'ip address flush' logic
		will attempt before giving up. The default is 10.  Zero (0)
		means loop until all addresses are removed.

	-f <FAMILY>
		Specifies the protocol family to use. The protocol family iden‐
		tifier can be one of inet, inet6, bridge, mpls or link.  If
		this option is not present, the protocol family is guessed from
		other arguments. If the rest of the command line does not give
		enough information to guess the family, ip falls back to the
		default one, usually inet or any.  link is a special family
		identifier meaning that no networking protocol is involved.

	-4	shortcut for -family inet.
	-6	shortcut for -family inet6.
	-B	shortcut for -family bridge.
	-M	shortcut for -family mpls.
	-0	shortcut for -family link.

	-o, -oneline
		Output each record on a single line, replacing line feeds with
		the '\' character. This is convenient when you want to count
		records with wc(1) or to grep(1) the output.

	-r, -resolve
		Use the system's name resolver to print DNS names instead of
		host addresses.

	-n <NETNS>
		Switches ip to the specified network namespace NETNS.  Actually
		it just simplifies executing of:
			ip netns exec NETNS ip ...
		with:
			ip -n NETNS ...

	-a, -all
		Executes specified command over all objects, it depends if com‐
		mand supports this option.

	-c, -color
		Use color output.

	-t, -timestamp
		Display current time when using monitor option.

	-ts, -tshort
		Like -timestamp, but use shorter format.

	-rc, -rcvbuf<SIZE>
		Set the netlink socket receive buffer size, defaults to 1MB.

	-iec   Print human readable rates in IEC units (e.g. 1Ki = 1024).

COMMAND
	Specifies the action to perform on the object.  The set of possible
	actions depends on the object type.  As a rule, it is possible to add,
	delete and show (or list ) objects, but some objects do not allow all
	of these operations or have some additional commands. The help command
	is available for all objects.

	Without a command, All objects print the respective usage summary.

SEE ALSO
	ip man OBJECT || ip OBJECT -man
	ip OBJECT man COMMAND || ip OBJECT COMMAND -man
`
