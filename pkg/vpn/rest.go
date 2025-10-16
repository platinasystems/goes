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
	"github.com/platinasystems/goes/v2/pkg/xcontext"
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

	RestVcsCheckInterval = 30 * time.Second
)

const (
	DnsQuery = "/dns-query"

	Rest                = "/rest"
	RestApprove         = "/rest/approve"
	RestCertify         = "/rest/certify"
	RestCheckin         = "/rest/checkin"
	RestCheckinExchange = "/rest/checkin/exchange"
	RestCheckinGuest    = "/rest/checkin/guest"
	RestDeny            = "/rest/deny"
	RestDumpSubscribers = "/rest/dump/subscribers"
	RestInvite          = "/rest/invite"
	RestReload          = "/rest/reload"
	RestShow            = "/rest/show"
	RestShowAddress     = "/rest/show/address"
	RestShowAdmins      = "/rest/show/admins"
	RestShowDomain      = "/rest/show/domain"
	RestShowExchanges   = "/rest/show/exchanges"
	RestShowPending     = "/rest/show/pending"
	RestShowPrefix      = "/rest/show/prefix"
	RestShowStart       = "/rest/show/start"
	RestShowStatus      = "/rest/show/status"
	RestShowSubscriber  = "/rest/show/subscriber"
	RestShowVCS         = "/rest/show/vcs"
	RestSubscribe       = "/rest/subscribe"
	RestUnsubscribe     = "/rest/unsubscribe"
	RestWhois           = "/rest/whois"
	RestWhoisAddressed  = "/rest/whois/addressed"
	RestWhoisId         = "/rest/whois/id"
	RestWhoisNamed      = "/rest/whois/named"
)

const RestOpCheckinExchangePort = "port"

const RestWhoisDepth = 8

var ErrKoApp = errors.New("ko app")
var ErrNilResponse = errors.New("rest: nil respone")

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

var (
	registryFile string
	restPort     = 8003
)

func RegistryPath() string { return xmain.ConfigFile(registryFile) }

var restFlags = xflag.Labels{
	xmain.ConfigFlag,
	cert.Flag,
	sig.Flag,
	xflag.Label{"registry",
		"Registry certificate file w/in current or config directory.",
		func() any {
			var ok bool
			registryFile, ok = xmain.LookupEnv("REGISTRY")
			if !ok {
				registryFile = "registry.pem"
			}
			return &registryFile
		},
	},
	xflag.Label{"port", "REST listener.", &restPort},
}

var RestRestartRequiredErr error

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

	rest.crt, err = cert.ReadFile(cert.Path())
	if err != nil {
		return err
	}

	if err = sig.Init(); err != nil {
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
		rest.reg, err = cert.ReadFile(RegistryPath())
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			rest.reg = rest.crt
		}
		if err = restExtractURL(); err != nil {
			return err
		}
		cfg.RootCAs.AddCert(rest.reg)
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

func Admin(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <subscriber>
RESTful registry administration.

{{flags .}}`)

	err := restFlags.Define()
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
	if err == nil {
		return nil
	}
	if rsp != nil && rsp.StatusCode == http.StatusUpgradeRequired {
		err = restUpgrade(ctx)
	}
	return err
}

func CheckinExchange(ctx context.Context) (uint16, error) {
	var id uint
	var port uint16

	buf := restAlloc()
	defer restFree(buf)

	rsp, err := restPut(ctx, buf, "", nil, RestCheckinExchange)
	if err != nil {
		return 0, err
	}
	if err = RestValidateCheckinResponse(rsp); err != nil {
		return 0, err
	}
	if _, err = fmt.Fscan(buf, &id, &port); err != nil {
		return 0, err
	}
	MyId = Id(id)
	MyLabel = MakeLabel(MyId, MyId)
	return port, nil
}

func CheckinGuest(ctx context.Context, encap []byte) (
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
	if err = RestValidateCheckinResponse(rsp); err != nil {
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
func Certify(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	var yes bool

	err := append(restFlags, xflag.Label{"y",
		"Yes, to write remote certificate to registry file.",
		&yes,
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

	if !yes {
		fmt.Fprintf(w, `Enter "yes" to write above to %s: `,
			RegistryPath())
		s, err := r.ReadString('\n')
		if err != nil && strings.TrimSpace(s) != "yes" {
			return err
		}
	}

	blk := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rsp.TLS.PeerCertificates[0].Raw,
	}
	wc, err := os.Create(RegistryPath())
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func Get(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [filename]
Get or list registry file(s).

{{flags .}}`)

	err := restFlags.Define()
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

