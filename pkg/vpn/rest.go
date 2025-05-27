// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

const RestTrailerVcsRevision = "Vcs-Revision"

const (
	RestService    = "service"
	RestSubscriber = "subscriber"
)

const (
	RestOp            = "op"
	RestOpApprove     = "approve"
	RestOpCertify     = "certify"
	RestOpCheckin     = "checkin"
	RestOpDeny        = "deny"
	RestOpDump        = "dump"
	RestOpPing        = "ping"
	RestOpReload      = "reload"
	RestOpShow        = "show"
	RestOpSubscribe   = "subscribe"
	RestOpUnsubscribe = "unsubscribe"
	RestOpWhois       = "whois"
)

const (
	RestDumpSubscribers = "subscribers"
)

const (
	RestShowActive      = "active"
	RestShowAddress     = "address"
	RestShowAdmins      = "admins"
	RestShowHosts       = "hosts"
	RestShowPending     = "pending"
	RestShowSubscriber  = "subscriber"
	RestShowTenant      = "tenant"
	RestShowVcsModified = "vcs.modified"
	RestShowVcsRevision = "vcs.revision"
)

const (
	RestWhoisAddress = "address"
	RestWhoisId      = "id"
	RestWhoisName    = "name"
)

var ErrVcsRevMismatch = errors.New("must upgrade")

// If [http.Response.StatusCode] != [http.StatusOK}, return non-nil, error
// encapsulating the [http.Response.Body]; otherwise, return nil.
func AssertOK(rsp *http.Response) error {
	if rsp.StatusCode == http.StatusOK {
		return nil
	}
	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	if err != nil || len(b) == 0 {
		err = errors.New(rsp.Status)
	} else {
		err = fmt.Errorf("%s, %s", rsp.Status, b)
	}
	return err
}

func RestAdmin(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("subscriber")
	}
	if err = rest.config(); err != nil {
		return err
	}
	_, err = rest.request(ctx, os.Stdout, http.MethodPut,
		RestOp, xflag.LastName(flag.CommandLine),
		RestSubscriber, args[0])
	return err
}

