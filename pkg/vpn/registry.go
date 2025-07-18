// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
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
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/kvc"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmaps"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

type GuestReceipt struct {
	Id     Id
	Prefix netip.Prefix

	ExchangePrecedence []string
}

type registry struct {
	start, vcsRev string

	cert *x509.Certificate

	indexed []*Subscriber

	addressed map[netip.Addr]*Subscriber

	subscriberExchangePrecedence map[string][]string

	admin,
	named, // guest or exchange
	pending map[string]*Subscriber

	rsvpC chan *rsvp

	hosts struct {
		addr map[string]netip.Addr
		name map[netip.Addr]string
	}

	http *http.Server
	url  *url.URL

	topAddr netip.Addr
}

func newRegistry() *registry {
	reg := new(registry)

	vpnStart = time.Now().UnixMicro()
	reg.start = strconv.FormatInt(vpnStart, 10)

	reg.vcsRev = xprogram.VcsRevision.String()

	reg.addressed = make(map[netip.Addr]*Subscriber)
	reg.admin = make(map[string]*Subscriber)
	reg.named = make(map[string]*Subscriber)
	reg.pending = make(map[string]*Subscriber)

	reg.subscriberExchangePrecedence = make(map[string][]string)

	reg.rsvpC = make(chan *rsvp, 1)

	reg.hosts.addr = make(map[string]netip.Addr)
	reg.hosts.name = make(map[netip.Addr]string)

	return reg
}

type rsvp struct {
	http.ResponseWriter
	req *http.Request
	qv  url.Values
	ch  chan empty
}

var rsvpPool = &sync.Pool{
	New: func() any {
		return &rsvp{
			ch: make(chan empty),
		}
	},
}

