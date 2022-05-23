// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/ipc"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var Reg = struct {
	IPC   ipc.Ipc
	mutex sync.Mutex
	l     []*x509.Certificate
}{
	IPC: ipc.Preface(fmt.Sprint(program.Base.String(), ".registry")),
}

func Registry(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
	xp uint,
) {
	defer wg.Done()

	tlsc, err := cert.Value()
	if err != nil {
		panic(err)
	}
	tlscs := []tls.Certificate{tlsc}
	if tlsc.Leaf == nil {
		panic("nil leaf")
	}
	if len(tlsc.Leaf.DNSNames) == 0 {
		panic("no DNS names")
	}
	sn := tlsc.Leaf.DNSNames[0]
	for c := range accept.With(ctx, ln, make(chan net.Conn, 4)) {
		wg.Add(1)
		sv := tls.Server(c, &tls.Config{
			Certificates: tlscs,
			ServerName:   sn,
			ClientAuth:   tls.RequireAnyClientCert,
		})
		go func(sv *tls.Conn) {
			wg.Done()
			defer sv.Close()

			err := sv.HandshakeContext(ctx)
			if err != nil {
				if err = suppress.Errors(err,
					context.Canceled,
					net.ErrClosed,
				); err != nil {
					Elog(err)
				}
				return
			}
			cs := sv.ConnectionState()
			if len(cs.PeerCertificates) == 0 {
				Elog(ErrNoPeer)
				return
			}
			binary.Write(sv, binary.BigEndian, uint16(xp))
			Reg.mutex.Lock()
			Reg.l = append(Reg.l, cs.PeerCertificates[0])
			Reg.mutex.Unlock()
		}(sv)
	}
}

func Subscribe(ctx context.Context, addr string) error {
	var name string
	tlsc, err := cert.Value()
	if err != nil {
		return err
	}
	var dl net.Dialer
	nw := "tcp"
	if len(addr) == 0 {
		nw = ipc.Network
		addr, err = Reg.IPC.Address()
		if err == nil {
			name, err = host.Name.Value()
		}
	} else {
		if i := strings.LastIndex(addr, ":"); i < 0 {
			return fmt.Errorf("%q: expect [<dns>]:<port>", addr)
		} else if i == 0 {
			name, err = host.Name.Value()
		} else {
			name = addr[:i]
		}
	}
	if err != nil {
		return err
	}
	c, err := dl.DialContext(ctx, nw, addr)
	if err != nil {
		return err
	}
	defer c.Close()
	cl := tls.Client(c, &tls.Config{
		Certificates:       []tls.Certificate{tlsc},
		ServerName:         name,
		InsecureSkipVerify: true,
	})
	if err = cl.HandshakeContext(ctx); err != nil {
		return err
	}
	cs := cl.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return ErrNoPeer
	}
	var port uint16
	binary.Read(cl, binary.BigEndian, &port)
	headers := map[string]string{
		"port": fmt.Sprint(port),
	}
	return certs.Exchanges.Add(headers, cs.PeerCertificates[0])
}
