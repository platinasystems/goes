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
	"runtime"
	"strconv"
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
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

type RestError struct {
	code int
	txt,
	body string
}

const contextApplicationPKCS8 = "application/pkcs8"
const dnsLookupTimeout = 30 * time.Second

const (
	RestUnixMicroStart = "X-Unix-Micro-Start"

	RestVcsRevision = "X-Vcs-Revision"

	RestVcsCheckInterval = 30 * time.Second
)

const (
	RestPathApprove = "/approve"

	RestPathCertify = "/certify"

	RestPathCheckinExchange = "/checkin/exchange"
	RestPathCheckinGuest    = "/checkin/guest"

	RestPathDeny = "/deny"

	RestPathDnsQuery = "/dns-query"

	RestPathDumpSubscribers = "/dmup/subscribers"

	RestPathInvite = "/invite"

	RestPathPing = "/ping"

	RestPathReload = "/reload"

	RestPathShowAddress    = "/show/address"
	RestPathShowAdmins     = "/show/admins"
	RestPathShowExchanges  = "/show/exchanges"
	RestPathShowPending    = "/show/pending"
	RestPathShowPrefix     = "/show/prefix"
	RestPathShowStart      = "/show/start"
	RestPathShowStatus     = "/show/status"
	RestPathShowSubscriber = "/show/subscriber"
	RestPathShowVCS        = "/show/vcs"

	RestPathSubscribe   = "/subscribe"
	RestPathUnsubscribe = "/unsubscribe"

	RestPathWhoisAddressed = "/whois/addressed"
	RestPathWhoisId        = "/whois/id"
	RestPathWhoisNamed     = "/whois/named"
)

const RestOpCheckinExchangePort = "port"

const RestWhoisDepth = 8

var ErrKoApp = errors.New("ko app")
var ErrRestartRequired = errors.New("restart required to complete upgrade")
var ErrNilResponse = errors.New("rest: nil respone")

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

var RestPaths = []string{
	RestPathApprove,
	RestPathCertify,
	RestPathCheckinExchange,
	RestPathCheckinGuest,
	RestPathDeny,
	RestPathDnsQuery,
	RestPathDumpSubscribers,
	RestPathInvite,
	RestPathPing,
	RestPathReload,
	RestPathShowAddress,
	RestPathShowAdmins,
	RestPathShowExchanges,
	RestPathShowPending,
	RestPathShowPrefix,
	RestPathShowStart,
	RestPathShowStatus,
	RestPathShowSubscriber,
	RestPathShowVCS,
	RestPathSubscribe,
	RestPathUnsubscribe,
	RestPathWhoisAddressed,
	RestPathWhoisId,
	RestPathWhoisNamed,
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

func RestAdmin(ctx context.Context, args []string) error {
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
	var path string
	switch op := xflag.LastName(flag.CommandLine); op {
	case "approve":
		path = restPath(RestPathApprove, args[0])
	case "deny":
		path = restPath(RestPathDeny, args[0])
	case "unsubscribe":
		path = restPath(RestPathUnsubscribe, args[0])
	default:
		return xerrors.ErrInvalid
	}
	_, err = restPut(ctx, os.Stdout, "", nil, path)
	return err
}

// RestCertify writes the peer certificate to [RegistryFile].
func RestCertify(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] https://<host>[:port]
Import registry certificate.

{{flags .}}`)

	var yes bool

	err := append(restFlags, xflag.Label{
		"y", "Yes, to write remote certificate to registry file.", &yes,
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

func RestGet(ctx context.Context, args []string) error {
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

func RestReload(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
RESTful reload registry configuration.

{{flags .}}`)

	err := restFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else {
		_, err = restPut(ctx, os.Stdout, "", nil, RestPathReload)
	}
	return err
}

const restShowUsageTemplate = `
usage: {{.Name}} [flags] [args]
RESTful query and print registry object.

{{flags .}}`

