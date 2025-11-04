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
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/sig"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

type RestError struct {
	code int
	txt,
	body string
}

type RestartError struct {
	err error
}

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

const (
	RestUnixMicroStart = "X-Unix-Micro-Start"

	RestVcsRevision = "X-Vcs-Revision"
)

const (
	DnsQuery = "/dns-query"

	RestApprove         = "/approve"
	RestCertify         = "/certify"
	RestCheckin         = "/checkin"
	RestCheckinExchange = RestCheckin + "/exchange"
	RestCheckinGuest    = RestCheckin + "/guest"
	RestDeny            = "/deny"
	RestDump            = "/dump"
	RestDumpSubscribers = RestDump + "/subscribers"
	RestInvite          = "/invite"
	RestReload          = "/reload"
	RestRevise          = "/revise"
	RestReviseIds       = RestRevise + "/ids"
	RestShow            = "/show"
	RestShowAddress     = RestShow + "/address"
	RestShowAdmins      = RestShow + "/admins"
	RestShowDomain      = RestShow + "/domain"
	RestShowExchanges   = RestShow + "/exchanges"
	RestShowPending     = RestShow + "/pending"
	RestShowPrefix      = RestShow + "/prefix"
	RestShowStart       = RestShow + "/start"
	RestShowStatus      = RestShow + "/status"
	RestShowSubscriber  = RestShow + "/subscriber"
	RestShowVCS         = RestShow + "/vcs"
	RestSubscribe       = "/subscribe"
	RestUnsubscribe     = "/unsubscribe"
	RestWhois           = "/whois"
	RestWhoisAddressed  = RestWhois + "/addressed"
	RestWhoisId         = RestWhois + "/id"
	RestWhoisNamed      = RestWhois + "/named"
)

var RestPrefixes = []string{
	DnsQuery,
	RestApprove,
	RestCertify,
	RestCheckin,
	RestDeny,
	RestDump,
	RestInvite,
	RestReload,
	RestRevise,
	RestShow,
	RestSubscribe,
	RestUnsubscribe,
	RestWhois,
}

type RestVcsCheck struct{}

const RestOpCheckinExchangePort = "port"

const (
	RestReqDepth = 8
	RestRspDepth = 8
)

var ErrKoApp = errors.New("ko app")
var ErrNilResponse = errors.New("rest: nil response")

var ErrCompleteUpgrade = errors.New("complete upgrade")
var ErrRestartCompleteUpgrade = NewRestartError(ErrCompleteUpgrade)
var ExitCompleteUpgrade = xerrors.
	NewExitError(RestartExitCode, ErrRestartCompleteUpgrade)

var ErrRecheckin = errors.New("re-checkin w/ registry")
var ErrRestartRecheckin = NewRestartError(ErrRecheckin)
var ExitRecheckin = xerrors.
	NewExitError(RestartExitCode, ErrRestartRecheckin)

func needsRestart(err error) bool {
	return errors.Is(err, ErrCompleteUpgrade) ||
		errors.Is(err, ErrRecheckin)
}

var rest struct {
	bufs sync.Pool
	crt,
	reg *x509.Certificate
	url *url.URL
	http.Client
	vcsrev string
	ips    []net.IP
	port   uint
	start  int64

	fault chan error

	// name, [Id], [netip.Addr], or []byte encoded [Id]'s
	reqC chan any
	// error, *Subscriber, or []byte encoded revised [Id]'s
	rspC chan any
}

var RestCertAkaFlag = xflag.Label{"cert", "aka. -ssl-client-cn", &cert.Client}

var RestPortFlag = xflag.Label{"port", "REST listener.", func() any {
	rest.port = 8003
	return &rest.port
}}

var RestFlags = xflag.Labels{
	xmain.ConfigFlag,
	cert.ClientFlag,
	cert.ServerFlag,
	sig.Flag,
	RestCertAkaFlag,
	RestPortFlag,
	xflag.Label{"registry", "aka. -ssl-server-dn", &cert.Server},
}

