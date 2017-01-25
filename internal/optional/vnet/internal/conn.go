// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/platinasystems/go/internal/sockfile"
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
		err = c.Conn.Close()
		c.Conn = nil
	}
	return
}

// Exec runs a vnet cli command and copies output to given io.Writer.
func (c *conn) Exec(w io.Writer, args ...interface{}) (err error) {
	// Send cli command to vnet.
	fmt.Fprintln(c, args...)
	for {
		// First read 32 bit network byte order length.
		var tmp [4]byte
		if _, err = c.Read(tmp[:]); err != nil {
			return
		}
		if l := int64(binary.BigEndian.Uint32(tmp[:])); l == 0 {
			// Zero length means end of vnet command output.
			break
		} else {
			// Otherwise copy input to output.
			if _, err = io.CopyN(w, c, l); err != nil {
				return
			}
		}
	}
	return
}
