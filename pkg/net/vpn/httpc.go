// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509certs"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"golang.org/x/sys/unix"
)

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

var httpTransport = sync.OnceValues(func() (*http.Transport, error) {
	var cert tls.Certificate
	k, err := vpnKeyFile()
	if err != nil {
		return nil, err
	}
	cert.PrivateKey = k.First()
	if cert.PrivateKey == nil {
		return nil, fmt.Errorf("%s: %w", k.Path, ErrNoPrivateKeys)
	}
	c, err := vpnCrtFile()
	if err != nil {
		return nil, err
	}
	if cert.Certificate = c.DERs(); cert.Certificate == nil {
		return nil, fmt.Errorf("%s: %w", c.Path, ErrNoCertificates)
	}

	cfg := new(tls.Config)
	cfg.Certificates = append(cfg.Certificates, cert)

	if cfg.RootCAs, err = x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	}

	c.Join(cfg.RootCAs)

	subs, err := vpnSubscriptionsFile()
	if err == nil {
		subs.Join(cfg.RootCAs)
	} else if !os.IsNotExist(err) && !errors.Is(err, unix.ENOENT) {
		return nil, err
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = cfg
	return t, nil
})

func httpDo(req *http.Request) (*http.Response, error) {
	transport, err := httpTransport()
	if err != nil {
		return nil, err
	}
	cl := &http.Client{Transport: transport}
	resp, err := cl.Do(req)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, berr := io.ReadAll(resp.Body)
			if berr == nil && len(body) > 0 {
				err = fmt.Errorf("%w, %s", err, body)
			}
			resp = nil
		}
	} else if resp == nil {
		err = egress.Mark(ErrInvalid)
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, berr := io.ReadAll(resp.Body)
		if berr != nil || len(body) == 0 {
			err = errors.New(resp.Status)
		} else {
			err = fmt.Errorf("%s, %s", resp.Status, body)
		}
		resp = nil
	}
	return resp, err
}

func httpAdmin(ctx context.Context, args []string) error {
	branch := goes.ContextBranch(ctx)
	op := branch[len(branch)-1]

	usage := map[string]string{
		"approve": `
usage: {{branch .}} ` + vpnRegistrySyntax + ` <name>
Approve named subscriber.`,
		"deny": `
usage: {{branch .}} ` + vpnRegistrySyntax + ` <name>
Deny named subscriber.`,
		"disable": `
usage: {{branch .}} ` + vpnRegistrySyntax + ` <name>
Disable admin privilege for named subscriber.`,
		"enable": `
usage: {{branch .}} ` + vpnRegistrySyntax + ` <name>
Enable admin privilege for named subscriber.`,
		"unsubscribe": `
usage: {{branch .}} ` + vpnRegistrySyntax + ` <name>
Unsubscribe named subscriber.`,
	}[op]

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}
	if nargs := len(args); nargs == 0 {
		return ErrNeedServer
	} else if nargs == 1 {
		return errors.New("need <name>")
	}
	svr, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	q := svr.Query()
	q.Set("op", op)
	q.Set("subscriber", args[1])
	svr.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		svr.String(), nil)
	if err != nil {
		return egress.Mark(err)
	}
	resp, err := httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(goes.ContextStdout(ctx), resp.Body)
	return err
}