func (rsvp *rsvp) isMethod(method string) bool {
	if rsvp.req.Method != method {
		rsvp.Header().Add(RestError, rsvp.req.Method)
		rsvp.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func (rsvp *rsvp) subscriber() string {
	if !rsvp.qv.Has(RestSubscriber) {
		rsvp.Header().Add(RestError, "unnamed subscriber")
		rsvp.WriteHeader(http.StatusBadRequest)
	}
	s := rsvp.qv.Get(RestSubscriber)
	if len(s) == 0 {
		rsvp.Header().Add(RestError, "invalid subscriber")
		rsvp.WriteHeader(http.StatusBadRequest)
	}
	return s
}

// virtual-link, this program may be fetched as MAIN-GOOS-GOARCH
var vlink = sync.OnceValue(func() string {
	return fmt.Sprintf("%s-%s-%s", xprogram.MainName(),
		runtime.GOOS, runtime.GOARCH)
})

// Registry is a web server providing a REST interface to persistent
// files and ephemeral tables.
func Registry(ctx context.Context, args []string) error {
	xlog.SetPrefixes("registry/")

	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
A RESTful WWW server and packet exchange.

{{flags .}}`)

	defineConfig()
	defineData()
	defineState()

	defineAdmins()
	defineCert()
	defineHosts()
	definePort()
	definePrefix()
	defineSig()
	defineVia()

	enableQuiet()
	enableTrace()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = signInit(); err != nil {
		return err
	}

	MyId = Id(0)
	MyLabel = MakeLabel(MyId, MyId)

	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)

	reg := newRegistry()

	if err = reg.loadAdminsFile(); err != nil {
		return err
	}
	if err = reg.loadHostsFile(); err != nil {
		return err
	}
	if err = reg.loadViaFile(); err != nil {
		return err
	}
	if err = reg.loadCerts(); err != nil {
		return err
	}

	if err = udpInit(vpnPort); err != nil {
		return err
	}
	defer udp.Close()

	reg.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", vpnPort),
		Handler: reg,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			ClientAuth: tls.RequireAnyClientCert,
		},
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	alarm := make(chan os.Signal, 2)
	signal.Notify(alarm, xsignal.Alarm)
	defer signal.Stop(alarm)

	wg.Go(func() { reg.shutdown(ctx) })
	wg.Go(reg.restsvc)
	wg.Go(udpStream)

	xlog.Trace.Println("start")
	defer cancel()
	defer xlog.Trace.Println("stopping...")

selection:
	for {
		select {
		case <-ctx.Done():
			break selection
		case <-alarm:
			xlog.Info = xlog.ToggleMute(xlog.Info)
			xlog.Trace = xlog.Mute(xlog.Trace)
		case rsvp, ok := <-reg.rsvpC:
			if !ok {
				break selection
			}
			reg.rest(rsvp)
		case m, ok := <-udpC:
			if !ok {
				break selection
			}
			if len(m.Data) < SizeofLabel {
				xlog.Errata.Print("underrun")
			} else {
				reg.fromUDP(ctx, m)
			}
			mp.Put(m)
		}
	}
	return nil
}

func (reg *registry) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if !req.TLS.HandshakeComplete {
		rsp.WriteHeader(http.StatusUnauthorized)
		rsp.Write([]byte("incomplete handshake"))
		return
	}
	if len(req.TLS.PeerCertificates) == 0 {
		rsp.WriteHeader(http.StatusUnauthorized)
		rsp.Write([]byte("no certificates"))
		return
	}
	qv := req.URL.Query()
	op := qv.Get(RestOp)

	if len(op) == 0 {
		reg.getFileOrDir(rsp, req)
		return
	}

	rsp.Header().Add(RestUnixMicroStart, reg.start)
	rsp.Header().Add(RestVcsRevision, reg.vcsRev)

	reg.rsvp(rsp, req, qv)
}

func (reg *registry) approve(rsvp *rsvp) {
	name := rsvp.subscriber()
	if len(name) == 0 {
		return
	}
	_, err := os.Stat(vpnStateDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(vpnStateDir, 0755)
		}
		if err != nil {
			rsvp.Header().Add(RestError, err.Error())
			rsvp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	sub, ok := reg.pending[name]
	if !ok {
		rsvp.Header().Add(RestError, name)
		rsvp.WriteHeader(http.StatusNotFound)
	}
	blk := pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: sub.CertDER,
	}
	if f, err := os.Create(sub.stateFileName()); err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
		return
	} else if err = pem.Encode(f, &blk); err != nil {
		f.Close()
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		f.Close()
	}
	delete(reg.pending, name)

	if err = reg.assignAddr(sub); err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
		return
	}
	reg.named[name] = sub
	sub.Id = Id(len(reg.indexed))
	reg.indexed = append(reg.indexed, sub)

	fmt.Fprintln(rsvp, "OK")
}

func (reg *registry) assignAddr(sub *Subscriber) error {
	var found bool
	cn := sub.cert.Subject.CommonName
	if sub.Addr, found = reg.hosts.addr[cn]; found {
		reg.addressed[sub.Addr] = sub
		return nil
	}
	if !reg.topAddr.IsValid() {
		reg.topAddr = vpnPrefix.Masked().Addr()
	} else {
		reg.topAddr = reg.topAddr.Next()
	}
	for {
		if _, found = reg.hosts.name[reg.topAddr]; !found {
			break
		}
		reg.topAddr = reg.topAddr.Next()
	}
	if !vpnPrefix.Contains(reg.topAddr) {
		return xerrors.Unavailable("address")
	}
	sub.Addr = reg.topAddr
	reg.addressed[sub.Addr] = sub
	return nil
}

func (reg *registry) checkin(rsvp *rsvp) *Subscriber {
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	cn := peer0.Subject.CommonName

	sub, ok := reg.named[cn]
	if !ok {
		rsvp.Header().Add(RestError, cn)
		rsvp.WriteHeader(http.StatusNotFound)
		return nil
	}
	sub.Id.Revise()

	return sub
}

func (reg *registry) checkinExchange(rsvp *rsvp) {
	sub := reg.checkin(rsvp)
	if sub == nil {
		return
	}
	if rsvp.qv.Has(RestOpCheckinExchangePort) {
		s := rsvp.qv.Get(RestOpCheckinExchangePort)
		u, err := strconv.ParseUint(s, 0, 16)
		if err != nil {
			rsvp.Header().Add(RestError, err.Error())
			rsvp.WriteHeader(http.StatusBadRequest)
			return
		}
		sub.Port = uint16(u)
	} else {
		rsvp.Header().Add(RestError, "no port")
		rsvp.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprint(rsvp, uint(sub.Id))
	xlog.Trace.Println("exchange", sub)
}

func (reg *registry) checkinGuest(rsvp *rsvp) {
	sub := reg.checkin(rsvp)
	if sub == nil {
		return
	}

	data, err := io.ReadAll(rsvp.req.Body)
	if err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusBadRequest)
		return
	}
	blk, _ := pem.Decode(data)
	if blk == nil {
		rsvp.Header().Add(RestError, "non-PEM input")
		rsvp.WriteHeader(http.StatusBadRequest)
		return
	}
	sub.CipherKeyDER = blk.Bytes

	rsvp.Header().Set("Content-Type", "application/json")
	bits := vpnPrefix.Bits()
	prefix := netip.PrefixFrom(sub.Addr, bits)
	receipt := GuestReceipt{
		Id:     sub.Id,
		Prefix: prefix,

		ExchangePrecedence: sub.ExchangePrecedence,
	}
	b, err := json.Marshal(&receipt)
	if err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
	}
	rsvp.Write(b)
	xlog.Trace.Println("guest", sub)
}

func (reg *registry) deny(rsvp *rsvp) {
	name := rsvp.subscriber()
	if len(name) == 0 {
		rsvp.Header().Add(RestError, "incomplete subscriber")
		rsvp.WriteHeader(http.StatusBadRequest)
	}
	if _, ok := reg.pending[name]; !ok {
		rsvp.Header().Add(RestError, name)
		rsvp.WriteHeader(http.StatusNotFound)
	} else {
		delete(reg.pending, name)
	}
}

func (reg *registry) dumpSubscribers(rsvp *rsvp) {
	blk := pem.Block{
		Type: "CERTIFICATE",
	}
	for _, sub := range reg.named {
		blk.Bytes = sub.CertDER
		if err := pem.Encode(rsvp, &blk); err != nil {
			fmt.Fprintln(rsvp, err)
			break
		}
	}
}

func (reg *registry) fromUDP(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)

	fi := fid.Index()
	if fi >= len(reg.indexed) {
		xlog.Trace.Println("dropped from unknown", fid)
		return
	}
	from := reg.indexed[fi]
	if got, want := fid.Version(), from.Id.Version(); got != want {
		xlog.Trace.Println("dropped from", from.name(),
			"with version", got, "!=", want)
		return
	}

	if fid == tid {
		from.hellohello(ctx, m)
		return
	}

	ti := tid.Index()
	if ti >= len(reg.indexed) {
		xlog.Trace.Println("dropped from", from.name(), "to unknown")
		return
	}
	to := reg.indexed[ti]
	if got, want := tid.Version(), to.Id.Version(); got != want {
		xlog.Trace.Println("dropped from", from.name(),
			"to", to.name(),
			"with version", got, "!=", want)
		return
	}
	if !to.via.IsValid() {
		xlog.Trace.Println("dropped from", from.name(),
			"to unaddressed", to.name())
		return
	}
	udp.WriteToUDPAddrPort(m.Data, to.via)
	xlog.Trace.Println("forward from", from.name(), "to", to.name())
}

func (reg *registry) getFileOrDir(w http.ResponseWriter, req *http.Request) {
	name := strings.TrimLeft(req.URL.Path, "/")
	if len(name) == 0 {
		names := []string{vlink()}
		if entries, err := os.ReadDir(vpnDataDir); err == nil {
			for _, sub := range entries {
				names = append(names, sub.Name())
			}
		}
		slices.Sort(names)
		for _, s := range names {
			fmt.Fprintln(w, s)
		}
		return
	} else if name == vlink() {
		f, err := os.Open(xprogram.Path())
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Fprint(w, err)
		} else {
			io.Copy(w, f)
			f.Close()
		}
		return
	}

	name = filepath.Join(vpnDataDir, name)
	if fi, err := os.Stat(name); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, fs.ErrPermission) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Fprint(w, err)
	} else if fi.IsDir() {
		if des, err := os.ReadDir(name); err != nil {
			if errors.Is(err, fs.ErrPermission) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Fprint(w, err)
		} else {
			var names []string
			for _, de := range des {
				names = append(names, de.Name())
			}
			slices.Sort(names)
			for _, s := range names {
				fmt.Fprintln(w, s)
			}
		}
	} else if f, err := os.Open(name); err != nil {
		if errors.Is(err, fs.ErrPermission) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Fprint(w, err)
	} else {
		_, err = io.Copy(w, f)
		f.Close()
	}
}

func (reg *registry) isMember(
	rsvp *rsvp, club map[string]*Subscriber,
) bool {
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	cn := peer0.Subject.CommonName
	sub, ok := club[cn]
	t := ok && peer0.Equal(sub.cert)
	if !t {
		rsvp.Header().Add(RestError, cn)
		rsvp.WriteHeader(http.StatusForbidden)
	}
	return t
}

func (reg *registry) isAdmin(rsvp *rsvp) bool {
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	return peer0.Equal(reg.cert) || reg.isMember(rsvp, reg.admin)
}

func (reg *registry) isSubscriber(rsvp *rsvp) bool {
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	return peer0.Equal(reg.cert) || reg.isMember(rsvp, reg.named)
}

func (reg *registry) loadAdminsFile() error {
	return xerrors.Suppress(kvc.
		RangeFile(vpnAdminsFile, reg.loadAdminsKeyValues),
		fs.ErrNotExist)
}

func (reg *registry) loadAdminsKeyValues(
	lno int, key string, values []string,
) error {
	reg.admin[key] = nil
	return nil
}

func (reg *registry) loadCerts() error {
	var err error

	reg.cert, err = readCertificateFile(vpnCertFile)
	if err != nil {
		return err
	}
	cn := reg.cert.Subject.CommonName
	sub := NewSubscriber(reg.cert)
	if err = reg.assignAddr(sub); err != nil {
		return err
	}
	sub.Id = 0
	reg.indexed = append(reg.indexed, sub)
	reg.named[cn] = sub

	var fns []string
	for _, dn := range []string{vpnConfigDir, vpnStateDir} {
		matches, err := filepath.Glob(filepath.Join(dn, "*.pem"))
		if err != nil {
			return err
		}
		fns = append(fns, matches...)
	}
	for _, fn := range fns {
		if fn == vpnCertFile {
			continue
		}
		c, err := readCertificateFile(fn)
		if err != nil {
			return err
		}
		cn = c.Subject.CommonName
		if _, exists := reg.named[cn]; exists {
			return fmt.Errorf("%s: duplicate", cn)
		}
		sub = NewSubscriber(c)
		if err = reg.assignAddr(sub); err != nil {
			return err
		}
		sub.Id = Id(len(reg.indexed))
		reg.indexed = append(reg.indexed, sub)
		reg.named[cn] = sub
		if _, isAdmin := reg.admin[cn]; isAdmin {
			reg.admin[cn] = sub
		}
		sub.ExchangePrecedence = reg.
			subscriberExchangePrecedence[sub.name()]
	}
	return nil
}

func (reg *registry) loadHostsFile() error {
	return xerrors.Suppress(kvc.
		RangeFile(vpnHostsFile, reg.loadHostsKeyValues),
		fs.ErrNotExist)
}

func (reg *registry) loadHostsKeyValues(
	lno int, key string, values []string,
) error {
	name := vpnHostsFile
	if len(values) < 0 {
		return xerrors.Incomplete(name, lno)
	}
	addr, err := netip.ParseAddr(key)
	if err != nil {
		return xerrors.Label(err, name, lno)
	}
	if !vpnPrefix.Contains(addr) {
		return xerrors.Range(name, lno)
	}
	reg.hosts.name[addr] = values[0]
	for _, hn := range values {
		reg.hosts.addr[hn] = addr
	}
	return nil
}

func (reg *registry) loadViaFile() error {
	return xerrors.Suppress(kvc.
		RangeFile(vpnViaFileName, reg.loadViaKeyValues),
		fs.ErrNotExist)
}

func (reg *registry) loadViaKeyValues(
	lno int, key string, values []string,
) error {
	name := vpnViaFileName
	if len(values) < 0 {
		return xerrors.Incomplete(name, lno)
	}
	reg.subscriberExchangePrecedence[name] = values
	return nil
}

func (reg *registry) rest(rsvp *rsvp) {
	defer func() { rsvp.ch <- done }()
	if len(rsvp.req.TLS.PeerCertificates) == 0 ||
		rsvp.req.TLS.PeerCertificates[0] == nil {
		rsvp.Header().Add(RestError, "no peer")
		rsvp.WriteHeader(http.StatusForbidden)
		return
	}
	switch op := rsvp.qv.Get(RestOp); op {
	case RestOpApprove:
		if rsvp.isMethod(http.MethodPut) && reg.isAdmin(rsvp) {
			reg.approve(rsvp)
		}
	case RestOpCertify:
		rsvp.isMethod(http.MethodGet)
		// empty response so that client may retrieve this
		// certificate from TLS negotiation.
	case RestOpCheckin:
		if rsvp.isMethod(http.MethodPut) && reg.isSubscriber(rsvp) {
			switch ci := rsvp.qv.Get(RestOpCheckin); ci {
			case "":
				rsvp.Header().Add(RestError, "incomlete")
				rsvp.WriteHeader(http.StatusBadRequest)
			case RestOpCheckinExchange:
				reg.checkinExchange(rsvp)
			case RestOpCheckinGuest:
				reg.checkinGuest(rsvp)
			default:
				rsvp.Header().Add(RestError,
					fmt.Sprint(ci, ": invalid"))
				rsvp.WriteHeader(http.StatusBadRequest)
			}
		}
	case RestOpDeny:
		if rsvp.isMethod(http.MethodPut) && reg.isAdmin(rsvp) {
			reg.deny(rsvp)
		}
	case RestOpDump:
		if rsvp.isMethod(http.MethodGet) && reg.isSubscriber(rsvp) {
			switch dumpreq := rsvp.qv.Get(RestOpDump); dumpreq {
			case RestOpDumpSubscribers:
				reg.dumpSubscribers(rsvp)
			default:
				rsvp.Header().Add(RestError, dumpreq)
				rsvp.WriteHeader(http.StatusBadRequest)
			}
		}
	case RestOpPing:
		if rsvp.isMethod(http.MethodGet) && reg.isSubscriber(rsvp) {
			fmt.Fprintln(rsvp, "OK")
		}
	case RestOpReload:
		if rsvp.isMethod(http.MethodPut) && reg.isAdmin(rsvp) {
			// FIXME reg.reload(rsvp)
		}
	case RestOpShow:
		if rsvp.isMethod(http.MethodGet) && reg.isSubscriber(rsvp) {
			switch subject := rsvp.qv.Get(RestOpShow); subject {
			case RestOpShowActive:
				reg.showActive(rsvp)
			case RestOpShowAddress:
				reg.showAddress(rsvp)
			case RestOpShowAdmins:
				keys := xmaps.Keys(reg.admin)
				slices.Sort(keys)
				for _, k := range keys {
					fmt.Fprintln(rsvp, "-", k)
				}
			case RestOpShowExchanges:
				/* FIXME
				for _, s := range vpn.cfg.Exchanges {
					fmt.Fprintln(w, "-", s)
				}
				*/
			case RestOpShowHosts:
				reg.showHosts(rsvp)
			case RestOpShowPending:
				reg.showPending(rsvp)
			case RestOpShowPrefix:
				fmt.Fprintln(rsvp, vpnPrefix)
			case RestOpShowStart:
				// no output, in response trailer
			case RestOpShowSubscriber:
				reg.showSubscriber(rsvp)
			case RestOpShowVcs:
				switch rsvp.qv.Get(RestOpShowVcs) {
				case RestOpShowVcsModified:
					fmt.Fprintln(rsvp, xprogram.VcsModified)
				case RestOpShowVcsRevision:
					// no output, use response trailer
				}
			default:
				rsvp.Header().Add(RestError, subject)
				rsvp.WriteHeader(http.StatusBadRequest)
			}
		}
	case RestOpSubscribe:
		if rsvp.isMethod(http.MethodPut) {
			reg.subscribe(rsvp)
		}
	case RestOpUnsubscribe:
		if rsvp.isMethod(http.MethodPut) && reg.isSubscriber(rsvp) {
			reg.unsubscribe(rsvp)
		}
	case RestOpWhois:
		if rsvp.isMethod(http.MethodGet) && reg.isSubscriber(rsvp) {
			reg.whois(rsvp)
		}
	default:
		rsvp.Header().Add(RestError, op)
		rsvp.WriteHeader(http.StatusBadRequest)
	}
}

func (reg *registry) restsvc() {
	xlog.Trace.Println("rest start")
	err := reg.http.ListenAndServeTLS(vpnCertFile, vpnSigFile)
	err = xerrors.Suppress(err, http.ErrServerClosed)
	if err == nil {
		xlog.Trace.Println("rest stopped")
	} else {
		xlog.Errata.Println("rest", err)
	}
}

func (reg *registry) rsvp(
	w http.ResponseWriter, req *http.Request, qv url.Values,
) {
	rsvp := rsvpPool.Get().(*rsvp)
	rsvp.ResponseWriter = w
	rsvp.req = req
	rsvp.qv = qv
	reg.rsvpC <- rsvp
	select {
	case <-req.Context().Done():
	case <-rsvp.ch:
		rsvp.ResponseWriter = nil
		rsvp.req = nil
		rsvp.qv = nil
		rsvpPool.Put(rsvp)
	}
}

func (reg *registry) showActive(rsvp *rsvp) {
	for _, sub := range reg.named {
		if len(sub.CipherKeyDER) == 0 {
			continue
		}
		fmt.Fprintln(rsvp, sub)
	}
}

func (reg *registry) showAddress(rsvp *rsvp) {
	if rsvp.qv.Has("arg0") {
		name := rsvp.qv.Get("arg0")
		sub, ok := reg.named[name]
		if !ok {
			rsvp.Header().Add(RestError, name)
			rsvp.WriteHeader(http.StatusNotFound)
		}
		fmt.Fprintln(rsvp, sub.Addr)
	} else {
		for _, sub := range reg.indexed {
			if sub.Addr.IsValid() {
				fmt.Fprintf(rsvp, "%s: %v\n",
					sub.name(), sub.Addr)
			}
		}
	}
}

func (reg *registry) showHosts(rsvp *rsvp) {
	if rsvp.qv.Has("arg0") {
		addr, err := netip.ParseAddr(rsvp.qv.Get("arg0"))
		if err != nil {
			rsvp.Header().Add(RestError, err.Error())
			rsvp.WriteHeader(http.StatusBadRequest)
		} else if name, ok := reg.hosts.name[addr]; !ok {
			rsvp.Header().Add(RestError, addr.String())
			rsvp.WriteHeader(http.StatusNotFound)
		} else {
			fmt.Fprintln(rsvp, name)
		}
	} else {
		addrs := make([]netip.Addr, 0, len(reg.hosts.name))
		for addr := range reg.hosts.name {
			addrs = append(addrs, addr)
		}
		slices.SortFunc(addrs, func(a, b netip.Addr) int {
			return a.Compare(b)
		})
		for _, addr := range addrs {
			name := reg.hosts.name[addr]
			fmt.Fprintf(rsvp, "%v\t%s\n", addr, name)
		}
	}
}

func (reg *registry) showPending(rsvp *rsvp) {
	if t, err := CertificatesTemplate(); err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
	} else {
		t.Execute(rsvp, reg.pending)
	}
}

func (reg *registry) showSubscriber(rsvp *rsvp) {
	if !rsvp.qv.Has("arg0") {
		names := xmaps.Keys(reg.named)
		slices.Sort(names)
		for _, name := range names {
			fmt.Fprintln(rsvp, name)
		}

	} else {
		name := rsvp.qv.Get("arg0")
		if sub, ok := reg.named[name]; !ok {
			rsvp.Header().Add(RestError, name)
			rsvp.WriteHeader(http.StatusNotFound)
		} else {
			fmt.Fprintln(rsvp, sub)
		}
	}
}

func (reg *registry) shutdown(ctx context.Context) {
	<-ctx.Done()
	ctx, cancel := context.
		WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	xlog.Trace.Print("shutdown ", reg.http.Addr, "...")
	reg.http.Shutdown(ctx)
}

// This has an empty response.  The client will retrieve the server cert
// through its TLS negotiation; then prompt the user to ise as root certificate
// authority.
func (reg *registry) subscribe(rsvp *rsvp) {
	c := rsvp.req.TLS.PeerCertificates[0]
	cn := c.Subject.CommonName
	if sub, found := reg.named[cn]; found {
		if c.Equal(sub.cert) {
			rsvp.Header().Add(RestError, "already subscribed")
			rsvp.WriteHeader(http.StatusGone)
		} else {
			rsvp.Header().Add(RestError, "name in use")
			rsvp.WriteHeader(http.StatusForbidden)
		}
	} else if _, found = reg.pending[cn]; found {
		rsvp.Header().Add(RestError, "subscription pending approval")
		rsvp.WriteHeader(http.StatusConflict)
	} else {
		reg.pending[cn] = NewSubscriber(c)
	}
}

func (reg *registry) unsubscribe(rsvp *rsvp) {
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	name := rsvp.subscriber()
	if len(name) == 0 {
		name = peer0.Subject.CommonName
	}
	if sub, found := reg.named[name]; !found {
		rsvp.Header().Add(RestError, name)
		rsvp.WriteHeader(http.StatusNotFound)
	} else if !peer0.Equal(sub.cert) && !reg.isAdmin(rsvp) {
		rsvp.Header().Add(RestError, fmt.Sprint("unsubscribe ", name))
		rsvp.WriteHeader(http.StatusForbidden)
	} else {
		delete(reg.addressed, sub.Addr)
		delete(reg.admin, name)
		delete(reg.named, name)
		reg.indexed[int(sub.Id)] = nil
		os.Remove(sub.stateFileName())
	}
}

func (reg *registry) whois(rsvp *rsvp) {
	var k, s string
	var sub *Subscriber
	var ok bool

	switch {
	case rsvp.qv.Has(RestWhoisAddress):
		k = RestWhoisAddress
		s = rsvp.qv.Get(RestWhoisAddress)
		addr, err := netip.ParseAddr(s)
		if err != nil {
			rsvp.Header().Add(RestError, err.Error())
			rsvp.WriteHeader(http.StatusBadRequest)
			return
		}
		sub, ok = reg.addressed[addr]
	case rsvp.qv.Has(RestWhoisId):
		k = RestWhoisId
		s = rsvp.qv.Get(RestWhoisId)
		id, err := ParseId(s)
		if err != nil {
			rsvp.Header().Add(RestError, err.Error())
			rsvp.WriteHeader(http.StatusBadRequest)
			return
		}
		i := id.Index()
		if i < len(reg.indexed) {
			sub, ok = reg.indexed[i], true
		}
	case rsvp.qv.Has(RestWhoisName):
		k = RestWhoisName
		s = rsvp.qv.Get(RestWhoisName)
		sub, ok = reg.named[s]
	default:
		rsvp.Header().Add(RestError,
			"no <address>, <id>, or <name>")
		rsvp.WriteHeader(http.StatusBadRequest)
		return
	}
	if !ok || sub == nil {
		rsvp.Header().Add(RestError, fmt.Sprint(k, " ", s))
		rsvp.WriteHeader(http.StatusNotFound)
	} else if b, err := json.MarshalIndent(sub, "", "  "); err != nil {
		rsvp.Header().Add(RestError, err.Error())
		rsvp.WriteHeader(http.StatusInternalServerError)
	} else {
		rsvp.Write(b)
	}
}