func Invite(ctx context.Context, name string, cipherText []byte) (
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

func Lookup(ctx context.Context, args []string) error {
	const class = xdnsmessage.ClassINET

	xflag.TemplateUsage(`
usage: {{.Name}} <address|name>
Print name of addressed, or address of named subscriber.

{{flags .}}`)

	err := restFlags.Define()
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

func QueueVcsCheck() {
	rest.whoisReqC <- nil
}

func Reload(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful reload registry configuration.

{{flags .}}`)

	err := restFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else {
		_, err = restPut(ctx, os.Stdout, "", nil, RestReload)
	}
	return err
}

func Show(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`)

	err := restFlags.Define()
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

func Subscribe(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful subscribe to VPN.

{{flags .}}`)

	err := restFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	clone := *rest.url
	clone.Path = RestSubscribe
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	_, err = restDo(os.Stdout, req)
	return err
}

func Update(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Download and install program update from registry.

{{flags .}}`)

	err := restFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else if err = AssertVcsMatch(ctx); err != nil {
	} else {
		fmt.Println(xprogram.Path(), "is up to date.")
	}
	return err
}

func Whois(ctx context.Context, v any) (*Subscriber, error) {
	var p string
	var sub *Subscriber
	var start int64
	buf := restAlloc()
	defer restFree(buf)
	switch t := v.(type) {
	case nil:
		p = RestShowStatus
	case Id:
		p = path.Join(RestWhoisId, fmt.Sprint(t.Index()))
	case int:
		p = path.Join(RestWhoisId, fmt.Sprint(t))
	case string:
		p = path.Join(RestWhoisNamed, fmt.Sprint(t))
	case netip.Addr:
		p = path.Join(RestWhoisAddressed, t.String())
	default:
		err := xerrors.Unsupported(fmt.Sprintf("%T", t))
		return nil, err
	}
	rsp, err := restGet(ctx, buf, p)
	if rsp != nil && rsp.StatusCode == http.StatusUpgradeRequired {
		err = restUpgrade(ctx)
	} else if start, err = getRegistryStart(rsp); err != nil {
		err = xerrors.NewExitError(RestartExitCode,
			NewRestartError(err))
	} else if RegistryStart != 0 && RegistryStart != start {
		err = ExitRecheckin
	} else if err == nil && v != nil {
		sub = new(Subscriber)
		err = json.Unmarshal(buf.Bytes(), sub)
		if err == nil {
			err = sub.validate()
		}
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
	s := fmt.Sprint("https://", rest.reg.DNSNames[0], ":", restPort)
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
	return restDo(w, req)
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

func RestValidateCheckinResponse(rsp *http.Response) (err error) {
	if rsp == nil {
		err = xerrors.Invalid("no checkin response")
	} else {
		vcsRev := xprogram.VcsRevision.String()
		regVcsRev := rsp.Header.Get(RestVcsRevision)
		if vcsRev != regVcsRev {
			err = fmt.Errorf("upgrade to %s", regVcsRev)
		} else {
			RegistryStart, err = getRegistryStart(rsp)
		}
	}
	return
}

func restWaitForResolution(ctx context.Context) error {
	const timeout = time.Minute
	var err error

	hn := rest.url.Hostname()
	rest.ips, err = WaitForResolution(ctx, "ip", hn, timeout)
	return err
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
			sub, err := Whois(ctx, q)
			if err == nil {
				xcontext.Queue(ctx, rest.whoisRspC, sub)
			} else if needsRestart(err) {
				RestRestartRequiredErr = err
				return
			} else {
				xlog.Errata.Print(err)
			}
		}
	}
}