func httpCertify(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} ` + vpnRegistrySyntax + `
Fetch registry certificate.`

	branch := goes.ContextBranch(ctx)
	op := branch[len(branch)-1]

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}

	if len(args) < 1 {
		return ErrNeedServer
	}

	svr, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	q := svr.Query()
	q.Set("op", op)
	svr.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		svr.String(), nil)
	if err != nil {
		return err
	}

	resp, err := httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.TLS == nil {
		return fmt.Errorf("%v: doesn't support TLS", svr)
	}
	if len(resp.TLS.PeerCertificates) == 0 {
		return fmt.Errorf("%v: no certificates", svr)
	}

	r := bufio.NewReader(goes.ContextStdin(ctx))
	w := goes.ContextStdout(ctx)

	subs, err := vpnSubscriptionsFile()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var peercerts []*x509.Certificate
	for _, peer := range resp.TLS.PeerCertificates {
		if !subs.Has(peer) {
			peercerts = append(peercerts, peer)
		}
	}
	if len(peercerts) == 0 {
		fmt.Fprintln(w, "no new certificates")
		return nil
	}

	t, err := x509certs.Template()
	if err != nil {
		return err
	}
	t.Execute(w, peercerts)

	fmt.Fprintf(w, `Enter "yes" to append above to %s: `, subs.Path)
	s, err := r.ReadString('\n')
	if err != nil && strings.TrimSpace(s) != "yes" {
		return err
	}

	for _, peer := range peercerts {
		pb := &pem.Block{
			Type:    "CERTIFICATE",
			Headers: map[string]string{},
			Bytes:   peer.Raw,
		}
		if err = subs.Add(pb, peer); err != nil {
			break
		} else {
			io.Copy(w, resp.Body)
			break
		}
	}

	return err
}

func httpCheckin(
	ctx context.Context,
	svr *url.URL,
	pubder,
	nonce []byte,
	optsvc ...netip.AddrPort,
) (
	lbl, via label.Label,
	addr netip.Addr,
	prefix netip.Prefix,
	err error,
) {
	clone := *svr
	q := clone.Query()
	q.Set("op", "checkin")
	if len(optsvc) > 0 {
		q.Set("service", optsvc[0].String())
	}
	clone.RawQuery = q.Encode()
	body := new(bytes.Buffer)
	err = egress.Mark(pem.Encode(body, &pem.Block{
		Type: "PUBLIC KEY",
		Headers: map[string]string{
			"nonce": hex.EncodeToString(nonce),
		},
		Bytes: pubder,
	}))
	if err != nil {
		return
	}
	req, err := egress.MarkResult(http.NewRequestWithContext(ctx,
		http.MethodPut, clone.String(), body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", contextApplicationPKCS8)
	resp, err := httpDo(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanWords)
	for err == nil && scanner.Scan() {
		key := scanner.Text()
		if !scanner.Scan() {
			break
		}
		val := scanner.Text()
		switch key {
		case "label:":
			lbl, err = egress.MarkResult(label.Parse(val))
		case "via:":
			via, err = egress.MarkResult(label.Parse(val))
		case "address:":
			addr, err = egress.MarkResult(netip.ParseAddr(val))
		case "prefix:":
			prefix, err = egress.MarkResult(netip.ParsePrefix(val))
		}
	}
	return
}

func httpPing(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} ` + vpnRegistrySyntax + `
Ping registry.`

	branch := goes.ContextBranch(ctx)
	op := branch[len(branch)-1]

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}
	if len(args) < 1 {
		return ErrNeedServer
	}
	svr, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	q := svr.Query()
	q.Set("op", op)
	svr.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		svr.String(), nil)
	if err != nil {
		return egress.Mark(err)
	}
	resp, err := httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(goes.ContextStdout(ctx), resp.Body)
	return err
}

func httpShow(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} ` + vpnRegistrySyntax + `
Retrieve and print registry object.`

	branch := goes.ContextBranch(ctx)
	op := fmt.Sprint("show-", branch[len(branch)-1])

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}
	if len(args) < 1 {
		return ErrNeedServer
	}
	svr, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	q := svr.Query()
	q.Set("op", op)
	svr.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		svr.String(), nil)
	if err != nil {
		return egress.Mark(err)
	}
	resp, err := httpDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(goes.ContextStdout(ctx), resp.Body)
	return err
}

func httpSubscribe(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} ` + vpnRegistrySyntax + `
Subscribe to registry.`

	branch := goes.ContextBranch(ctx)
	op := branch[len(branch)-1]

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}
	if len(args) < 1 {
		return ErrNeedServer
	}
	svr, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	q := svr.Query()
	q.Set("op", op)
	svr.RawQuery = q.Encode()
	buf := new(bytes.Buffer)
	c, err := vpnCrtFile()
	if err != nil {
		return err
	}
	if err = c.Dump(buf); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		svr.String(), buf)
	if err != nil {
		return egress.Mark(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	transport, err := httpTransport()
	if err != nil {
		return err
	}
	transport.TLSClientConfig.InsecureSkipVerify = true
	cl := &http.Client{Transport: transport}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrInvalid
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
	} else {
		_, err = io.Copy(goes.ContextStdout(ctx), resp.Body)
	}
	return err
}

func httpWhoIs(ctx context.Context, svr *url.URL, qname, qvalue string) (
	*pem.Block, error,
) {
	q := svr.Query()
	q.Set("op", "whois")
	q.Set(qname, qvalue)
	svr.RawQuery = q.Encode()
	req, err := egress.MarkResult(http.
		NewRequestWithContext(ctx, http.MethodGet, svr.String(), nil))
	if err != nil {
		return nil, err
	}
	resp, err := httpDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := egress.MarkResult(io.ReadAll(resp.Body))
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(data)
	if blk == nil {
		return blk, egress.Mark(ErrNotPEM)
	}
	return blk, nil
}

func httpWhoIsAddressed(ctx context.Context, svr *url.URL, addr netip.Addr) (
	*pem.Block, error,
) {
	return httpWhoIs(ctx, svr, "address", addr.String())

}

func httpWhoIsLabelled(ctx context.Context, svr *url.URL, lbl label.Label) (
	*pem.Block, error,
) {
	return httpWhoIs(ctx, svr, "label", lbl.String())
}