// Rest.certify adds the peer certificate to [Reg].
func RestCertify(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("registry")
	}
	rest.url, err = url.Parse(args[0])
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}

	tp := rest.Client.Transport.(*http.Transport)
	sv := tp.TLSClientConfig.InsecureSkipVerify
	defer func() {
		tp.TLSClientConfig.InsecureSkipVerify = sv
	}()
	tp.TLSClientConfig.InsecureSkipVerify = true

	rsp, err := rest.request(ctx, os.Stdout, http.MethodPut,
		RestOp, RestOpCertify)
	if err != nil {
		return err
	}
	if rsp.TLS == nil {
		return fmt.Errorf("%v: doesn't support TLS", rest.url)
	}
	if len(rsp.TLS.PeerCertificates) == 0 {
		return fmt.Errorf("%v: no certificates", rest.url)
	}

	r := bufio.NewReader(os.Stdin)
	w := os.Stdout

	if t, err := CertificatesTemplate(); err != nil {
		return err
	} else if err = t.Execute(w, rsp.TLS.PeerCertificates); err != nil {
		return err
	}

	rfn := filepath.Join(vpnConfigDir, vpnRegistry)
	fmt.Fprintf(w, `Enter "yes" to write above to %s: `, rfn)
	s, err := r.ReadString('\n')
	if err != nil && strings.TrimSpace(s) != "yes" {
		return err
	}

	blk := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rsp.TLS.PeerCertificates[0].Raw,
	}
	wc, err := os.OpenFile(rfn, oCreate, 0644)
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func RestGet(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [filename]
Get or list registry file(s).

{{flags .}}`)

	var rest rest

	rest.defineFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}
	args = flag.CommandLine.Args()
	_, err = rest.request(ctx, os.Stdout, http.MethodGet, args...)
	if errors.Is(err, ErrVcsRevMismatch) {
		err = nil
	}
	return err
}

func RestPing(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful ping registry.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}
	_, err = rest.request(ctx, os.Stdout, http.MethodGet,
		RestOp, RestOpPing)
	return err
}

func RestReload(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful reload registry configuration.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}
	_, err = rest.request(ctx, os.Stdout, http.MethodPut,
		RestOp, RestOpReload)
	return err
}

func RestShow(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}
	_, err = rest.request(ctx, os.Stdout, http.MethodGet,
		RestOp, RestOpShow,
		RestOpShow, xflag.LastName(flag.CommandLine))
	if errors.Is(err, ErrVcsRevMismatch) {
		err = nil
	}
	return err
}

func RestSubscribe(ctx context.Context, args []string) error {
	var rest rest

	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	rest.defineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if err = rest.config(); err != nil {
		return err
	}

	clone := *rest.url
	q := clone.Query()
	q.Set(RestOp, RestOpSubscribe)
	clone.RawQuery = q.Encode()
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	rsp, err := rest.Client.Do(req)
	if err != nil {
		return err
	}
	if err = AssertOK(rsp); err != nil {
		return err
	}
	defer rsp.Body.Close()
	_, err = io.Copy(os.Stdout, rsp.Body)
	return err
}

type rest struct {
	bufs sync.Pool
	crt,
	reg *x509.Certificate
	sig *Signatures
	url *url.URL
	http.Client
	vcsrev string
}

func (rest *rest) alloc() *bytes.Buffer {
	return rest.bufs.Get().(*bytes.Buffer)
}

func (*rest) defineFlags() {
	defineCert()
	defineConfigDir()
	defineRegistry()
	defineSig()
	defineVPN()
}

func (rest *rest) free(buf *bytes.Buffer) {
	buf.Reset()
	rest.bufs.Put(buf)
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
		xlog.Info.Println("checkin try", try)
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

func (rest *rest) config() error {
	rest.bufs.New = func() any { return new(bytes.Buffer) }
	rest.vcsrev = xprogram.VcsRevision.String()
	fn := filepath.Join(vpnConfigDir, vpnCert)
	cs, err := certificates(fn)
	if err != nil {
		return err
	} else if len(cs) == 0 {
		return xerrors.Invalid(fn)
	} else {
		rest.crt = cs[0]
	}

	rest.sig, err = NewSignatures(filepath.Join(vpnConfigDir, vpnSig))
	if err != nil {
		return err
	}

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

	if cl := flag.CommandLine.Name(); !strings.HasSuffix(cl, "certify") {
		rfn := filepath.Join(vpnConfigDir, vpnRegistry)
		cs, err = certificates(rfn)
		if err != nil {
			return err
		}
		rest.reg = cs[0]
		if err = rest.xregurl(); err != nil {
			return err
		}
		cfg.RootCAs.AddCert(rest.reg)
		if len(vpnVPN) > 0 {
			rest.url = rest.url.JoinPath(vpnVPN)
		}
	}

	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.TLSClientConfig = cfg
	if strings.HasPrefix(rest.url.Host, "127.0.0.1") ||
		strings.HasPrefix(rest.url.Host, "localhost") {
		tp.TLSClientConfig.InsecureSkipVerify = true
	}
	rest.Client.Transport = tp
	return nil
}

// - Without args, copy registry virtual directory listing.
// - With one arg, copy registry file.
// - With one or more (key, value) pairs, RESTful registry query response.
func (rest *rest) request(
	ctx context.Context, w io.Writer, method string, args ...string,
) (*http.Response, error) {
	clone := *rest.url
	switch len(args) {
	case 0:
		clone.Path = ""
	case 1:
		clone.Path = args[0]
	default:
		q := clone.Query()
		for i := 0; i < len(args); i += 2 {
			q.Set(args[i], args[i+1])
		}
		clone.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, clone.String(), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := rest.Do(req)
	if err == nil {
		if err = AssertOK(rsp); err == nil {
			defer rsp.Body.Close()
			_, err = io.Copy(w, rsp.Body)
		}
		if rsp.Trailer.Get(RestTrailerVcsRevision) != rest.vcsrev {
			err = ErrVcsRevMismatch
		}
	}
	return rsp, err
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
	rsp *http.Response,
	err error,
) {
	clone := *rest.url
	q := clone.Query()
	q.Set(RestOp, RestOpCheckin)
	if optsvc.Addr().IsValid() {
		q.Set(RestService, optsvc.String())
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
	if rsp, err = rest.Do(req); err != nil {
		return
	}
	if err = AssertOK(rsp); err != nil {
		return
	}
	defer rsp.Body.Close()
	scanner := bufio.NewScanner(rsp.Body)
	scanner.Split(bufio.ScanWords)
	for err == nil && scanner.Scan() {
		key := scanner.Text()
		if !scanner.Scan() {
			break
		}
		val := scanner.Text()
		switch key {
		case "id:":
			id, err = xerrors.MarkResult(box.ParseId(val))
		case "via:":
			via, err = xerrors.MarkResult(box.ParseId(val))
		case "address:":
			addr, err = xerrors.MarkResult(netip.ParseAddr(val))
		case "prefix:":
			prefix, err = xerrors.MarkResult(netip.ParsePrefix(val))
		}
	}
	return
}

// Forward cloned [http.Response.Body] to [RestOpWhois] [Rest.Request] of
// [netip.Addr] or [box.ID] value.
func (rest *rest) whois(
	ctx context.Context, ch chan<- *bytes.Buffer, value any,
) error {
	buf := rest.alloc()
	key := RestWhoisId
	if _, isAddr := value.(netip.Addr); isAddr {
		key = RestWhoisAddress
	}
	_, err := rest.request(ctx, buf, http.MethodGet,
		RestOp, RestOpWhois,
		key, fmt.Sprint(value))
	if err != nil || !xcontext.Queue(ctx, ch, buf) {
		rest.free(buf)
	}
	return err
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
