// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const WWWEchoPort = 8080

func WWWEcho(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<address>:<port>]
Echo paths of http request (default listen at <:{{.Port}}>)
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		if len(path) > 1 && path[1] == "daemon" {
			path[1] = "start"
		}
		return style.Usage(usage, struct {
			Path []string
			Port int
		}{path, WWWEchoPort})
	}

	a := fmt.Sprint(":", WWWEchoPort)
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

	style.Noteln("start", a, "service")
	err := srv.ListenAndServe()
	cancel()
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	style.Noteln("stopped", a, "service")
	return err
}

func WWWPing(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<host>]
Ping echo host.

<host>
    [<ip6>]:<port>
    <ip4>:<port>
    <name>:<port>

The default is 127.0.0.1:{{.Port}}.
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, struct {
			Path []string
			Port int
		}{path, WWWEchoPort})
	}
	url := fmt.Sprint("http://127.0.0.1:", WWWEchoPort, "/hello")
	if len(args) > 0 {
		url = fmt.Sprint("http://", args[0], "/hello")
	}
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		nl := []byte{'\n'}
		io.Copy(w, resp.Body)
		w.Write(nl)
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
	style.Note("done")
	cctx, cancel := context.
		WithTimeout(context.Background(), timeout)
	defer cancel()
	style.Note("shutdown...")
	srv.Shutdown(cctx)
}
