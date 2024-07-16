// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const WWWEchoPort = ":8080"

func WWWEcho(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [<address>:<port>]
WWW echo server that responds with path of http request.

Default: “`+WWWEchoPort+`”
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	args = flag.Args()

	a := WWWEchoPort
	if len(args) > 0 {
		a = args[0]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path[1:])
	})

	srv := &http.Server{Addr: a}
	cctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go wwwEchoShutdown(cctx, &wg, srv)

	fmt.Println("start", a, "service")
	defer fmt.Println("stopped", a, "service")

	err = srv.ListenAndServe()
	cancel()
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

func WWWPing(ctx context.Context, args []string) error {
	const nl = "\n"

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [host]
Ping WWW echo server.

Host:
  - [<ip6>]:<port>
  - <ip4>:<port>
  - <name>:<port>

Default: “127.0.0.1`+WWWEchoPort+`”
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	args = flag.Args()

	url := fmt.Sprint("http://127.0.0.1", WWWEchoPort, "/hello")
	if len(args) > 0 {
		url = fmt.Sprint("http://", args[0], "/hello")
	}
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		io.Copy(os.Stdout, resp.Body)
		os.Stdout.WriteString(nl)
	}
	return err
}

func wwwEchoShutdown(
	ctx context.Context,
	wg *sync.WaitGroup,
	srv *http.Server,
) {
	const timeout = 10 * time.Second
	defer wg.Done()
	<-ctx.Done()
	fmt.Println("done")
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	fmt.Println("shutdown...")
	srv.Shutdown(cctx)
}
