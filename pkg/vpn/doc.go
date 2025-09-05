/*
Package vpn provides these [goes] [Features] to implement and manage a
secure, Virtual Private Network.
This package is imported by the full featured [goes command] and the standalone
[goes-vpn] command.

# Daemons

This package includes three service daemons:
[Registry], alternate [Exchange], and [Guest].

# Prerequisites

Each service daemon requires a [VpnSigFile] and [VpnCertFile]
w/in [VpnConfigDir] or specified path name.

Use [NewEd25519] to generate a [PEM] encoded [Ed25519] signature key file.

	goes new vpn signature [flags]		# or,
	goes-vpn new signature [flags]

To validate and print the algorithm of the existing signature,

	goes show vpn signature [flags]		# or,
	goes-vpn show signature [flags]

Use [CreateCertificate] make a PEM encoded, x509 certificate file,

	goes new vpn certificate [flags]	# or,
	goes-vpn new certificate [flags]

To validate and format an existing certificate,

	goes show vpn certificate [flags]	# or,
	goes-vpn show certificate [flags]

In addition, the registry may be configured with these optional,
VpnConfigDir or specified path files:
[VpnAdminsFile], [VpnExchangesFile], and [VpnHostsFile].

The [VpnAdminsFile] is line separated list of certificate common names
permitted to administrate VPN subscriptions.

The [VpnExchangesFile] is a line separated list of key/value(s) assigning guest
exchange preference in left (highest) to right (lowest) order.
The registry is the implied lowest precedent exchange.

	guesta	exchange0 exchange1
	guestb	exchange0 exchange1
	guestc	exchange0 exchange1

The [VpnHostsFile] is a line separated list of key/value(s) with [/etc/hosts]
format that assigns static VPN addresses.

	guesta: fc00:1234::1
	guestb: fc00:1234::2
	guestc: fc00:1234::3

# Registry

The registry is web server providing static assets and a [REST] interface to
adminster and introduce VPN subscribers.
It's also the default, secure packet exchange.

After making the prerequisite and optional files, start the registry with,

	goes start vpn registry [flags]		# or,
	goes-vpn start registry [flags]

See,

	goes help start vpn <daemon>		# or,
	goes-vpn help start <daemon>

This opens a [TLS] listener with the referenced certificate and signature.
It also opens [UDP] listener for secure packet forwarding.

A complete list of assets and [REST] functions may be retrieved with,

	curl --tlsv1.3 [-k] https://<host>[:port]

Once started, you may use these commands on the registry itself, or any
permitted administrator, to issue and print the results of respective REST
command.

	goes show vpn <object>			# or,
	goes-vpn show <object>

	object: admins, exchanges, guests, hosts, pending, start, status,
		subscriber

# Subscription

If the registry certificate hasn't yet been uploaded to an authority,
the subscriber (an alt-exchange or guest) must [RestCertify] the registry with,

	goes[-]vpn certify [flags] https://<host>[:port]

This will pull the certificate from the [TLS] [REST] response.

It must then request a VPN subscription with,

	goes[-]vpn subscribe [flags]

An admin must then “approve” or “deny” the subscription request with,

	goes[-]vpn <approve|deny> [flags] <subscriber>

The admin may also “unsubscribe” an existing subscriber with,

	goes[-]vpn <unsubscribe> [flags] <subscriber>

In addition, any subscriber may unsubscribe themself with,

	goes[-]vpn <unsubscribe> [flags]

# Alt-[Exchange]

An admin may then start alternate packet exchanges to scale and localize
throughput.

	goes start vpn exchange [flags]		# or,
	goes-vpn start exchange [flags]

This will use [RestExchangeCheckin] to register it's own certificate with
contained public signature.
It will then [RestValidateCheckinResponse] the registry response before parsing
the assigned VPN [Id] and service port number from the response body.

It then forwards the secured packets received on the service port to the
authenticated destination.

## [Guest]

A new guest may retrieve its program from the registry with,

	curl --tlsv1.3 --insecure -o PATH/goes[-vpn] \
		https://<host>[:port]/goes[-vpn]-<goos>-<goarch>

To start a guest,

	goes start vpn guest [flags]		# or,
	goes-vpn start guest [flags]

This generates a new [mlkem.DecapsulationKey768] and then extracts and forwards
its public [mlkem.EncapsulationKey768] as raw bytes w/in the body of a
[RestGuestCheckin].
After [RestValidateCheckinResponse] this [json.Unmarshal] the [GuestReceipt]
containing the assigned [Id], VPN prefix, and exchane precedence.
The daemon then creates a tunnel interface with the assigned prefix before
starting a main loop to lock-free forward between the UDP listening socket and
the tunnel interface.

# Auto-upgrade

Before respective checkin, each guest and alt-exchange uses
[RestAssertVcsMatch] to validate its [debug.BuildInfo]["vcs.revision"].
If the registry returns [http.StatusUpgradeRequired], the respective daemon
will try to download and overwrite itself and exit with [xos.EX_TEMPFAIL].
The service wrapper (e.g. systemd) should then restart the respective daemon.

# Authenticated Forwarding

Guests periodically sends an unencrypted [NewGreeting] to each of its assigned
exchanges service ports.
This unencrypted greeting has signed current and daemon start times
appended by the guest's assigned [Id] as both FROM and TO.

The exchange recognizes the greeting from the matching identifiers.
It then verify's the signed time stamps before updating it's service table for
the referenced guest and sending an equivalent reply.

Similarly, the guest recognizes and validates the exchange's acknowledgement
before updating it's forwarding table.

# Invitation

The guests establish shared encapsulation through the registry.
When a guest recieve's a tunneled packet to an unknown VPN address,
or a secured UDP packet from an uknown Id, the guest first asks
[RestWhois] the peer's assigned Id and address along with it's
registered [mlkem.EncapsulationKey768].
The guest then uses [mlkem.EncapsulationKey768.Encapsulate] with the
encap key to generate a new shared key and companion cipher text.
It then registers the cipher text with [RestInvite].
The first invite wins meaning that if the peer had first issued an invite, that
cipher text would be used as the shared key instead of the one generated by the
initiating guest.

# Secure Packet Exchange

Guests use [crypto/cipher.NewGCMWithRandomNonce]
to secure [Unicast], [UDP] tunneled [IP] packets with
Authenticated Encrypted and Associated Data [AEAD]

[Multicast] and [Broadcast] packets are ignored.

The prefaced random [NONCE] is block ciphered after it's used to
[GCM] cipher the attached packet.

The attached authenticated data has 4-byte, big-endian FROM and TO subscriber
identifiers uniquely assigned by the registry.

# Version

The most significant 4 bits of the subscriber's assigned [Id] is its version
that the registry increments on each checkin.  If peers detect change, they
repeat the shared key handshake.

# References

[AEAD]: https://en.wikipedia.org/wiki/Authenticated_encryption
[Broadcast]: https://en.wikipedia.org/wiki/Broadcasting_(networking)
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
