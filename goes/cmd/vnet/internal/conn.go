// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/platinasystems/atsock"
	"strings"
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
		c.Conn, err = atsock.Dial("vnet")
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
func (c *conn) Exec(w io.Writer, args ...string) (err error) {
	var werr error

	// Send cli command to vnet.
	fmt.Fprintf(c, "%s\n", strings.Join(args, " "))

	// Ignore pipe error e.g. vnet command | head
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)

	for {
		// First read 32 bit network byte order length.
		var tmp [4]byte
		if _, err = c.Read(tmp[:]); err != nil {
			return
		}
		if l := int64(binary.BigEndian.Uint32(tmp[:])); l == 0 {
			// Zero length means end of vnet command output.
			break
		} else if werr == nil {
			// Otherwise copy input to output up to first error
			_, werr = io.CopyN(w, c, l)
		}
	}
	return
}
