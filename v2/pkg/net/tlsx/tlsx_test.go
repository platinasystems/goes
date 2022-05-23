// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/authorized"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
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

	Elog = tb.Error
	Elogf = tb.Errorf
	Log = tb.Log
	Logf = tb.Logf

	sigctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancel(sigctx)
	defer cancel()

	ln, err := IPC.Listen()
	if err != nil {
		tb.Fatal(err)
	}
	defer ln.Close()

	const sigalg = x509.PureEd25519

	template := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(1),
		SignatureAlgorithm: sigalg,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		Subject: pkix.Name{
			Organization: []string{"Platina", "Systems"},
			CommonName:   "tlsx test",
		},
	}

	pk, pkblk, err := keycert.NewPrivateKey(sigalg)
	if err != nil {
		tb.Fatal(err)
	}

	x509cert, x509blk, err := keycert.NewX509Certificate(pk, &template)
	if err != nil {
		tb.Fatal(err)
	}

	tlscert, err := keycert.NewTLSCertificate(x509blk, pkblk)
	if err != nil {
		tb.Fatal(err)
	}

	cert.Preload(tlscert, x509cert)
	authorized.Preload()

	certs.Clients.TestLoad()
	certs.Exchanges.TestLoad()

	wg.Add(1)
	go Exchange(ctx, &wg, ln)

	wg.Add(1)
	go Accept(ctx, &wg, "", selection.Map{
		"cat":  cat.Func,
		"echo": echo.Func,
	}.Select)

	cn, err := DialAndHandshake(ctx, "")
	if err != nil {
		tb.Fatal(err)
	}
	defer cn.Close()

	got := new(strings.Builder)

	err = Req(ctx, cn, nil, got, "connect", cert.SKI.String())
	if err != nil {
		tb.Fatal(err)
	}
	if svc := got.String(); svc != SVC {
		tb.Fatal(svc)
	}

	if t, ok := tb.(*testing.T); ok {
		t.Run("echo", func(t *testing.T) {
			const want = "hello world\n"
			got.Reset()
			err := Req(ctx, cn, nil, got,
				"echo", "hello", "world")
			if err != nil {
				t.Fatal(err)
			}
			if gots := got.String(); gots != want {
				t.Errorf("%q", got)
			} else if false {
				t.Logf("%q", got)
			}
		})
		t.Run("cat", func(t *testing.T) {
			const want = "sample input"
			got.Reset()
			err := Req(ctx, cn, strings.NewReader(want), got,
				"cat", "-")
			if err != nil {
				t.Fatal(err)
			}
			if gots := got.String(); gots != want {
				t.Errorf("%q", got)
			}
		})
	} else {
		want := new(strings.Builder)
		for i := uint(0); i < n; i++ {
			want.Reset()
			fmt.Fprintln(want, "hello", "world", i)
			got.Reset()
			err := Req(ctx, cn, nil, got,
				"echo", "hello", "world", i)
			if err != nil {
				t.Fatal(err)
			}
			gots, wants := got.String(), want.String()
			if gots != wants {
				t.Errorf("%q", gots)
			}
		}
	}
}
