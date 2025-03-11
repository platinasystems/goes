/*
Package “vpn” provides these [goes] [Features] to implement and manage a
secure, Virtual Private Network.

# Prerequisites

  - [Certificate] creates a PEM encoded x509 certificate file.

    <goes> new vpn certificate [flags]

  - [Ed25519] creates a PEM encoded ed25519 signature key file.

    <goes> new vpn signature [flags]

# Daemons

  - [Registry] is a web server providing a REST interface to its
    ephemeral tables; and persistent, read-only [PathConfigDir] and
    read-write [PathStateDir] subscriber certificates.

  - [Exchange] is a UDP server that forwards ciphered packets between guest's.

  - [Guest] is a UDP server that forwards ciphered packets between an exchange
    and a network tunnel interface.

    <goes> start vpn <daemon> [flags]

# Registry administration

  - [Approve] VPN subscription.

  - [Deny] VPN subscription.

  - [Ping] registry.

  - [Revoke] VPN subscription.

    <goes> vpn <admin> [flags]

# Registry information

  - [Admins] prints the names of authorized VPN administrators.

  - [Pending] prints requesting subscriber certificates.

  - [Subscribers] prints the names of current subscribers.

    <goes> show vpn <object> [flags]

# Client administration

  - [Certify] adds registry to subscriptions.

  - [Subscribe] requests VPN subscription.

  - [Unsubscribe] from VPN.

    <goes> vpn <admin> [flags]

# Client information

  - [Certificaté] prints parsed certificate.

  - [Subscriptions] prints parsed registry subscriptions.

  - [Signature] prints algorithm.

    <goes> show vpn <object> [flags]

# References

[goes]: https://github.com/platinasystems/goes/v2/pkg/goes

[goes-util]: https://github.com/platinasystems/goes/v2/pkg/goes-util
*/
package vpn
