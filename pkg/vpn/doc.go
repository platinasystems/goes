/*
Package “vpn” provides these [goes] [Features] to implement and manage a
secure, Virtual Private Network.

# Prerequisites

  - [Certificate] creates a PEM encoded x509 certificate file.

    <goes> new vpn certificate [flags]

  - [Ed25519] creates a PEM encoded ed25519 signature key file.

    <goes> new vpn signature [flags]

  - [ValueOfStringFlag]([NameConfigFlag]), an admin supplied file
    defining the VPN's prefix and persistent address assignments.

# Config Example

	vpn:
	  prefix: fc00:1234::/64
	  address:
	    guesta: fc00:1234::1
	    guestb: fc00:1234::2
	    guestc: fc00:1234::3
	    exchange0: fc00:1234::10
	    exchange1: fc00:1234::11
	    exchange2: fc00:1234::12

# Daemons

  - [Registry] is a web server providing a [REST] interface to adminster
    and introduce subscribers.

  - [Exchange] is a VPN subscriber providing secure packet forwarding.

  - [Guest] is a VPN subscriber that encode/decodes secure packets
    between an exchange and a network tunnel interface.

    <goes> start vpn <daemon> [flags]

# Secure Packet Exchange

This VPN uses [crypto/cipher.NewGCMWithRandomNonce]
to secure [UDP] tunneled [IP] packets with
Authenticated Encrypted and Associated Data [AEAD]

The prefaced random [NONCE] is the encrypted
with the the cipher block wrapped by the [GCM].

The attached authenticated data has the FROM and TO subscriber
identifiers uniquely assigned by the registry.

[Unicast] packets are secured with guest/guest keys derived from
registry introduction and are simply forwarded by the exchange.

[Multicast] packets are secured with guest/exchange keys and are decoded
by the exchange before replicating, encoding and forwarding to each of the
other active guests.

# Registry administration

  - [RestAdmin] “approve” <candidate>

  - [RestAdmin] “deny”  <candidate>

  - [RestAdmin] “unsubscribe” <subscriber>

  - [RestPing]

    <goes> vpn <admin> [flags]

# Registry information

  - [RestShow] “active”

    Print the names of checked-in subscribers.

  - [RestShow] “address” [subscriber]

    Print the assigned address of the named or all subscribers.

  - [RestShow] “hosts” [address]

    Print the addressed or all subscriber assignments.

  - [RestShow] “pending”

    Print the requesting subscriber certificates.

  - [RestShow] “subscriber”

    Prints the namesd or all current subscribers.

  - [RestShow] “tenant”

    Print subscriber at given address.

    <goes> show vpn <object> [flags]

# Client administration

  - [RestCertify] <url>

    Adds registry certificate to root.

  - [RestSubscribe]

    Requests VPN subscription.

  - [Unsubscribe]

    Remove self from VPN.

    <goes> vpn <admin> [flags]

# Client information

  - [ShowCertificate]

    Prints formatted certificate.

  - [ShowSignature]

    Prints algorithm.

    <goes> show vpn <object> [flags]

# References

[AEAD]: https://en.wikipedia.org/wiki/Authenticated_encryption
[GCM]: https://en.wikipedia.org/wiki/Galois/Counter_Mode
[goes]: https://github.com/platinasystems/goes/v2/pkg/goes
[IP]: https://en.wikipedia.org/wiki/Internet_Protocol
[Multicast]: https://en.wikipedia.org/wiki/Multicast
[NONCE]: https://en.wikipedia.org/wiki/Nonce
[REST]: https://en.wikipedia.org/wiki/REST
[Unicast]: https://en.wikipedia.org/wiki/Unicast
*/
package vpn
