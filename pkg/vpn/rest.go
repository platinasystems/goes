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
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

const (
	RestUnixMicroStart = "X-Unix-Micro-Start"

	RestVcsRevision = "X-Vcs-Revision"
)

const (
	RestPathApprove = "/approve"

	RestPathCertify = "/certify"

	RestPathCheckinExchange = "/checkin/exchange"
	RestPathCheckinGuest    = "/checkin/guest"

	RestPathDeny = "/deny"

	RestPathDnsQuery = "/dns-query"

	RestPathDumpSubscribers = "/dmup/subscribers"

	RestPathPing = "/ping"

	RestPathReload = "/reload"

	RestPathShowAddress    = "/show/address"
	RestPathShowAdmins     = "/show/admins"
	RestPathShowExchanges  = "/show/exchanges"
	RestPathShowGuests     = "/show/guests"
	RestPathShowHosts      = "/show/hosts"
	RestPathShowPending    = "/show/pending"
	RestPathShowPrefix     = "/show/prefix"
	RestPathShowStart      = "/show/start"
	RestPathShowStatus     = "/show/status"
	RestPathShowSubscriber = "/show/subscriber"
	RestPathShowVCS        = "/show/vcs"

	RestPathStatic = "/static"

	RestPathSubscribe   = "/subscribe"
	RestPathUnsubscribe = "/unsubscribe"

	RestPathWhoisAddressed = "/whois/addressed"
	RestPathWhoisId        = "/whois/id"
	RestPathWhoisNamed     = "/whois/named"
)

const RestOpCheckinExchangePort = "port"

const RestWhoisDepth = 8

var (
	ErrNoVCS  = errors.New("no VCS revision")
	ErrBadVCS = errors.New("mismatched VCS revision")
)

func defineRestFlags() {
	defineConfig()
	defineCert()
	defineRegistry()
	defineRegistryPort()
	defineSig()
	defineVPN()
}

var rest struct {
	bufs sync.Pool
	crt,
	reg *x509.Certificate
	url *url.URL
	http.Client
	vcsrev string
	ips    []net.IP

	fault chan error

	whoisReqC chan any // name, [Id], or [netip.Addr]
	whoisRspC chan *Subscriber
}

func restInit() error {
	var err error

	rest.bufs.New = func() any { return new(bytes.Buffer) }
	rest.vcsrev = xprogram.VcsRevision.String()
	rest.fault = make(chan error, 1)
	rest.whoisReqC = make(chan any, RestWhoisDepth)
	rest.whoisRspC = make(chan *Subscriber, RestWhoisDepth)

	rest.crt, err = readCertificateFile(vpnCertFile)
	if err != nil {
		return err
	}

	if err = signInit(); err != nil {
		return err
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{rest.crt.Raw},
				PrivateKey:  signPriv,
			},
		},
	}

	if rcas, err := x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	} else {
		cfg.RootCAs = rcas
	}

	if cl := flag.CommandLine.Name(); !strings.HasSuffix(cl, "certify") {
		rest.reg, err = readCertificateFile(vpnRegistryFile)
		if err != nil {
			return err
		}
		if err = restExtractURL(); err != nil {
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

func RestAdmin(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("subscriber")
	}

	if err = restInit(); err != nil {
		return err
	}
	var path string
	switch op := xflag.LastName(flag.CommandLine); op {
	case "approve":
		path = restPath(RestPathApprove, args[0])
	case "deny":
		path = restPath(RestPathDeny, args[0])
	default:
		return xerrors.Invalid(op)
	}
	_, err = restPut(ctx, os.Stdout, "", nil, path)
	return err
}

// Rest.certify adds the peer certificate to [Reg].
func RestCertify(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	defineRestFlags()
	yes := flag.CommandLine.Bool("y", false,
		fmt.Sprint("Yes, write remote certificate to ",
			vpnRegistryFile))
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("registry")
	}

	if err = restInit(); err != nil {
		return err
	}

	rest.url, err = url.Parse(args[0])
	if err != nil {
		return err
	}

	tp := rest.Client.Transport.(*http.Transport)
	sv := tp.TLSClientConfig.InsecureSkipVerify
	defer func() {
		tp.TLSClientConfig.InsecureSkipVerify = sv
	}()
	tp.TLSClientConfig.InsecureSkipVerify = true

	rsp, err := restGet(ctx, os.Stdout, RestPathCertify)
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

	if !*yes {
		fmt.Fprintf(w, `Enter "yes" to write above to %s: `,
			vpnRegistryFile)
		s, err := r.ReadString('\n')
		if err != nil && strings.TrimSpace(s) != "yes" {
			return err
		}
	}

	blk := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rsp.TLS.PeerCertificates[0].Raw,
	}
	wc, err := os.OpenFile(vpnRegistryFile, oCreate, 0644)
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

	defineRestFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	path := new(strings.Builder)
	path.WriteString(RestPathStatic)
	if args = flag.CommandLine.Args(); len(args) > 0 {
		if !strings.HasPrefix(args[0], "/") {
			path.WriteRune('/')
		}
		path.WriteString(args[0])
	}
	_, err = restGet(ctx, os.Stdout, path.String())
	if errors.Is(err, ErrBadVCS) {
		err = nil
	}
	return err
}

