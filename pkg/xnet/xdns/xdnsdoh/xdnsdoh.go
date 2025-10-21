// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DNS Over HTTPS
package xdnsdoh

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
)

var (
	URL     string
	URLFlag = xflag.Label{"doh-url", `
DNS Over HTTPS server URL.
(or $<main>_DOH_URL, $DOH_URL, $CF_DNS_RESOLVER_URL)`[1:], func() any {
		if s, ok := xmain.LookupEnv("DOH_URL"); ok {
			URL = s
		} else if s, ok = os.LookupEnv("DOH_URL"); ok {
			URL = s
		} else if s, ok = os.LookupEnv("CF_DNS_RESOLVER_URL"); ok {
			URL = s
		}
		return &URL
	}}
)

// From [x509.SystemCertPool]:
//
//	On Unix systems other than macOS the environment variables
//	SSL_CERT_FILE and SSL_CERT_DIR can be used to override the system
//	default locations for the SSL certificate file and SSL certificate
//	files directory, respectively.  The latter can be a colon-separated
//	list.
//
// Or enable “insecureSkipVerify”.
func New(insecureSkipVerify bool, url string) xdns.Asker {
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: insecureSkipVerify,
	}
	if rcas, err := x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	} else {
		cfg.RootCAs = rcas
	}
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.TLSClientConfig = cfg
	return doh{&http.Client{Transport: tp}, url}
}

func Asker(cl *http.Client, url string) xdns.Asker {
	return doh{cl, url}
}

type doh struct {
	*http.Client
	url string
}

func (doh doh) Ask(ctx context.Context, b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, doh.url, r)
	if err != nil {
		return b[:0], err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	rsp, err := doh.Do(req)
	if err != nil {
		return b[:0], err
	}
	if rsp == nil {
		return b[:0], errors.New("nil")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		sb := new(strings.Builder)
		io.Copy(sb, rsp.Body)
		if sb.Len() == 0 {
			return b[:0], errors.New(rsp.Status)
		}
		return b[:0], fmt.Errorf("%s, %s", rsp.Status, sb.String())
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, rsp.Body)
	return buf.Bytes(), nil
}
