// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package wget

import (
	"bufio"
	"context"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
)

const WgetUsage = `
usage: {{.Name}} [flags] <url>
A non-interactive network downloader.

{{flags .}}`

const wgetRel = "../"

var (
	Wget_B xflag.URL
	Wget_i,
	Wget_O string
	Wget_P = "."
)

var WgetFlags = xflag.Labels{
	xmain.ConfigFlag,
	xdnsdoh.ConfigFlag,
	{"i", "Read URLs from named file or stdin if '-'.", &Wget_i},
	{"B", `
URL to resolve input file's "../" prefaced relative links.`[1:], &Wget_B},
	{"O", `
Concentate downloads to named file, or stdout if "-",
instead of each to the base path of the url.`[1:], &Wget_O},
	{"P", "Directory prefix of downloaded files.", &Wget_P},
}

func Wget(ctx context.Context, args []string) error {
	xflag.TemplateUsage(WgetUsage)
	err := WgetFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	urls, err := wgetURLs(flag.Args())
	if err != nil {
		return err
	} else if len(urls) == 0 {
		return xerrors.Incomplete("URL(s)")
	}

	client, err := wgetClient()
	if err != nil {
		return err
	}

	return wgetFetch(ctx, client, urls)
}

func wgetClient() (*http.Client, error) {
	resolver, err := xdnsdoh.Resolver()
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  resolver,
	}
	tp, err := cert.NewTransport()
	if err != nil {
		return nil, err
	}
	tp.DialContext = dialer.DialContext
	return &http.Client{Transport: tp}, nil
}

func wgetFetch(
	ctx context.Context,
	client *http.Client,
	urls []*url.URL,
) (err error) {
	var (
		o   io.WriteCloser
		req *http.Request
		rsp *http.Response
	)
	if len(Wget_O) > 0 {
		if Wget_O == "-" {
			o = os.Stdout
		} else {
			if strings.IndexRune(Wget_O, os.PathSeparator) != 0 {
				Wget_O = filepath.Join(Wget_P, Wget_O)
			}
			if o, err = os.Create(Wget_O); err != nil {
				return
			}
			defer o.Close()
		}
	}

	for _, u := range urls {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet,
			u.String(), http.NoBody)
		if err != nil {
			break
		}
		if rsp, err = client.Do(req); err != nil {
			break
		}
		if rsp == nil {
			err = xerrors.Invalid("response")
			break
		}
		if o != nil {
			_, err = io.Copy(o, rsp.Body)
		} else {
			var f io.WriteCloser
			fn := path.Base(u.Path)
			if len(fn) == 0 || fn == "/" {
				f, err = os.CreateTemp(Wget_P, "Wget_*")
			} else {
				f, err = os.Create(filepath.Join(Wget_P, fn))
			}
			if err == nil {
				_, err = io.Copy(f, rsp.Body)
				f.Close()
			}
			rsp.Body.Close()
		}
		if err != nil {
			break
		}
	}
	return
}

func wgetURLs(args []string) (urls []*url.URL, err error) {
	var u *url.URL
	fqdn := func(*url.URL) {}
	if xdnsdoh.HasSearch() {
		fqdn = func(u *url.URL) {
			host, port := xdnsdoh.FQDN(u.Hostname()), u.Port()
			if len(port) > 0 {
				u.Host = net.JoinHostPort(host, port)
			} else {
				u.Host = host
			}
		}
	}
	for _, s := range args {
		if u, err = url.Parse(s); err != nil {
			return
		}
		if !u.IsAbs() {
			u.Scheme = "https"
		} else if !strings.HasPrefix(u.Scheme, "http") {
			err = xerrors.Unsupported(u.Scheme)
			return
		}
		if len(u.Host) == 0 {
			u.Host = "localhost"
		}
		fqdn(u)
		urls = append(urls, u)
	}
	if len(Wget_i) == 0 {
		return urls, nil
	}
	burl := Wget_B.URL()
	burlIsAbs := burl.IsAbs()
	if burlIsAbs && !strings.HasPrefix(burl.Scheme, "http") {
		err = xerrors.Unsupported(burl.Scheme)
		err = xerrors.Label(err, "flag", "B")
		return
	}
	var i *os.File
	if Wget_i == "-" {
		i = os.Stdin
	} else if i, err = os.Open(Wget_i); err != nil {
		return
	} else {
		defer i.Close()
	}
	sc := bufio.NewScanner(i)
	for lno := 1; err == nil && sc.Scan(); lno++ {
		line := sc.Text()
		if strings.HasPrefix(line, wgetRel) {
			if !burlIsAbs {
				err = xerrors.Incomplete("B")
				err = xerrors.Label(err, "flag")
				return
			}
			u = new(url.URL)
			*u = *burl
			u.JoinPath(strings.TrimPrefix(line, wgetRel))
			urls = append(urls, u)
		} else if u, err = url.Parse(line); err != nil {
			err = xerrors.Label(err, Wget_i, lno)
		} else {
			urls = append(urls, u)
		}
	}
	return
}