func restInit() error {
	if len(cert.Server) == 0 {
		return errors.New("no -registry, -ssl-server-dn" +
			", $" + xmain.EnvPrefix() + "SERVER_DN" +
			", or $SSL_SERVER_DN")
	}
	rest.bufs.New = func() any { return new(bytes.Buffer) }
	rest.vcsrev = xprogram.VcsRevision.String()
	rest.fault = make(chan error, 1)
	rest.reqC = make(chan any, RestReqDepth)
	rest.rspC = make(chan any, RestRspDepth)

	if cs, err := cert.ClientCerts(); err != nil {
		return err
	} else {
		rest.crt = cs[0]
	}

	if err := sig.Init(); err != nil {
		return err
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{rest.crt.Raw},
				PrivateKey:  sig.Priv,
			},
		},
	}

	if rcas, err := x509.SystemCertPool(); err != nil {
		cfg.RootCAs = x509.NewCertPool()
	} else {
		cfg.RootCAs = rcas
	}

	if cl := flag.CommandLine.Name(); !strings.HasSuffix(cl, "certify") {
		if cs, err := cert.ServerCerts(); err != nil {
			return err
		} else {
			rest.reg = cs[0]
			if err = restExtractURL(); err != nil {
				return err
			}
			cfg.RootCAs.AddCert(rest.reg)
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

func RestAdminReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.CommandLine.Args(); len(args) == 0 {
		return xerrors.Incomplete("subscriber")
	} else if err = restInit(); err != nil {
		return err
	}
	var p string
	switch op := xflag.LastName(flag.CommandLine); op {
	case "approve":
		p = path.Join(RestApprove, args[0])
	case "deny":
		p = path.Join(RestDeny, args[0])
	case "unsubscribe":
		p = path.Join(RestUnsubscribe, args[0])
	default:
		return xerrors.ErrInvalid
	}
	_, err = restPut(ctx, os.Stdout, "", nil, p)
	return err
}

// REST get registry status to validate version.
// If [http.Response.StatusCode] == [http.StatusUpgradeRequired],
// fetch and install upgrade then return [xerrors.ExitError]
// to force [os.Exit] with [xos.EX_TEMPFAIL].
func AssertVcsMatch(ctx context.Context) error {
	rsp, err := restGet(ctx, io.Discard, RestShowStatus)
	if rsp == nil {
		if err == nil {
			err = ErrNilResponse
		}
	} else if rsp.Header.Get(RestVcsRevision) != rest.vcsrev {
		err = restUpgrade(ctx)
	}
	return err
}

func RestCheckinExchangeReq(ctx context.Context) (uint16, error) {
	var id uint
	var port uint16

	buf := restAlloc()
	defer restFree(buf)

	rsp, err := restPut(ctx, buf, "", nil, RestCheckinExchange)
	if err != nil {
		return 0, err
	}
	if rest.start, err = registryStart(rsp); err != nil {
		return 0, err
	}
	if _, err = fmt.Fscan(buf, &id, &port); err != nil {
		return 0, err
	}
	MyId = Id(id)
	MyLabel = MakeLabel(MyId, MyId)
	return port, nil
}

func RestCheckinGuestReq(ctx context.Context, encap []byte) (
	*GuestReceipt, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	receipt := new(GuestReceipt)
	rsp, err := restPut(ctx, buf,
		"application/octet-stream", bytes.NewReader(encap),
		RestCheckinGuest)
	if err != nil {
		return receipt, err
	}
	if rest.start, err = registryStart(rsp); err != nil {
		return receipt, err
	}
	if err = json.Unmarshal(buf.Bytes(), &receipt); err != nil {
		return receipt, err
	}
	MyId = receipt.Id
	MyLabel = MakeLabel(MyId, MyId)
	return receipt, nil
}

// Certify writes the peer certificate to [RegistryFile].
func RestCertifyReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	var yes bool
	err := append(RestFlags, xflag.Label{
		"y", "Yes, to write remote certificate.", &yes,
	}).Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.CommandLine.Args(); len(args) == 0 {
		return xerrors.Incomplete("registry")
	} else if rest.url, err = url.Parse(args[0]); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	tp := rest.Client.Transport.(*http.Transport)
	sv := tp.TLSClientConfig.InsecureSkipVerify
	defer func() {
		tp.TLSClientConfig.InsecureSkipVerify = sv
	}()
	tp.TLSClientConfig.InsecureSkipVerify = true

	rsp, err := restGet(ctx, os.Stdout, RestCertify)
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

	if t, err := cert.NewTemplate(); err != nil {
		return err
	} else if err = t.Execute(w, rsp.TLS.PeerCertificates); err != nil {
		return err
	}

	fn := cert.Server
	if !strings.HasSuffix(fn, cert.Ext) {
		fn = fmt.Sprint(fn, cert.Ext)
	}
	if strings.IndexRune(fn, os.PathSeparator) < 0 {
		fn = xmain.ConfigFile(fn)
	}
	if !yes {
		fmt.Fprint(w, `Enter "yes" to write above to `, fn, ": ")
		s, err := r.ReadString('\n')
		if err != nil || strings.TrimSpace(s) != "yes" {
			return err
		}
	}

	blk := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rsp.TLS.PeerCertificates[0].Raw,
	}
	wc, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func RestGetReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [filename]
Get or list registry file(s).

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	path := new(strings.Builder)
	if args = flag.CommandLine.Args(); len(args) > 0 {
		if !strings.HasPrefix(args[0], "/") {
			path.WriteRune('/')
		}
		path.WriteString(args[0])
	}
	_, err = restGet(ctx, os.Stdout, path.String())
	return err
}

