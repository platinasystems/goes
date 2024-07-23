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
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/x509certs"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

type Rest string

// Admins prints the names of authorized VPN administrators.
const Admins Rest = "admins"

// Approve VPN subscription.
const Approve Rest = "approve"

// Certify adds registry to subscriptions.
const Certify Rest = "certify"

// Deny VPN subscription.
const Deny Rest = "deny"

// Disable subscriber admin privilege.
const Disable Rest = "disable"

// Enable subscriber admin privilege.
const Enable Rest = "enable"

// Pending prints requesting subscriber certificates.
const Pending Rest = "pending"

// Ping registry.
const Ping Rest = "ping"

// Revoke VPN subscription.
const Revoke Rest = "revoke"

// Subscribe requests VPN subscription.
const Subscribe Rest = "subscribe"

// Subscribers prints the names of current subscribers.
const Subscribers Rest = "subscribers"

// Unsubscribe from VPN.
const Unsubscribe Rest = "unsubscribe"

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

var restTransport = sync.OnceValues(func() (*http.Transport, error) {
	var cert tls.Certificate
	k, err := keyFile()
	if err != nil {
		return nil, err
	}
	cert.PrivateKey = k.First()
	if cert.PrivateKey == nil {
		return nil, xerrors.Invalid(k.Path)
	}
	c, err := crtFile()
	if err != nil {
		return nil, err
	}
	if cert.Certificate = c.DERs(); cert.Certificate == nil {
		return nil, xerrors.Invalid(c.Path)
	}

	cfg := new(tls.Config)
	cfg.Certificates = append(cfg.Certificates, cert)

	if cfg.RootCAs, err = x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	}

	c.Join(cfg.RootCAs)

	subs, err := subscriptionsFile()
	if err == nil {
		subs.Join(cfg.RootCAs)
	} else if !os.IsNotExist(err) && !ErrIsNOENT(err) {
		return nil, err
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = cfg
	return t, nil
})

func isLocalhost(req *http.Request) bool {
	return strings.HasPrefix(req.URL.Host, "127.0.0.1") ||
		strings.HasPrefix(req.URL.Host, "localhost")
}

func rest(req *http.Request) (*http.Response, error) {
	transport, err := restTransport()
	if err != nil {
		return nil, err
	}
	if isLocalhost(req) {
		transport.TLSClientConfig.InsecureSkipVerify = true
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
		err = xerrors.Invalid("response")
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

func (op Rest) admin(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	opts.reg = DefaultRegistry

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	} else if len(args) == 0 {
		return xerrors.Incomplete("subscriber")
	}

	subscriber := args[0]

	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", string(op))
	q.Set("subscriber", subscriber)
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return err
	}
	resp, err := rest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

// Rest.certify adds registry to subscriptions.
func (op Rest) certify(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Add registry to subscriptions.

{{flags .}}`)

	opts.reg = DefaultRegistry

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	}

	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", string(op))
	clone.RawQuery = q.Encode()

	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return err
	}

	resp, err := rest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.TLS == nil {
		return fmt.Errorf("%v: doesn't support TLS", clone)
	}
	if len(resp.TLS.PeerCertificates) == 0 {
		return fmt.Errorf("%v: no certificates", clone)
	}

	r := bufio.NewReader(os.Stdin)
	w := os.Stdout

	subs, err := subscriptionsFile()
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

func restCheckin(
	ctx context.Context,
	pubder,
	nonce []byte,
	optsvc netip.AddrPort,
) (
	id, via box.Id,
	addr netip.Addr,
	prefix netip.Prefix,
	err error,
) {
	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", "checkin")
	if optsvc.Addr().IsValid() {
		q.Set("service", optsvc.String())
	}
	clone.RawQuery = q.Encode()
	body := new(bytes.Buffer)
	err = xerrors.Mark(pem.Encode(body, &pem.Block{
		Type: "PUBLIC KEY",
		Headers: map[string]string{
			"nonce": hex.EncodeToString(nonce),
		},
		Bytes: pubder,
	}))
	if err != nil {
		return
	}
	req, err := xerrors.MarkResult(http.NewRequestWithContext(ctx,
		http.MethodPut, clone.String(), body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", contextApplicationPKCS8)
	resp, err := rest(req)
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
		case "id:":
			id, err = xerrors.MarkResult(ParseId(val))
		case "via:":
			via, err = xerrors.MarkResult(ParseId(val))
		case "address:":
			addr, err = xerrors.MarkResult(netip.ParseAddr(val))
		case "prefix:":
			prefix, err = xerrors.MarkResult(netip.ParsePrefix(val))
		}
	}
	return
}

func (op Rest) ping(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
RESTful ping registry.

{{flags .}}`)

	opts.reg = DefaultRegistry

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	}

	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", string(op))
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	resp, err := rest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func (obj Rest) show(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
RESTful query and print registry object.

{{flags .}}`)

	opts.reg = DefaultRegistry

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	}

	op := fmt.Sprint("show-", string(obj))

	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", op)
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	resp, err := rest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func (op Rest) subscribe(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	opts.reg = DefaultRegistry

	err := parseOpts(ctx, args)
	if err != nil {
		return err
	}

	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", string(op))
	clone.RawQuery = q.Encode()
	buf := new(bytes.Buffer)
	c, err := crtFile()
	if err != nil {
		return err
	}
	if err = c.Dump(buf); err != nil {
		return err
	}
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), buf)
	if err != nil {
		return xerrors.Mark(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	transport, err := restTransport()
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
		return xerrors.Incomplete("response")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
	} else {
		_, err = io.Copy(os.Stdout, resp.Body)
	}
	return err
}

func restWhoIs(ctx context.Context, qname, qvalue string) (
	*pem.Block, error,
) {
	clone := *opts.regurl
	q := clone.Query()
	q.Set("op", "whois")
	q.Set(qname, qvalue)
	clone.RawQuery = q.Encode()
	req, err := xerrors.MarkResult(http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil))
	if err != nil {
		return nil, err
	}
	resp, err := rest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := xerrors.MarkResult(io.ReadAll(resp.Body))
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(data)
	if blk == nil {
		return blk, xerrors.Invalid("encoding")
	}
	return blk, nil
}

func restWhoIsAddressed(ctx context.Context, addr netip.Addr) (
	*pem.Block, error,
) {
	return restWhoIs(ctx, "address", addr.String())

}

func restWhoIsIdentified(ctx context.Context, id box.Id) (
	*pem.Block, error,
) {
	return restWhoIs(ctx, "id", fmt.Sprint(IdIndex(id)))
}
