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
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const WWWEchoPort = 8080

const WWWEchoUsageTemplate = `
usage: {{.Path}} [<address>:<port>]
Echo paths of http request.

The default is <:{{.Port}}}>.`

func WWWEchoUsageData(path []string) any {
	return struct {
		Path string
		Port int
	}{
		Path: strings.Join(path, ""),
		Port: WWWEchoPort,
	}
}

func WWWEcho(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		path := ctxparm.Strings.In(ctx)
		if len(path) > 1 && path[1] == "daemon" {
			path[1] = "start"
		}
		return usage.Error(WWWEchoUsageTemplate[1:],
			WWWEchoUsageData(path))
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

	w := ctxparm.Writer.In(ctx)
	fmt.Fprintln(w, "start", a, "service")
	err := srv.ListenAndServe()
	cancel()
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	fmt.Fprintln(w, "stopped", a, "service")
	return err
}

const WWWPingUsageTemplate = `
usage: {{.Path}} [<host>]
Ping echo host.

<host>
  [<ip6>]:<port>
  <ip4>:<port>
  <name>:<port>

The default host is 127.0.0.1:{{.Port}}.`

func WWWPing(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		path := ctxparm.Strings.In(ctx)
		return usage.Error(WWWPingUsageTemplate[1:],
			WWWEchoUsageData(path))
	}
	w := ctxparm.Writer.In(ctx)
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
	w := ctxparm.Writer.In(ctx)
	<-ctx.Done()
	fmt.Fprintln(w, "done")
	cctx, cancel := context.
		WithTimeout(context.Background(), timeout)
	defer cancel()
	fmt.Fprintln(w, "shutdown...")
	srv.Shutdown(cctx)
}
