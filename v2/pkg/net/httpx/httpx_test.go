// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package httpx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/net/ipc"
)

func Test(t *testing.T) { test(t, 1) }

func Benchmark1(b *testing.B)      { bm(b, 1) }
func Benchmark10(b *testing.B)     { bm(b, 10) }
func Benchmark100(b *testing.B)    { bm(b, 100) }
func Benchmark1000(b *testing.B)   { bm(b, 1000) }
func Benchmark10000(b *testing.B)  { bm(b, 10000) }
func Benchmark100000(b *testing.B) { bm(b, 100000) }

func bm(b *testing.B, n uint) {
	for i := 0; i < b.N && !b.Failed(); i++ {
		test(b, n)
	}
}

func test(tb testing.TB, n uint) {
	var wg sync.WaitGroup
	defer wg.Wait()
	var err error
	const (
		sv = iota
		cl
		roles
	)
	var gen = [roles]struct {
		name string
		pk   keycert.PrivateKey
		cert struct {
			x509 *x509.Certificate
			tls  [1]tls.Certificate
		}
		block struct {
			pk   *pem.Block
			x509 *pem.Block
		}
	}{
		sv: {name: "server.https.goes"},
		cl: {name: "client.https.goes"},
	}

	now := time.Now()
	expire := now.Add(365 * 24 * time.Hour)
	for i := 0; i < roles; i++ {
		gen[i].pk, gen[i].block.pk, err = keycert.
			NewPrivateKey(x509.PureEd25519)
		if err != nil {
			tb.Fatal(err)
		}
		gen[i].cert.x509, gen[i].block.x509, err = keycert.
			NewX509Certificate(gen[i].pk, &x509.Certificate{
				SignatureAlgorithm: x509.PureEd25519,
				SerialNumber:       big.NewInt(int64(i) + 1),
				DNSNames:           []string{gen[i].name},
				NotBefore:          now,
				NotAfter:           expire,
				IsCA:               true,
				KeyUsage: x509.KeyUsageDigitalSignature |
					x509.KeyUsageCertSign,
			})
		if err != nil {
			tb.Fatal(err)
		}
		gen[i].cert.tls[0], err = keycert.
			NewTLSCertificate(gen[i].block.x509, gen[i].block.pk)
		if err != nil {
			tb.Fatal(err)
		}
	}

	Elog = tb.Error

	sigctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancel(sigctx)
	defer cancel()

	ln, err := ipc.New().Listen()
	if err != nil {
		tb.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		if req.Proto != "HTTP/2.0" {
			w.Write([]byte(req.Proto))
		} else {
			io.Copy(w, req.Body)
		}
	})

	svcfg, err := NewServerCfg(
		gen[sv].name,
		gen[sv].cert.tls[:],
		gen[cl].cert.x509,
	)
	if err != nil {
		tb.Fatal(err)
	}

	svhttps := NewService(svcfg, mux)
	svhttps.Start(&wg, ln, "", "")
	defer svhttps.Shutdown(ctx)

	clcfg, err := NewClientCfg(
		gen[sv].name,
		gen[cl].cert.tls[:],
		gen[sv].cert.x509,
	)
	if err != nil {
		tb.Fatal(err)
	}

	nw := ln.Addr().Network()
	addr := ln.Addr().String()
	auth := NewAuthority(nw, addr)
	dl := &net.Dialer{}
	clhttps := NewClient(clcfg, dl, nw, addr)

	want := new(strings.Builder)
	got := new(strings.Builder)
	for i := uint(0); i < n; i++ {
		want.Reset()
		got.Reset()
		fmt.Fprint(want, "hello world ", i)
		rsp, err := clhttps.Post(auth.Join("echo"), "text/plain",
			strings.NewReader(want.String()))
		if err != nil {
			tb.Error(err)
		} else {
			defer rsp.Body.Close()
			io.Copy(got, rsp.Body)
			gots, wants := got.String(), want.String()
			if gots != wants {
				tb.Errorf("%q", gots)
			}
		}
	}
}
