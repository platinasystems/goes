// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type Client struct {
	http.Client
	dialer  *net.Dialer
	network string
	address string
	cfg     *tls.Config
}

// A nil dialer is replaced by new(net.Dialer).
//
// A non-empty {network, address} overrides {"tcp", HOST} of
// GET and PUT URLs <https://[USER"@"]HOST[":"PORT]/...>
// So, for unix socket clients, use {"unix", PATH_OR_ABSTRACT_NAME}
// and <https://unix/...> URLs.
func NewClient(
	cfg *tls.Config,
	dialer *net.Dialer,
	network, address string,
) *Client {
	if dialer == nil {
		dialer = new(net.Dialer)
	}
	xport := &http.Transport{
		TLSClientConfig:    cfg,
		DisableCompression: true,
		ForceAttemptHTTP2:  true,
	}
	cl := &Client{
		Client: http.Client{
			Transport: xport,
			Timeout:   3 * time.Second,
		},
		dialer:  dialer,
		network: network,
		address: address,
		cfg:     cfg,
	}
	xport.DialTLSContext = cl.dial
	return cl
}

func (cl *Client) dial(
	ctx context.Context,
	network, address string,
) (net.Conn, error) {
	if len(cl.network) > 0 {
		network = cl.network
	}
	if len(cl.address) > 0 {
		address = cl.address
	}
	if cl.dialer.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cl.dialer.Timeout)
		defer cancel()
	}
	if !cl.dialer.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, cl.dialer.Deadline)
		defer cancel()
	}
	conn, err := cl.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	tlsconn := tls.Client(conn, cl.cfg)
	if err := tlsconn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return tlsconn, nil
}
