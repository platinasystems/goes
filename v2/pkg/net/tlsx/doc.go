// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

/*
This package provides [TLS] network in which hosts respond to guest requests
through an exchange.

The rendezvous protocol begins with both host and guest clients connecting
to then handshaking with an exchange registration service that responds with
the 2-byte, big-endian rendezvous service port.

	client   exchange
	│←— HANDSHAKE —→│
	│←———————— PORT │

The TLS handshake delivers the client's (host or guest) certifcate to the
exchange that places it in a pending queue.  The handshake also delivers the
exchange certificate to the clients that add these to their list of root
certificate authorities.  The exchange administrator approves or denys the
pending certificates while the client adminstrator(s) confirm each exchange.

After an adminstrative approval, the host and guest make authenticated TLS
connections with the exchange service port to rendezvous with through this
protocol.

	host        exchange        guest
	│←— HANDSHAKE —→║←— HANDSHAKE —→│
	│ ACCEPT ——————→║               │
	│←————————— ACK ║               │
	│               …               │
	│               ║←————— CONNECT │
	│←———————— RING ║               │
	│ ACK ───—─────→║               │
	│               ║ ACK —————————→│
	│               …               │
	│←─── DATA ─—──→║←───— DATA ───→│
	│               …               │

Each of these signals are encoded with one or more type, length, value
(TLV) format messages.

After handshake, the host notifies the exchange that it's prepared to service
guest requests with the ACCEPT signal.  It may, and should open more than one
TLS connection on redundant exchanges with scale such that it always has at
least one connection pending service.  After the exchange's positive
acknowledgment, the host switches to its service mode and waiting for a RING
with caller ID.

Likewise, after handshake the guest makes a rendezvous request with a CONNECT
signal that the exchange matches to a host connection.  Upon match, the
exchange issues a RING signal to the host and relays guest acknowledgment.
Thereafter, the exchange continues to forward uninterpreted data streams
between guest and host connections.

The service streams usually begin with TLV encoded guest requests and host
acknowledgment that persist until the guest closes its connection.  However,
some requests may change the service mode to another encoding for an
interactive TTY session or end-to-end encryption.

All exchange client's may both consume and provide service over
simultaneous connections; so, each service need only be a simplex
request/acknowledge protocol.

# ACCEPT

The host requests service registry with this TLV signal.

	accept

# CONNECT

The guest requests rendevous with a host connection with these TLV
arguments:

	connect SUBJECT_KEY_ID

Where SUBJECT_KEY_ID is the hexadecimal string representing the
respective field within the host [x509] certificate.

# RING

The exchange relays a guest call request with these TLV arguments:

	ring SUBJECT_KEY_ID

Where again, SUBJECT_KEY_ID is the hexadecimal string representing the
respective field within the guest [x509] certificate.  The host positive
acknowledgment includes its main module reference, e.g.

	PATH@VERSION

# References

[PEM]: https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail
[TLS]: https://en.wikipedia.org/wiki/Transport_Layer_Security
[x509]: https://en.wikipedia.org/wiki/X.509
*/
package tlsx
