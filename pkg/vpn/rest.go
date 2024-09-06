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
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

func RestAdmin(ctx context.Context, args []string) error {
	var rest rest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	rest.xFlag = AdminFlag()
	err := rest.flags(args)
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("subscriber")
	}

	clone := *rest.url
	q := clone.Query()
	q.Set("op", xflag.LastName(flag.CommandLine))
	q.Set("subscriber", args[0])
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return err
	}
	resp, err := rest.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

// Rest.certify adds the peer certificate to CONFIG_HOME/registry.pem.
func RestCertify(ctx context.Context, args []string) error {
	var rest rest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	rest.xFlag = GuestFlag()
	err := rest.flags(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	if len(args) == 0 {
		return xerrors.Incomplete("registry")
	}
	rest.url, err = url.Parse(args[0])
	if err != nil {
		return err
	}
	rest.tp.TLSClientConfig.InsecureSkipVerify = true

	clone := *rest.url
	q := clone.Query()
	q.Set("op", xflag.LastName(flag.CommandLine))
	clone.RawQuery = q.Encode()

	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return err
	}

	resp, err := rest.do(req)
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

	if t, err := CertificatesTemplate(); err != nil {
		return err
	} else if err = t.Execute(w, resp.TLS.PeerCertificates); err != nil {
		return err
	}

	fmt.Fprintf(w, `Enter "yes" to write above to %s: `, *rest.rFlag)
	s, err := r.ReadString('\n')
	if err != nil && strings.TrimSpace(s) != "yes" {
		return err
	}

	blk := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: resp.TLS.PeerCertificates[0].Raw,
	}
	wc, err := os.OpenFile(*rest.rFlag, oCreate, 0644)
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func RestPing(ctx context.Context, args []string) error {
	var rest rest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
RESTful ping registry.

{{flags .}}`)

	rest.xFlag = AdminFlag()
	err := rest.flags(args)
	if err != nil {
		return err
	}

	clone := *rest.url
	q := clone.Query()
	q.Set("op", xflag.LastName(flag.CommandLine))
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	resp, err := rest.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func RestShow(ctx context.Context, args []string) error {
	var rest rest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`)

	rest.xFlag = AdminFlag()
	err := rest.flags(args)
	if err != nil {
		return err
	}

	clone := *rest.url
	q := clone.Query()
	q.Set("op", "show")
	q.Set("obj", xflag.LastName(flag.CommandLine))
	for i, arg := range flag.Args() {
		q.Set(fmt.Sprint("arg", i), arg)
	}
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	resp, err := rest.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func RestSubscribe(ctx context.Context, args []string) error {
	var rest rest

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	rest.xFlag = GuestFlag()
	err := rest.flags(args)
	if err != nil {
		return err
	}

	clone := *rest.url
	q := clone.Query()
	q.Set("op", xflag.LastName(flag.CommandLine))
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	cl := &http.Client{Transport: rest.tp}
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

type rest struct {
	kFlag,
	rFlag,
	vpnFlag,
	xFlag *string
	crt,
	reg *x509.Certificate
	sig *Signatures
	url *url.URL
	tp  *http.Transport
}

func (rest *rest) checkin(
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
	const period = 5 * time.Second
	for try := 1; true; try++ {
		var resp *http.Response
		verbose.Println("checkin try", try)
		id, via, addr, prefix, resp, err = rest.tryCheckin(
			ctx, pubder, nonce, optsvc,
		)
		if err == nil {
			return
		} else if resp == nil {
			return
		} else if resp.StatusCode != http.StatusTooEarly {
			return
		} else if try == 10 {
			err = xerrors.Unavailable("exchange")
			return
		}
		time.Sleep(period)
	}
	return
}

func (rest *rest) do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Host, "127.0.0.1") ||
		strings.HasPrefix(req.URL.Host, "localhost") {
		rest.tp.TLSClientConfig.InsecureSkipVerify = true
	}
	cl := &http.Client{Transport: rest.tp}
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
		if resp.StatusCode != http.StatusTooEarly {
			resp = nil
		}
	}
	return resp, err
}

func (rest *rest) flags(args []string) error {
	rest.kFlag = KeyFlag()
	rest.rFlag = RegistryFlag()
	rest.vpnFlag = VpnFlag()

	err := qvFlags(args)
	if err != nil {
		return err
	}

	cs, err := certificates(*rest.xFlag)
	if err != nil {
		return err
	} else if len(cs) == 0 {
		return xerrors.Invalid(*rest.xFlag)
	} else {
		rest.crt = cs[0]
	}
	verbose.Println("crt file:", *rest.xFlag)

	if rest.sig, err = NewSignatures(*rest.kFlag); err != nil {
		return err
	}
	verbose.Println("sig file:", *rest.kFlag)

	cs, err = certificates(*rest.rFlag)
	if err != nil {
		if !strings.HasSuffix(flag.CommandLine.Name(), "certify") ||
			!os.IsNotExist(err) {
			return err
		}
	} else if len(cs) == 0 {
		return xerrors.Invalid(*rest.rFlag)
	} else {
		rest.reg = cs[0]
	}

	if err = rest.xregurl(); err != nil {
		return err
	}
	if len(*rest.vpnFlag) > 0 {
		rest.url = rest.url.JoinPath(*rest.vpnFlag)
	}
	verbose.Println("url file:", rest.url)

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{rest.crt.Raw},
				PrivateKey:  rest.sig.First(),
			},
		},
	}

	if rcas, err := x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	} else {
		cfg.RootCAs = rcas
	}

	cfg.RootCAs.AddCert(rest.reg)

	rest.tp = http.DefaultTransport.(*http.Transport).Clone()
	rest.tp.TLSClientConfig = cfg
	return nil
}

func (rest *rest) tryCheckin(
	ctx context.Context,
	pubder,
	nonce []byte,
	optsvc netip.AddrPort,
) (
	id, via box.Id,
	addr netip.Addr,
	prefix netip.Prefix,
	resp *http.Response,
	err error,
) {
	clone := *rest.url
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
	if resp, err = rest.do(req); err != nil {
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

func (rest *rest) whois(ctx context.Context, qname, qvalue string) (
	*pem.Block, error,
) {
	clone := *rest.url
	q := clone.Query()
	q.Set("op", "whois")
	q.Set(qname, qvalue)
	clone.RawQuery = q.Encode()
	req, err := xerrors.MarkResult(http.
		NewRequestWithContext(ctx, http.MethodGet, clone.String(), nil))
	if err != nil {
		return nil, err
	}
	resp, err := rest.do(req)
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

func (rest *rest) whoisAddressed(ctx context.Context, addr netip.Addr) (
	*pem.Block, error,
) {
	return rest.whois(ctx, "address", addr.String())

}

func (rest *rest) whoisIdentified(ctx context.Context, id box.Id) (
	*pem.Block, error,
) {
	return rest.whois(ctx, "id", fmt.Sprint(IdIndex(id)))
}

// eXtract registry url from its certificate.
func (rest *rest) xregurl() error {
	if rest.reg == nil {
		return xerrors.Unavailable("registry certificate")
	}
	if len(rest.reg.URIs) > 0 {
		rest.url = rest.reg.URIs[0]
		if len(rest.url.Scheme) == 0 {
			rest.url.Scheme = "https"
		}
		return nil
	}
	if len(rest.reg.DNSNames) == 0 {
		return xerrors.Invalid("no registry URL or DNS")
	}
	var err error
	s := fmt.Sprint("https://", rest.reg.DNSNames[0], ":8003")
	rest.url, err = url.Parse(s)
	return err
}
