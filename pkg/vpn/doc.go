/*
Package vpn provides these [goes] [Features] to implement and manage a
secure, Virtual Private Network.
This package is imported by the full featured [goes command] and the
standalone [goes-vpn] command.

# DAEMONS

This package includes three service daemons:
[Registry], alternate [Exchange], and [Guest].

# PREREQUISITES

Each service daemon requires a [sig.File] and [cert.Client] common name
of a certificate within the [PEM] encoded files of [xmain.ConfigDir].

Use [sig.New] to generate a [PEM] encoded [Ed25519] signature key file.

	goes new vpn signature [flags]
	goes-vpn new signature [flags]

To validate and print the algorithm of the existing signature,

	goes show signature [flags]
	goes-vpn show signature [flags]

Use [cert.New] to create a PEM encoded, x509 certificate file,

	goes new certificate [flags]
	goes-vpn new certificate [flags]

The default output file is “cert.pem” within [xmain.ConfigDir].

To validate and format an existing certificate,

	goes show certificate [flags]
	goes-vpn show certificate [flags]

In addition, the registry may be configured with these [xmain.ConfigDir]
files: [AdminsFile], [ExchangesFile], [HostsFile], and [ZonesFile].

The [AdminsFile] is a line separated list of certificate common names
permitted to administrate VPN subscriptions.

	guesta
	guestb

The [ExchangesFile] is a line separated list of key/value(s) assigning
guest exchange preference in left (highest) to right (lowest) order.
The registry is implied as the lowest precedent exchange.

	guesta	exchange0 exchange1
	guestb	exchange0 exchange1
	guestc	exchange0 exchange1

The [HostsFile] is a line separated list of key/value(s) with
[/etc/hosts] format that assigns static VPN addresses.

	guesta: fc00:1234::1
	guestb: fc00:1234::2
	guestc: fc00:1234::3

The [ZonesFile] is a line separated list of key/value(s) that
assign guests to specific or all (*) VPN zones.
It may contain up to eight unique zone names.

	guesta *
	guestb backend

Unspecified guests are assigned to the default, unnamed zone.

# DOH

Guest pkg/net-tools like ping and netcat may be configured to [DNS]
lookup VPN peers through the registry's [DOH] service with a
[xdnsdoh.ConfigFile] containing,

	url	https://registry.vpn.default.svc.cluster.local:8003/dns-query
	search	example.platina.io

# REGISTRY

The registry is a web server providing static assets and a [REST]
interface to administer and introduce VPN subscribers.
It's also the default, secure packet exchange.

After making the prerequisite and optional files, start the registry with,

	goes start vpn registry [flags]
	goes-vpn start registry [flags]

See,

	goes help start vpn registry
	goes-vpn help start registry

This opens a [TLS] listener with the indicated certificate and signature.
It also opens [UDP] listener for secure packet exchange.

A complete list of assets and [REST] functions may be retrieved with,

	curl --tlsv1.3 [-k] https://<host>[:port]

Once started, you may use these commands on the registry itself, or any
permitted administrator, to issue and print the results of respective REST
command.

	goes show vpn <object>
	goes-vpn show <object>

	object: admins, exchanges, pending, start, status, subscriber

# SUBSCRIPTION

If the registry certificate hasn't yet been uploaded to an authority,
the subscriber (an alt-exchange or guest) must [RestCertify] the registry with,

	goes[-]vpn certify [flags] https://<host>[:port]

This will pull the registry certificate from the [TLS] [REST] response.

The subscriber then requests a VPN subscription with,

	goes[-]vpn subscribe [flags]

The registry must be specified by flag or environment variable. See,

	goes[-]vpn help subscribe

An admin must then “approve” or “deny” the subscription request with,

	goes[-]vpn approve [flags] <subscriber> [zone]...
	goes[-]vpn deny [flags] <subscriber>

The admin may also “unsubscribe” an existing subscriber with,

	goes[-]vpn <unsubscribe> [flags] <subscriber>

In addition, any subscriber may unsubscribe themself with,

	goes[-]vpn <unsubscribe> [flags]

# ALT-EXCHANGE

An admin may then start alternate packet exchanges to scale and localize
throughput.

	goes start vpn exchange [flags]
	goes-vpn start exchange [flags]

The exchange makes a [RestCheckinExchangeReq] to register itself.
If successful, the registry responds with an assigned VPN [Id] and
service port number.

The exchange then forwards the secured packets received on the service
port to the authenticated destination.

# GUEST

A new guest may retrieve its program from the registry with,

	curl --tlsv1.3 --insecure -o PATH/goes[-vpn] \
		https://<host>[:port]/goes[-vpn]-<GOOS>-<GOARCH>

To start a guest,

	goes start vpn guest [flags]		# or,
	goes-vpn start guest [flags]

This generates a new [mlkem.DecapsulationKey768] and then extracts and forwards
its public [mlkem.EncapsulationKey768] as raw bytes w/in the body of a
[RestCheckinGuestReq].
If successful, the registry responds with a [json] encoded [GuestReceipt]
containing the assigned [Id], prefix and exchange precedence.

The daemon then creates a tunnel interface with the assigned prefix before
starting a main loop to lock-free forward between the UDP listening socket and
the tunnel interface.

# AUTO-UPGRADE

Each [REST] request and response includes a [RestVcsRevision] header containing
the sender's [xprogram.VcsRevision].
The registry rejects checkin requests with outdated version.
Likewise, the exchange and guest daemons will assert responding version match.
If mismatched, the respective daemon will try to download and overwrite
itself and exit with [xos.EX_TEMPFAIL].
The service wrapper (e.g. systemd) should then restart the respective daemon.

# AUTHENTICATED FORWARDING

Guests periodically send an unencrypted [NewGreeting] to each of its assigned
exchanges.
This unencrypted greeting has signed current and daemon start times
appended by the guest's assigned [Id] as both FROM and TO.

The exchange recognizes the greeting from the matching identifiers.
It then verify's the signed time stamps before updating it's service table for
the referenced guest and sending an equivalent reply.

Similarly, the guest recognizes and validates the exchange's acknowledgement
before updating it's forwarding table.

# INVITATION

The guests establish shared encapsulation through the registry.
When a guest recieve's a tunneled packet to an unknown VPN address,
or a secured UDP packet from an uknown Id, the guest first makes a
[RestWhoisReq] with the unknown address or Id.
The registry responds with the peer's assigned Id and address along with
it's registered [mlkem.EncapsulationKey768].
The guest then uses [mlkem.EncapsulationKey768.Encapsulate] with the
encap key to generate a new shared key and companion cipher text.
It then makes a [RestInviteReq] to register the cipher text.
The first invite wins meaning that if the peer had first issued an invite, that
cipher text would be used as the shared key instead of the one generated by the
later guest.

# SECURE PACKET EXCHANGE

Guests use [crypto/cipher.NewGCMWithRandomNonce]
to secure [Unicast], [UDP] tunneled [IP] packets with
Authenticated Encrypted and Associated Data [AEAD]

[Multicast] and [Broadcast] packets are ignored.

The prefaced random [NONCE] is block ciphered after it's used to
[GCM] cipher the attached packet.

The attached authenticated data has 4-byte, big-endian FROM and TO subscriber
identifiers uniquely assigned by the registry.

# VERSION

The most significant 4 bits of the subscriber's assigned [Id] is its version
that the registry increments on each checkin.
Guests periodically request [RestReviseIds] and then repeat the shared
key handshake for every changed Id.

# REFERENCES

[AEAD]: https://en.wikipedia.org/wiki/Authenticated_encryption
[Broadcast]: https://en.wikipedia.org/wiki/Broadcasting_(networking)
[DNS]: https://en.wikipedia.org/wiki/Domain_Name_System
[DOH]: https://en.wikipedia.org/wiki/DNS_over_HTTPS
[Ed25519]: https://en.wikipedia.org/?title=Ed25519&redirect=no
[/etc/hosts]: https://en.wikipedia.org/wiki/Hosts_(file)
[GCM]: https://en.wikipedia.org/wiki/Galois/Counter_Mode
[goes]: https://github.com/platinasystems/goes/v2/pkg/goes
[goes command]: https://github.com/platinasystems/goes/v2
[goes-vpn]: https://github.com/platinasystems/goes/v2/cmd/goes-vpn
[IP]: https://en.wikipedia.org/wiki/Internet_Protocol
[Multicast]: https://en.wikipedia.org/wiki/Multicast
[NONCE]: https://en.wikipedia.org/wiki/Nonce
[PEM]: https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail
[REST]: https://en.wikipedia.org/wiki/REST
[TLS]: https://en.wikipedia.org/wiki/Transport_Layer_Security
[UDP]: https://en.wikipedia.org/wiki/User_Datagram_Protocol
[Unicast]: https://en.wikipedia.org/wiki/Unicast
*/
package vpn
