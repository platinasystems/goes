// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"encoding/pem"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/log/style"
)

type Client struct {
	sync.RWMutex
	Confirmation Confirmation
	ID           uint32
	Exchange     *box.Cipher
	remote       map[uint32]*box.Cipher
}

func NewClient() (cl *Client, err error) {
	cl = &Client{
		remote: make(map[uint32]*box.Cipher),
	}
	cl.Exchange, err = box.NewCipher()
	return
}

func (cl *Client) Peer(pemdata []byte) error {
	pb, _ := pem.Decode(pemdata)
	if pb == nil {
		return egress.Marked(ErrInvalid)
	}
	if err := cl.Exchange.ECDH(pb); err != nil {
		return egress.Marked(err)
	}
	if s, ok := cl.Exchange.PublicKey.Remote.Headers["confirmation"]; !ok {
		return egress.Marked(ErrUnconfirmed)
	} else if _, err := fmt.Sscan(s, &cl.Confirmation); err != nil {
		return egress.Marked(err)
	}
	if s, ok := cl.Exchange.PublicKey.Remote.Headers["id"]; !ok {
		return egress.Marked(ErrUnidentified)
	} else if _, err := fmt.Sscan(s, &cl.ID); err != nil {
		return egress.Marked(err)
	}
	return nil
}

func (cl *Client) Remote(id uint32) (remote *box.Cipher, ok bool) {
	cl.RLock()
	defer cl.RUnlock()
	remote, ok = cl.remote[id]
	return
}

func (cl *Client) AddRemote(id uint32, pub *pem.Block) (
	remote *box.Cipher, err error,
) {
	cl.Lock()
	defer cl.Unlock()
	var ok bool
	remote, ok = cl.remote[id]
	if !ok {
		remote, err = cl.Exchange.Clone(pub)
		if err == nil {
			cl.remote[id] = remote
		}
	}
	return
}

// Until cancelled, send periodic keep alives (empty boxes) to all members.
func (cl *Client) KeepAlive(
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.Conn,
	period time.Duration,
) {
	defer wg.Done()
	defer style.Recovery(context.Canceled)

	bx := box.New().Empty().ToAll().From(cl.ID).
		Close(cl.Exchange).Seal(cl.Exchange)
	defer bx.Recycle()

	t := time.NewTicker(period)
	defer t.Stop()

	for {
		if _, err := conn.Write(bx); err != nil {
			panic(err)
		}
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		case <-t.C:
		}
	}
}