func RestShow(ctx context.Context, args []string) error {
	xflag.TemplateUsage(restShowUsageTemplate)

	err := restFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	var path strings.Builder
	path.WriteString("/show/")
	path.WriteString(xflag.LastName(flag.CommandLine))
	for _, s := range flag.CommandLine.Args() {
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

	err := restFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if err = restInit(); err != nil {
		return err
	}

	clone := *rest.url
	clone.Path = RestPathSubscribe
	req, err := http.
		NewRequestWithContext(ctx, http.MethodPut, clone.String(), nil)
	if err != nil {
		return xerrors.Mark(err)
	}
	_, err = restDo(os.Stdout, req)
	return err
}

func RestUpdate(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Download and install program update from registry.

{{flags .}}`)

	err := restFlags.Define()
	if err != nil {
	} else if err = flag.CommandLine.Parse(args); err != nil {
	} else if err = restInit(); err != nil {
	} else if err = RestAssertVcsMatch(ctx); err != nil {
	} else {
		fmt.Println(xprogram.Path(), "is up to date.")
	}
	return err
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

func restAlloc() *bytes.Buffer {
	return rest.bufs.Get().(*bytes.Buffer)
}

// REST get registry status to validate version.
// If [http.Response.StatusCode] == [http.StatusUpgradeRequired],
// fetch and install upgrade then return [xerrors.ExitError]
// to force [os.Exit] with [xos.EX_TEMPFAIL].
func RestAssertVcsMatch(ctx context.Context) error {
	rsp, err := restGet(ctx, io.Discard, RestPathShowStatus)
	if err == nil {
		return nil
	}
	if rsp != nil && rsp.StatusCode == http.StatusUpgradeRequired {
		err = restUpgrade(ctx)
	}
	return err
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

func RestExchangeCheckin(ctx context.Context) (uint16, error) {
	var id uint
	var port uint16

	buf := restAlloc()
	defer restFree(buf)

	rsp, err := restPut(ctx, buf, "", nil, RestPathCheckinExchange)
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

func RestGuestCheckin(ctx context.Context, encap []byte) (
	*GuestReceipt, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	receipt := new(GuestReceipt)
	rsp, err := restPut(ctx, buf,
		"application/octet-stream", bytes.NewReader(encap),
		RestPathCheckinGuest)
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

func RestInvite(ctx context.Context, name string, cipherText []byte) (
	[]byte, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	_, err := restRequest(ctx, http.MethodPut, buf,
		"application/octet-stream", bytes.NewBuffer(cipherText),
		restPath(RestPathInvite, name))
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buf.Bytes()), nil
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
	return xerrors.NewExitError(UpgradeExitCode, ErrRestartRequired)
}

func RestValidateCheckinResponse(rsp *http.Response) error {
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
	RegistryStart = i
	return nil
}

func restWaitForResolution(ctx context.Context) error {
	const timeout = time.Minute
	var err error

	hn := rest.url.Hostname()
	rest.ips, err = WaitForResolution(ctx, "ip", hn, timeout)
	return err
}

func RestQueueVcsCheck() {
	rest.whoisReqC <- nil
}

func RestWhois(ctx context.Context, v any) (*Subscriber, error) {
	var path string
	var sub *Subscriber
	buf := restAlloc()
	defer restFree(buf)
	switch t := v.(type) {
	case nil:
		path = RestPathShowStatus
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
	rsp, err := restGet(ctx, buf, path)
	if rsp != nil && rsp.StatusCode == http.StatusUpgradeRequired {
		err = restUpgrade(ctx)
	} else if err == nil && v != nil {
		sub = new(Subscriber)
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
			sub, err := RestWhois(ctx, q)
			if err == nil {
				xcontext.Queue(ctx, rest.whoisRspC, sub)
			} else if errors.Is(err, ErrRestartRequired) {
				RestRestartRequiredErr = err
				return
			} else {
				xlog.Errata.Print(err)
			}
		}
	}
}