func RestInviteReq(ctx context.Context, name string, cipherText []byte) (
	[]byte, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	_, err := restRequest(ctx, http.MethodPut, buf,
		"application/octet-stream", bytes.NewBuffer(cipherText),
		path.Join(RestInvite, name))
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buf.Bytes()), nil
}

func RestLookupReq(ctx context.Context, args []string) error {
	const class = xdnsmessage.ClassINET

	xflag.TemplateUsage(`
usage: {{.Name}} <address|name>
Print name of addressed, or address of named subscriber.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.CommandLine.Args(); len(args) == 0 {
		return err
	} else if err = restInit(); err != nil {
		return err
	}
	clone := *rest.url
	clone.Path = DnsQuery
	doh := xdnsdoh.Asker(&rest.Client, clone.String())
	b := xdnsmessage.MakeBuffer()
	name := args[0]
	types := []xdnsmessage.Type{xdnsmessage.TypeA, xdnsmessage.TypeAAAA}
	if addr, err := netip.ParseAddr(args[0]); err == nil {
		name = xdnsmessage.Reverse(addr)
		types[0] = xdnsmessage.TypePTR
		types = types[:1]
	}
	us := xdnsmessage.MakeUniqueString(name)
	for _, t := range types {
		var rsp xdnsmessage.Message
		q := xdnsmessage.NewQuery(true, us, class, t)
		if b, err = q.AppendTo(b[:0]); err != nil {
			return err
		}
		if b, err = doh.Ask(ctx, b); err != nil {
			return err
		}
		if err = rsp.UnmarshalBinary(b); err != nil {
			return err
		}
		if rsp.ID != q.ID {
			return fmt.Errorf("id %d != %d", rsp.ID, q.ID)
		}
		for _, a := range rsp.Answers {
			fmt.Println(name, a)
		}
	}
	return nil
}

func RestReloadReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful reload registry configuration.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else {
		_, err = restPut(ctx, os.Stdout, "", nil, RestReload)
	}
	return err
}

func RestShowReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	s := path.Join(RestShow, xflag.LastName(flag.CommandLine))
	for _, arg := range flag.CommandLine.Args() {
		s = path.Join(s, arg)
	}
	_, err = restGet(ctx, os.Stdout, s)
	return err
}

func RestSubscribeReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else {
		_, err = restPut(ctx, os.Stdout, "", nil, RestSubscribe)
	}
	return err
}

func RestUpdateReq(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Download and install program update from registry.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else if err = AssertVcsMatch(ctx); err != nil {
	} else {
		fmt.Println(xprogram.Path(), "is up to date.")
	}
	return err
}

func RestWhoisReq(ctx context.Context, v any) (*Subscriber, error) {
	var p string
	buf := restAlloc()
	defer restFree(buf)
	switch t := v.(type) {
	case Id:
		p = path.Join(RestWhoisId, fmt.Sprint(t.Index()))
	case int:
		p = path.Join(RestWhoisId, fmt.Sprint(t))
	case string:
		p = path.Join(RestWhoisNamed, fmt.Sprint(t))
	case netip.Addr:
		p = path.Join(RestWhoisAddressed, t.String())
	default:
		return nil, fmt.Errorf("unsupported %T", v)
	}
	if _, err := restGet(ctx, buf, p); err != nil {
		return nil, err
	}
	sub := new(Subscriber)
	err := json.Unmarshal(buf.Bytes(), sub)
	if err == nil {
		err = sub.validate()
	}
	return sub, err
}

func (re *RestError) Code() int {
	return re.code
}

func (re *RestError) Error() string {
	if len(re.body) == 0 {
		return re.txt
	}
	return fmt.Sprint(re.txt, ", ", re.body)
}

func NewRestartError(err error) RestartError {
	return RestartError{err}
}

func (x RestartError) Error() (s string) {
	if x.err != nil {
		s = fmt.Sprint("restart required: ", x.err)
	}
	return
}

func (x RestartError) Unwrap() error {
	return x.err
}

func restAlloc() *bytes.Buffer {
	return rest.bufs.Get().(*bytes.Buffer)
}

func restDo(w io.Writer, req *http.Request) (*http.Response, error) {
	req.Header.Set(RestVcsRevision, rest.vcsrev)
	rsp, err := rest.Do(req)
	if err != nil {
		if rsp != nil {
			rsp.Body.Close()
		}
		return nil, fmt.Errorf("rest: %w", err)
	}
	if rsp == nil {
		return nil, ErrNilResponse
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusOK {
		io.Copy(w, rsp.Body)
		return rsp, nil
	}
	sb := new(strings.Builder)
	io.Copy(sb, rsp.Body)
	return rsp, &RestError{rsp.StatusCode, rsp.Status, sb.String()}
}

func restFree(buf *bytes.Buffer) {
	buf.Reset()
	rest.bufs.Put(buf)
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
	s := fmt.Sprint("https://", rest.reg.DNSNames[0], ":", rest.port)
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
	var start int64

	rsp, err := restRequest(ctx, http.MethodGet, w, "", nil, path, kv...)
	if rsp.StatusCode == http.StatusUpgradeRequired {
		err = restUpgrade(ctx)
	} else if err != nil {
		// skip to common return
	} else if start, err = registryStart(rsp); err != nil {
		err = NewRestartError(err)
		err = xerrors.NewExitError(RestartExitCode, err)
	} else if rest.start != 0 && rest.start != start {
		err = ExitRecheckin
	}
	return rsp, err
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

func restQueueReq(ctx context.Context, v any) (ok bool) {
	select {
	case <-ctx.Done():
	case rest.reqC <- v:
		ok = true
	default:
		xlog.Errata.Println("can't queue req:", v)
	}
	return
}

func restQueueReviseIds(ctx context.Context, b []byte) bool {
	var ok bool
	if n := len(b) / SizeofId; n > 0 {
		ok = restQueueReq(ctx, b)
	}
	return ok
}

func restQueueVcsCheck(ctx context.Context) bool {
	return restQueueReq(ctx, RestVcsCheck{})
}

func restQueueWhois(ctx context.Context, subref any) bool {
	return restQueueReq(ctx, subref)
}

func restQueueRsp(ctx context.Context, v any) (ok bool) {
	select {
	case <-ctx.Done():
	case rest.rspC <- v:
		ok = true
	default:
		xlog.Errata.Println("can't queue rsp:", v)
	}
	return
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
	return restDo(w, req)
}

func restReviseIds(ctx context.Context, b []byte) error {
	buf := restAlloc()
	defer restFree(buf)
	_, err := restPut(ctx, buf,
		"application/octet-stream; big-endian=true",
		bytes.NewReader(b),
		RestReviseIds)
	if err == nil {
		copy(b, buf.Bytes())
	}
	return err
}

// Fetch and install upgrade then return [xerrors.ExitError]
// to force [os.Exit] with [xos.EX_TEMPFAIL].
func restUpgrade(ctx context.Context) error {
	const ocreate = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	const cantUpgrade = "can't upgrade"
	if xprogram.IsKoApp() {
		return xerrors.Label(ErrKoApp, cantUpgrade)
	}
	xp := xprogram.Path()
	xpSave := fmt.Sprint(xp, "~")
	xpPlus := fmt.Sprint(xp, "+")
	mainPlatform := fmt.Sprint(xmain.PackageName(),
		"-", runtime.GOOS,
		"-", runtime.GOARCH)
	fi, err := os.Stat(xp)
	if err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	for _, s := range []string{xpSave, xpPlus} {
		err = os.Remove(s)
		if err != nil && errors.Is(err, fs.ErrPermission) {
			return xerrors.Label(err, cantUpgrade)
		}
	}
	f, err := os.OpenFile(xpPlus, ocreate, fi.Mode())
	if err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	_, err = restGet(ctx, f, mainPlatform)
	f.Close()
	if err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	if err = os.Link(xp, xpSave); err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	if err = os.Remove(xp); err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	if err = os.Link(xpPlus, xp); err != nil {
		return xerrors.Label(err, cantUpgrade)
	}
	return ExitCompleteUpgrade
}

func restVcsCheck(ctx context.Context) error {
	buf := restAlloc()
	defer restFree(buf)
	_, err := restGet(ctx, buf, RestShowStatus)
	return err
}

func restWaitForResolution(ctx context.Context) error {
	const timeout = time.Minute
	var err error

	hn := rest.url.Hostname()
	rest.ips, err = WaitForResolution(ctx, "ip", hn, timeout)
	return err
}

func restReqService(ctx context.Context) {
	cn := rest.crt.Subject.CommonName

	xlog.Trace.Println("start", cn, "whois request service")
	defer xlog.Trace.Println("stopped", cn, "whois request service")
	defer close(rest.rspC)

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-rest.reqC:
			if !ok {
				return
			}
			if b, ok := v.([]byte); ok {
				if err := restReviseIds(ctx, b); err != nil {
					restQueueRsp(ctx, err)
				} else {
					restQueueRsp(ctx, b)
				}
			} else if _, ok := v.(RestVcsCheck); ok {
				if err := restVcsCheck(ctx); err != nil {
					restQueueRsp(ctx, err)
				}
			} else if sub, err := RestWhoisReq(ctx, v); err != nil {
				restQueueRsp(ctx, err)
			} else if sub == nil {
				restQueueRsp(ctx, errors.New("nil sub"))
			} else {
				restQueueRsp(ctx, sub)
			}
		}
	}
}