func RestReload(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful reload registry configuration.

{{flags .}}`)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	_, err = restPut(ctx, os.Stdout, "", nil, RestPathReload)
	return err
}

const restShowUsageTemplate = `
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`

func RestShow(ctx context.Context, args []string) error {
	xflag.TemplateUsage(restShowUsageTemplate)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	var path strings.Builder
	path.WriteString("/show/")
	path.WriteString(xflag.LastName(flag.CommandLine))
	for _, s := range args {
		path.WriteRune('/')
		path.WriteString(s)
	}
	_, err = restGet(ctx, os.Stdout, path.String())
	return err
}

func RestSubscribe(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	clone := *rest.url
	clone.Path = RestPathSubscribe
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	req.Header.Set("Content-Type", "application/x-pem-file")
	rsp, err := rest.Client.Do(req)
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
	} else if rsp.StatusCode == http.StatusOK {
		_, err = io.Copy(os.Stdout, rsp.Body)
	} else {
		sb := new(strings.Builder)
		io.Copy(sb, rsp.Body)
		err = fmt.Errorf("%s, %s", rsp.Status, sb)
	}
	return err
}

func restAlloc() *bytes.Buffer {
	return rest.bufs.Get().(*bytes.Buffer)
}

func restFree(buf *bytes.Buffer) {
	buf.Reset()
	rest.bufs.Put(buf)
}

func restExchangeCheckin(ctx context.Context) error {
	var id uint

	buf := restAlloc()
	defer restFree(buf)

	path := restPath(RestPathCheckinExchange, fmt.Sprint(vpnExchangePort))
	rsp, err := restPut(ctx, buf, "", nil, path)
	if err != nil {
		return err
	}
	if err = restValidateCheckinResponse(rsp); err != nil {
		return err
	}
	if _, err = fmt.Fscan(buf, &id); err != nil {
		return err
	}
	MyId = Id(id)
	MyLabel = MakeLabel(MyId, MyId)
	return nil
}

// eXtract registry url from its certificate.
func restExtractURL() error {
	var err error

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
	s := fmt.Sprint("https://", rest.reg.DNSNames[0], ":", vpnRegistryPort)
	rest.url, err = url.Parse(s)
	return err
}

func restGet(
	ctx context.Context,
	// rsp body
	w io.Writer,
	path string,
	// optional query {key, value} pairs
	kv ...string,
) (*http.Response, error) {
	return restRequest(ctx, http.MethodGet, w, "", nil, path, kv...)
}

func restGuestCheckin(ctx context.Context, pubpem []byte) (
	*GuestReceipt, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	receipt := new(GuestReceipt)
	rsp, err := restPut(ctx, buf,
		"application/x-pem-file", bytes.NewReader(pubpem),
		RestPathCheckinGuest)
	if err != nil {
		return receipt, err
	}
	if err = restValidateCheckinResponse(rsp); err != nil {
		return receipt, err
	}
	if err = json.Unmarshal(buf.Bytes(), &receipt); err != nil {
		return receipt, err
	}
	MyId = receipt.Id
	MyLabel = MakeLabel(MyId, MyId)
	return receipt, nil
}

// Join stringed args with "/".
func restPath(args ...string) string {
	var path strings.Builder
	for _, s := range args {
		if !strings.HasPrefix(s, "/") {
			path.WriteRune('/')
		}
		path.WriteString(s)
	}
	return path.String()
}

func restPut(
	ctx context.Context,
	// rsp body
	w io.Writer,
	// req content Type and body
	ct string, r io.Reader,
	path string,
	// optional query {key, value} pairs
	kv ...string,
) (*http.Response, error) {
	return restRequest(ctx, http.MethodPut, w, ct, r, path, kv...)
}
func restQueueWhois(ctx context.Context, v any) bool {
	return xcontext.Queue(ctx, rest.whoisReqC, v)
}

func restRequest(
	ctx context.Context,
	method string,
	// rsp body
	w io.Writer,
	// req content Type and body
	ct string, r io.Reader,
	path string,
	// optional query {key, value} pairs
	kv ...string,
) (*http.Response, error) {
	if len(rest.ips) == 0 {
		if err := restWaitForResolution(ctx); err != nil {
			return nil, err
		}
	}

	clone := *rest.url
	clone.Path = path
	q := clone.Query()
	for i, n := 0, len(kv); i < n; i += 2 {
		var v string
		if i < n-1 {
			v = kv[i+1]
		}
		q.Set(kv[i], v)
	}
	clone.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, clone.String(), r)
	if err != nil {
		return nil, err
	}
	if len(ct) > 0 {
		req.Header.Set("Content-Type", ct)
	}
	rsp, err := rest.Do(req)
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
		return rsp, err
	}
	if s := rsp.Header.Get(RestVcsRevision); len(s) == 0 {
		return rsp, ErrNoVCS
	} else if len(rest.vcsrev) == 0 {
		rest.vcsrev = s
	} else if rest.vcsrev != s {
		return rsp, ErrBadVCS
	}
	if rsp.StatusCode == http.StatusOK {
		_, err = io.Copy(w, rsp.Body)
	} else {
		sb := new(strings.Builder)
		io.Copy(sb, rsp.Body)
		err = fmt.Errorf("%s, %s", rsp.Status, sb)
	}
	return rsp, err
}

func restValidateCheckinResponse(rsp *http.Response) error {
	if rsp == nil {
		return xerrors.Invalid("checkin response")
	}

	vcsRev := xprogram.VcsRevision.String()
	regVcsRev := rsp.Header.Get(RestVcsRevision)
	if vcsRev != regVcsRev {
		return fmt.Errorf("upgrade to %s", regVcsRev)
	}

	s := rsp.Header.Get(RestUnixMicroStart)
	if len(s) == 0 {
		return xerrors.Unavailable("registry start time")
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return xerrors.Label(err, "registry start")
	}
	vpnStart = i
	return nil
}

func restWaitForResolution(ctx context.Context) error {
	const timeout = time.Minute
	var err error

	hn := rest.url.Hostname()
	rest.ips, err = WaitForResolution(ctx, "ip", hn, timeout)
	return err
}

func restWhois(ctx context.Context, v any) (*Subscriber, error) {
	var path string
	buf := restAlloc()
	defer restFree(buf)
	switch t := v.(type) {
	case Id:
		path = restPath(RestPathWhoisId, fmt.Sprint(t.Index()))
	case int:
		path = restPath(RestPathWhoisId, fmt.Sprint(t))
	case string:
		path = restPath(RestPathWhoisNamed, t)
	case netip.Addr:
		path = restPath(RestPathWhoisAddressed, t.String())
	default:
		xlog.Errata.Printf("%T: unsupported", t)
		return nil, xerrors.Unsupported(fmt.Sprintf("%T", t))
	}
	sub := new(Subscriber)
	_, err := restGet(ctx, buf, path)
	if err == nil {
		err = json.Unmarshal(buf.Bytes(), sub)
		if err == nil {
			err = sub.validate()
		}
	}
	return sub, err
}

func restWhoisService(ctx context.Context) {
	cn := rest.crt.Subject.CommonName

	xlog.Trace.Println("start", cn, "whois request service")
	defer xlog.Trace.Println("stopped", cn, "whois request service")
	defer close(rest.whoisRspC)

	for {
		select {
		case <-ctx.Done():
			return
		case q, ok := <-rest.whoisReqC:
			if !ok {
				return
			}
			sub, err := restWhois(ctx, q)
			if err != nil {
				xlog.Errata.Print(err)
			} else {
				xcontext.Queue(ctx, rest.whoisRspC, sub)
			}
		}
	}
}
