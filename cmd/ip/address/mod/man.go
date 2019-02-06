package mod

const Man = `
DESCRIPTION
	The address is a protocol (IPv4 or IPv6) address attached to a network
	device. Each device must have at least one address to use the corre‐
	sponding protocol. It is possible to have several different addresses
	attached to one device. These addresses are not discriminated, so that
	the term alias is not quite appropriate for them and we do not use it
	in this document.

	The ip address command displays addresses and their properties, adds
	new addresses and deletes old ones.


   ip address add - add new protocol address.
	dev IFNAME
	      the name of the device to add the address to.


	local ADDRESS (default)
	      the address of the interface. The format of the address depends
	      on the protocol. It is a dotted quad for IP and a sequence of
	      hexadecimal halfwords separated by colons for IPv6. The ADDRESS
	      may be followed by a slash and a decimal number which encodes
	      the network prefix length.


	peer ADDRESS
	      the address of the remote endpoint for pointopoint interfaces.
	      Again, the ADDRESS may be followed by a slash and a decimal num‐
	      ber, encoding the network prefix length. If a peer address is
	      specified, the local address cannot have a prefix length. The
	      network prefix is associated with the peer rather than with the
	      local address.


	broadcast ADDRESS
	      the broadcast address on the interface.

	      It is possible to use the special symbols '+' and '-' instead of
	      the broadcast address. In this case, the broadcast address is
	      derived by setting/resetting the host bits of the interface pre‐
	      fix.


	label LABEL
	      Each address may be tagged with a label string.  In order to
	      preserve compatibility with Linux-2.0 net aliases, this string
	      must coincide with the name of the device or must be prefixed
	      with the device name followed by colon.


	scope SCOPE_VALUE
	      the scope of the area where this address is valid.  The avail‐
	      able scopes are listed in file /etc/iproute2/rt_scopes.  Prede‐
	      fined scope values are:

	              global - the address is globally valid.

	              site - (IPv6 only, deprecated) the address is site
	              local, i.e. it is valid inside this site.

	              link - the address is link local, i.e. it is valid only
	              on this device.

	              host - the address is valid only inside this host.


	valid_lft LFT
	      the valid lifetime of this address; see section 5.5.4 of RFC
	      4862. When it expires, the address is removed by the kernel.
	      Defaults to forever.


	preferred_lft LFT
	      the preferred lifetime of this address; see section 5.5.4 of RFC
	      4862. When it expires, the address is no longer used for new
	      outgoing connections. Defaults to forever.


	home   (IPv6 only) designates this address the "home address" as
	      defined in RFC 6275.


	mngtmpaddr
	      (IPv6 only) make the kernel manage temporary addresses created
	      from this one as template on behalf of Privacy Extensions
	      (RFC3041). For this to become active, the use_tempaddr sysctl
	      setting has to be set to a value greater than zero.  The given
	      address needs to have a prefix length of 64. This flag allows to
	      use privacy extensions in a manually configured network, just
	      like if stateless auto-configuration was active.


	nodad  (IPv6 only) do not perform Duplicate Address Detection (RFC
	      4862) when adding this address.


	noprefixroute
	      Do not automatically create a route for the network prefix of
	      the added address, and don't search for one to delete when
	      removing the address. Changing an address to add this flag will
	      remove the automatically added prefix route, changing it to
	      remove this flag will create the prefix route automatically.


	autojoin
	      Joining multicast groups on Ethernet level via ip maddr command
	      does not work if connected to an Ethernet switch that does IGMP
	      snooping since the switch would not replicate multicast packets
	      on ports that did not have IGMP reports for the multicast
	      addresses.

	      Linux VXLAN interfaces created via ip link add vxlan have the
	      group option that enables them to do the required join.

	      Using the autojoin flag when adding a multicast address enables
	      similar functionality for Openvswitch VXLAN interfaces as well
	      as other tunneling mechanisms that need to receive multicast
	      traffic.

   ip address delete - delete protocol address
	Arguments: coincide with the arguments of ip addr add.  The device name
	is a required argument. The rest are optional.  If no arguments are
	given, the first address is deleted.

   ip address flush - flush protocol addresses
	This command flushes the protocol addresses selected by some criteria.


	This command has the same arguments as show except that type and master
	selectors are not supported.  Another difference is that it does not
	run when no arguments are given.


	Warning: This command and other flush commands are unforgiving. They
	will cruelly purge all the addresses.


	With the -statistics option, the command becomes verbose. It prints out
	the number of deleted addresses and the number of rounds made to flush
	the address list.  If this option is given twice, ip address flush also
	dumps all the deleted addresses in the format described in the previous
	subsection.

EXAMPLES
	ip address add 2001:0db8:85a3::0370:7334/64 dev eth1
	   Adds an IPv6 address to network interface eth1.

	ip address delete 2001:0db8:85a3::0370:7334/64 dev eth1
	   Delete the IPv6 address added above.

	ip address flush dev eth4 scope global
	   Removes all global IPv4 and IPv6 addresses from device eth4. With‐
	   out 'scope global' it would remove all addresses including IPv6
	   link-local ones.

SEE ALSO
	ip man address || ip address -man
	man ip || ip -man

AUTHOR
	Original Manpage by Michail Litvak <mci@owl.openwall.com>
`
