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
	RestError = "X-Error"

	RestUnixMicroStart = "X-Unix-Micro-Start"

	RestVcsRevision = "X-Vcs-Revision"
)

const RestSubscriber = "subscriber"

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
	RestOpCheckinExchange = "exchange"
	RestOpCheckinGuest    = "guest"
)

const RestOpCheckinExchangePort = "port"

const RestOpDumpSubscribers = "subscribers"

const (
	RestOpShowActive     = "active"
	RestOpShowAddress    = "address"
	RestOpShowAdmins     = "admins"
	RestOpShowExchanges  = "exchanges"
	RestOpShowHosts      = "hosts"
	RestOpShowPending    = "pending"
	RestOpShowPrefix     = "prefix"
	RestOpShowStart      = "start"
	RestOpShowSubscriber = "subscriber"
	RestOpShowTenant     = "tenant"
	RestOpShowVcs        = "vcs"
)

const (
	RestOpShowVcsModified = "modified"
	RestOpShowVcsRevision = "revision"
)

const (
	RestWhoisAddress = "address"
	RestWhoisId      = "id"
	RestWhoisName    = "name"
	RestWhoisDepth   = 8
)

var (
	ErrNoVCS  = errors.New("no VCS revision")
	ErrBadVCS = errors.New("mismatched VCS revision")
)

func defineRestFlags() {
	defineConfig()
	defineCert()
	defineRegistry()
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
	_, err = restRequest(ctx, os.Stdout, http.MethodPut, "", nil,
		RestOp, xflag.LastName(flag.CommandLine),
		RestSubscriber, args[0])
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

	rsp, err := restRequest(ctx, os.Stdout, http.MethodPut, "", nil,
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

	args = flag.CommandLine.Args()
	_, err = restRequest(ctx, os.Stdout, http.MethodGet, "", nil, args...)
	if errors.Is(err, ErrBadVCS) {
		err = nil
	}
	return err
}

func RestPing(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
RESTful ping registry.

{{flags .}}`)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	_, err = restRequest(ctx, os.Stdout, http.MethodGet, "", nil,
		RestOp, RestOpPing)
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

	_, err = restRequest(ctx, os.Stdout, http.MethodPut, "", nil,
		RestOp, RestOpReload)
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

	show := xflag.LastName(flag.CommandLine)
	op := []string{
		RestOp, RestOpShow,
		RestOpShow, show,
	}
	for i, arg := range flag.Args() {
		op = append(op, fmt.Sprint("arg", i), arg)
	}
	rsp, err := restRequest(ctx, os.Stdout, http.MethodGet, "", nil, op...)
	if err == nil || errors.Is(err, ErrBadVCS) {
		if show == RestOpShowStart {
			var i int64
			s := rsp.Header.Get(RestUnixMicroStart)
			i, err = strconv.ParseInt(s, 10, 64)
			if err == nil {
				fmt.Println(time.UnixMicro(i))
			}
		}
	}
	return err
}

func RestShowVcs(ctx context.Context, args []string) error {
	xflag.TemplateUsage(restShowUsageTemplate)

	defineRestFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = restInit(); err != nil {
		return err
	}

	show := xflag.LastName(flag.CommandLine)
	op := []string{
		RestOp, RestOpShow,
		RestOpShow, RestOpShowVcs,
		RestOpShowVcs, show,
	}
	rsp, err := restRequest(ctx, os.Stdout, http.MethodGet, "", nil, op...)
	if err == nil || errors.Is(err, ErrBadVCS) {
		if show == RestOpShowVcsRevision {
			fmt.Println(rsp.Header.Get(RestVcsRevision))
		}
	}
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
	if rsp != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
	} else if rsp.StatusCode == http.StatusOK {
		_, err = io.Copy(os.Stdout, rsp.Body)
	} else if s := rsp.Header.Get(RestError); len(s) > 0 {
		err = fmt.Errorf("%s, %s", rsp.Status, s)
	} else {
		err = errors.New(rsp.Status)
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

	rsp, err := restRequest(ctx, buf, http.MethodPut,
		"", nil,
		RestOp, RestOpCheckin,
		RestOpCheckin, RestOpCheckinExchange,
		RestOpCheckinExchangePort, fmt.Sprint(vpnPort))
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
	s := fmt.Sprint("https://", rest.reg.DNSNames[0], ":", defaultPort)
	rest.url, err = url.Parse(s)
	return err
}

func restGuestCheckin(ctx context.Context, pubpem []byte) (
	*GuestReceipt, error,
) {
	buf := restAlloc()
	defer restFree(buf)

	receipt := new(GuestReceipt)
	rsp, err := restRequest(ctx, buf, http.MethodPut,
		"application/x-pem-file", bytes.NewReader(pubpem),
		RestOp, RestOpCheckin,
		RestOpCheckin, RestOpCheckinGuest)
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

func restQueueWhois(ctx context.Context, v any) bool {
	return xcontext.Queue(ctx, rest.whoisReqC, v)
}

// - Without args, copy registry virtual directory listing.
// - With one arg, copy registry file.
// - With one or more (key, value) pairs, RESTful registry query response.
func restRequest(
	ctx context.Context,
	// rsp body
	w io.Writer,
	method string,
	// req body
	contentType string, r io.Reader,
	args ...string,
) (*http.Response, error) {
	if len(rest.ips) == 0 {
		if err := restWaitForResolution(ctx); err != nil {
			return nil, err
		}
	}

	clone := *rest.url
	switch len(args) {
	case 0:
		clone.Path = ""
	case 1:
		clone.Path = args[0]
	default:
		q := clone.Query()
		for i, n := 0, len(args); i < n; i += 2 {
			var v string
			if i < n-1 {
				v = args[i+1]
			}
			q.Set(args[i], v)
		}
		clone.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, clone.String(), r)
	if err != nil {
		return nil, err
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
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
	} else if s := rsp.Header.Get(RestError); len(s) > 0 {
		err = fmt.Errorf("%s, %s", rsp.Status, s)
	} else {
		err = errors.New(rsp.Status)
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

func restWhois(ctx context.Context, q any) (*Subscriber, error) {
	var k, s string
	buf := restAlloc()
	defer restFree(buf)
	switch t := q.(type) {
	case Id:
		k = RestWhoisId
		s = fmt.Sprint(t.Index())
	case int:
		k = RestWhoisId
		s = fmt.Sprint(t)
	case string:
		k = RestWhoisName
		s = t
	case netip.Addr:
		k = RestWhoisAddress
		s = t.String()
	default:
		xlog.Errata.Printf("%T: unsupported", t)
		return nil, xerrors.Unsupported(fmt.Sprintf("%T", t))
	}
	sub := new(Subscriber)
	_, err := xerrors.MarkResult(restRequest(ctx, buf, http.MethodGet,
		"", nil,
		RestOp, RestOpWhois,
		RestOpWhois, k,
		k, s))
	if err == nil {
		err = xerrors.Mark(json.Unmarshal(buf.Bytes(), sub))
		if err == nil {
			err = xerrors.Mark(sub.validate())
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
