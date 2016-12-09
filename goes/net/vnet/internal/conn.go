// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package internal

import (
	"net"
	"sync"

	"github.com/platinasystems/go/goes/sockfile"
)

var Conn conn

type conn struct {
	net.Conn
	sync.Mutex
}

func (c *conn) Connect() (err error) {
	c.Lock()
	defer c.Unlock()
	if c.Conn == nil {
		c.Conn, err = sockfile.Dial("vnet")
	}
	return
}

func (c *conn) Close() (err error) {
	c.Lock()
	defer c.Unlock()
	if c.Conn != nil {
		err = c.Close()
		c.Conn = nil
	}
	return
}
