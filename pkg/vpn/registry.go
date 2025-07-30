// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
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
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

type GuestReceipt struct {
	Id     Id
	Prefix netip.Prefix

	ExchangePrecedence []string
}

type invitation struct {
	from, to Id
	text     []byte
}

type registry struct {
	start, vcsRev string

	cert *x509.Certificate

	indexed []*Subscriber

	invitations []*invitation

	addressed map[netip.Addr]*Subscriber

	subscriberExchangePrecedence map[string][]string

	admin,
	named, // guest or exchange
	pending map[string]*Subscriber

	rsvpC chan *rsvp

	fromVpnC <-chan *xnet.Msg
	toVpnC   chan<- *xnet.Msg

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
	defineExchangePort()
	defineDomain()
	defineHosts()
	definePrefix()
	defineRegistryPort()
	defineSig()
	defineVia()

	enableQuiet()
	enableTrace()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if len(vpnDomain) == 0 {
		return xerrors.Invalid("domain")
	}

	if !strings.HasSuffix(vpnDomain, ".") {
		vpnDomain += "."
	}
	if !strings.HasPrefix(vpnDomain, ".") {
		vpnDomain = "." + vpnDomain
	}
	if err = signInit(); err != nil {
		return err
	}

	MyId = Id(0)
	MyLabel = MakeLabel(MyId, MyId)

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

	defer xlog.Trace.Println("stopped main")
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)

	reg.fromVpnC, reg.toVpnC, err = startUDP(vpnExchangePort)
	if err != nil {
		return err
	}
	defer close(reg.toVpnC)

	reg.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", vpnRegistryPort),
		Handler: reg,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			ClientAuth: tls.RequestClientCert,
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

	xlog.Trace.Println("start main")
	defer cancel()
	defer xlog.Trace.Println("stopping main...")

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
		case m, ok := <-reg.fromVpnC:
			if !ok {
				break selection
			}
			if len(m.Data) < SizeofLabel {
				xlog.Errata.Print("underrun")
				mp.Put(m)
			} else {
				reg.fromVpn(ctx, m)
			}
		}
	}
	return nil
}

func (reg *registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if !req.TLS.HandshakeComplete {
		http.Error(w, "incomplete handshake", http.StatusUnauthorized)
		return
	}

	w.Header().Add(RestUnixMicroStart, reg.start)
	w.Header().Add(RestVcsRevision, reg.vcsRev)

	if strings.HasPrefix(req.URL.Path, RestPathStatic) {
		if req.Method != http.MethodGet {
			http.Error(w, req.Method, http.StatusMethodNotAllowed)
		} else {
			name := strings.TrimPrefix(req.URL.Path, RestPathStatic)
			reg.getFileOrDir(w, name)
		}
		return
	}

	// single thread REST operations
	rsvp := rsvpPool.Get().(*rsvp)
	rsvp.ResponseWriter = w
	rsvp.req = req

	ctx := req.Context()
	select {
	case <-ctx.Done():
		rsvpPool.Put(rsvp)
	case reg.rsvpC <- rsvp:
		select {
		case <-ctx.Done():
		case <-rsvp.doneC:
			rsvp.ResponseWriter = nil
			rsvp.req = nil
			rsvpPool.Put(rsvp)
		}
	}
}

func (reg *registry) approve(rsvp *rsvp) {
	name := rsvp.trimPrefix(RestPathApprove)
	if len(name) == 0 {
		http.Error(rsvp, "incomplete subscriber", http.StatusBadRequest)
		return
	}
	_, err := os.Stat(vpnStateDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(vpnStateDir, 0755)
		}
		if err != nil {
			http.Error(rsvp, err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
	sub, ok := reg.pending[name]
	if !ok {
		http.Error(rsvp, name, http.StatusNotFound)
		return
	}
	blk := pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: sub.CertDER,
	}
	if f, err := os.Create(sub.stateFileName()); err != nil {
		http.Error(rsvp, err.Error(),
			http.StatusInternalServerError)
		return
	} else if err = pem.Encode(f, &blk); err != nil {
		f.Close()
		http.Error(rsvp, err.Error(),
			http.StatusInternalServerError)
		return
	} else {
		f.Close()
	}
	delete(reg.pending, name)

	if err = reg.assignAddr(sub); err != nil {
		http.Error(rsvp, err.Error(),
			http.StatusInternalServerError)
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
	if !reg.hasCertificate(rsvp) {
		return nil
	}

	peer0 := rsvp.req.TLS.PeerCertificates[0]
	cn := peer0.Subject.CommonName

	sub, ok := reg.named[cn]
	if !ok {
		http.Error(rsvp, cn, http.StatusNotFound)
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
	s := rsvp.trimPrefix(RestPathCheckinExchange)
	if len(s) > 0 {
		u, err := strconv.ParseUint(s, 0, 16)
		if err != nil {
			http.Error(rsvp, err.Error(), http.StatusBadRequest)
			return
		}
		sub.Port = uint16(u)
	} else {
		http.Error(rsvp, "no port", http.StatusBadRequest)
		return
	}
	fmt.Fprint(rsvp, uint(sub.Id))
	xlog.Trace.Println("checked in exchange", sub)
}

func (reg *registry) checkinGuest(rsvp *rsvp) {
	var err error

	sub := reg.checkin(rsvp)
	if sub == nil {
		return
	}

	sub.EncapKey, err = io.ReadAll(rsvp.req.Body)
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
		return
	}

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
		http.Error(rsvp, err.Error(), http.StatusInternalServerError)
	} else {
		rsvp.Write(b)
		xlog.Trace.Println("checked in guest", sub)
	}
}

func (reg *registry) deny(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathDeny)
	if len(s) == 0 {
		http.Error(rsvp, "incomplete subscriber", http.StatusBadRequest)
	}
	if _, ok := reg.pending[s]; !ok {
		http.Error(rsvp, s, http.StatusNotFound)
	} else {
		delete(reg.pending, s)
	}
}

func (reg *registry) dnsAnswer(query *xdnsmessage.Message) *xdnsmessage.Message {
	reply := xdnsmessage.NewMessage()
	reply.Addr = query.Addr
	reply.ID = query.ID
	reply.HF = xdnsmessage.HFResponse
	reply.OpCode = query.OpCode
	reply.Questions = query.Questions
	if len(query.Questions) == 0 {
		reply.RCode = xdnsmessage.RCodeFormatError
		return reply
	}
	if query.OpCode != xdnsmessage.OpCodeQuery {
		reply.RCode = xdnsmessage.RCodeNotImplemented
		return reply
	}
	for _, q := range query.Questions {
		uniqueName := q.Name
		name := uniqueName.String()
		if !strings.HasSuffix(name, ".") {
			name += "."
		}
		if !strings.HasSuffix(name, vpnDomain) {
			xlog.Trace.Print("domain(", name, ") !=", vpnDomain)
			continue
		}
		name = strings.TrimSuffix(name, vpnDomain)
		sub, ok := reg.named[name]
		if !ok {
			xlog.Trace.Println(name, "not found")
			continue
		}
		reply.RCode = xdnsmessage.RCodeSuccess
		var resource xdnsmessage.TypedResource
		if sub.Addr.Is4() {
			resource = xdnsmessage.TypeAResource{sub.Addr}
		} else if sub.Addr.Is6() {
			resource = xdnsmessage.TypeAAAAResource{sub.Addr}
		}
		if resource != nil {
			xlog.Trace.Println(name, sub.Addr)
			reply.Answers = append(reply.Answers,
				xdnsmessage.WireResource{
					Name:          uniqueName,
					Duration:      3600 * time.Second,
					Class:         xdnsmessage.ClassINET,
					TypedResource: resource,
				})
		}
	}
	return reply
}

func (reg *registry) dnsQuery(rsvp *rsvp) {
	var b []byte
	var err error

	if rsvp.req.Method == http.MethodPost {
		b, err = io.ReadAll(rsvp.req.Body)
	} else if ev := rsvp.req.URL.Query().Get("dns"); len(ev) == 0 {
		err = xerrors.Incomplete("dns")
	} else if b, err = base64.StdEncoding.DecodeString(ev); err != nil {
		b, err = base64.URLEncoding.DecodeString(ev)
	}
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
		xlog.Errata.Println(err)
		return
	}

	query := xdnsmessage.NewMessage()
	defer query.Free()

	if err = query.UnmarshalBinary(b); err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
		xlog.Errata.Println(err)
		return
	}

	answer := reg.dnsAnswer(query)
	defer answer.Free()

	rsvp.Header().Set("Content-Type", "application/dns-message")
	b, err = answer.AppendTo(make([]byte, 0, 4<<10))
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusInternalServerError)
		xlog.Errata.Println(err)
	} else if _, err = rsvp.Write(b); err != nil {
		xlog.Errata.Print(err)
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

func (reg *registry) fromVpn(ctx context.Context, m *xnet.Msg) {
	tid, fid := ScanLabel(m.Data)

	if fi := fid.Index(); fi >= len(reg.indexed) {
		xlog.Trace.Println("dropped from unknown", fid)
	} else if from := reg.indexed[fi]; from.Id.Version() != fid.Version() {
		xlog.Trace.Print("dropped ", from.name(),
			", version ", from.Id.Version(), " != ", fid.Version())
	} else if fid == tid {
		if from.helloIsOK(m) {
			from.via = m.AddrPort
			if hello := newGreeting(0); hello != nil {
				hello.AddrPort = m.AddrPort
				mp.Queue(ctx, reg.toVpnC, hello)
			}
		}
	} else if ti := tid.Index(); ti >= len(reg.indexed) {
		xlog.Trace.Println("dropped", from.name(), "-> unknown")
	} else if to := reg.indexed[ti]; to.Id.Version() != tid.Version() {
		xlog.Trace.Print("dropped ", from.name(), " -> ", to.name(),
			", version ", to.Id.Version(), " != ", tid.Version())
	} else if !to.via.IsValid() {
		xlog.Trace.Println("dropped", from.name(), "-> unaddressed",
			to.name())
	} else {
		m.AddrPort = to.via
		mp.Queue(ctx, reg.toVpnC, m)
		xlog.Trace.Println("forward", from.name(), "->", to.name())
		return
	}
	mp.Put(m)
}

func (reg *registry) getFileOrDir(w http.ResponseWriter, name string) {
	xlog.Trace.Println(http.MethodGet, name)
	ecode := http.StatusInternalServerError
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
				ecode = http.StatusForbidden
			}
			http.Error(w, err.Error(), ecode)
		} else {
			io.Copy(w, f)
			f.Close()
		}
		return
	}

	name = filepath.Join(vpnDataDir, name)
	if fi, err := os.Stat(name); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			ecode = http.StatusNotFound
		} else if errors.Is(err, fs.ErrPermission) {
			ecode = http.StatusForbidden
		}
		http.Error(w, err.Error(), ecode)
	} else if fi.IsDir() {
		if des, err := os.ReadDir(name); err != nil {
			if errors.Is(err, fs.ErrPermission) {
				ecode = http.StatusForbidden
			}
			http.Error(w, err.Error(), ecode)
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
			ecode = http.StatusForbidden
		}
		http.Error(w, err.Error(), ecode)
	} else {
		_, err = io.Copy(w, f)
		f.Close()
	}
}

func (reg *registry) hasCertificate(rsvp *rsvp) bool {
	t := len(rsvp.req.TLS.PeerCertificates) > 0 &&
		rsvp.req.TLS.PeerCertificates[0] != nil
	if !t {
		const emsg = "no client certificate"
		xlog.Errata.Println(emsg)
		http.Error(rsvp, emsg, http.StatusUnauthorized)
	}
	return t
}

func (reg *registry) invite(rsvp *rsvp) {
	from := reg.named[rsvp.req.TLS.PeerCertificates[0].Subject.CommonName]
	text, err := io.ReadAll(rsvp.req.Body)
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
	}
	s := rsvp.trimPrefix(RestPathInvite)
	if len(s) == 0 {
		http.Error(rsvp, "incomplete subscriber", http.StatusBadRequest)
	}
	to, ok := reg.named[s]
	if !ok {
		xlog.Errata.Println(s, "not found in", xmaps.Keys(reg.named))
		http.Error(rsvp, s, http.StatusNotFound)
	}
	for i, x := range reg.invitations {
		if x.from == to.Id && x.to == from.Id {
			rsvp.Write(x.text)
			reg.invitations = slices.Delete(reg.invitations, i, i+1)
			return
		}
	}
	reg.invitations = append(reg.invitations, &invitation{
		from: from.Id,
		to:   to.Id,
		text: text,
	})
}

func (reg *registry) isMember(
	rsvp *rsvp, club map[string]*Subscriber,
) bool {
	if !reg.hasCertificate(rsvp) {
		return false
	}
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	if peer0.Equal(reg.cert) {
		return true
	}
	cn := peer0.Subject.CommonName
	sub, ok := club[cn]
	t := ok && peer0.Equal(sub.cert)
	if !t {
		http.Error(rsvp, cn, http.StatusForbidden)
	}
	return t
}

func (reg *registry) isAdmin(rsvp *rsvp) bool {
	return reg.isMember(rsvp, reg.admin)
}

func (reg *registry) isSubscriber(rsvp *rsvp) bool {
	return reg.isMember(rsvp, reg.named)
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

func (reg *registry) reload(rsvp *rsvp) {
	fmt.Fprint(rsvp, "FIXME reload admins, hosts, and via")
}

func (reg *registry) rest(rsvp *rsvp) {
	defer rsvp.done()
	method, path := rsvp.req.Method, rsvp.req.URL.Path
	xlog.Trace.Println(method, path)
	switch method {
	case http.MethodGet:
		switch {
		case path == RestPathCertify:
			// empty response so that client may retrieve this
			// certificate from TLS negotiation.
		case path == RestPathDnsQuery:
			reg.dnsQuery(rsvp)
		case strings.HasPrefix(path, RestPathShowAddress):
			if reg.isSubscriber(rsvp) {
				reg.showAddress(rsvp)
			}
		case strings.HasPrefix(path, RestPathShowAdmins):
			if reg.isSubscriber(rsvp) {
				keys := xmaps.Keys(reg.admin)
				slices.Sort(keys)
				for _, k := range keys {
					fmt.Fprintln(rsvp, "-", k)
				}
			}
		case path == RestPathShowExchanges:
			if reg.isSubscriber(rsvp) {
				for _, sub := range reg.indexed {
					if sub.Port != 0 {
						fmt.Fprintln(rsvp, sub)
					}
				}
			}
		case path == RestPathShowGuests:
			if reg.isSubscriber(rsvp) {
				for _, sub := range reg.indexed {
					if len(sub.EncapKey) > 0 {
						fmt.Fprintln(rsvp, sub)
					}
				}
			}
		case strings.HasPrefix(path, RestPathShowHosts):
			if reg.isSubscriber(rsvp) {
				reg.showHosts(rsvp)
			}
		case path == RestPathShowPending:
			if reg.isSubscriber(rsvp) {
				reg.showPending(rsvp)
			}
		case path == RestPathShowPrefix:
			if reg.isSubscriber(rsvp) {
				fmt.Fprintln(rsvp, vpnPrefix)
			}
		case path == RestPathShowStart:
			fmt.Fprintln(rsvp, time.UnixMicro(vpnStart))
		case path == RestPathShowStatus:
			if reg.isSubscriber(rsvp) {
				fmt.Fprintln(rsvp, "OK")
			}
		case strings.HasPrefix(path, RestPathShowSubscriber):
			if reg.isSubscriber(rsvp) {
				reg.showSubscriber(rsvp)
			}
		case path == RestPathShowVCS:
			fmt.Fprint(rsvp, xprogram.VcsRevision)
			if xprogram.VcsModified.String() == "true" {
				fmt.Fprint(rsvp, "*")
			}
			fmt.Fprintln(rsvp)
		case strings.HasPrefix(path, RestPathWhoisAddressed):
			if reg.isSubscriber(rsvp) {
				reg.whoisAddressed(rsvp)
			}
		case strings.HasPrefix(path, RestPathWhoisId):
			if reg.isSubscriber(rsvp) {
				reg.whoisId(rsvp)
			}
		case strings.HasPrefix(path, RestPathWhoisNamed):
			if reg.isSubscriber(rsvp) {
				reg.whoisNamed(rsvp)
			}
		default:
			http.Error(rsvp, path, http.StatusNotFound)
		}
	case http.MethodPost:
		switch {
		case path == RestPathDnsQuery:
			reg.dnsQuery(rsvp)
		default:
			http.Error(rsvp, path, http.StatusNotFound)
		}
	case http.MethodPut:
		switch {
		case strings.HasPrefix(path, RestPathApprove):
			if reg.isAdmin(rsvp) {
				reg.approve(rsvp)
			}
		case strings.HasPrefix(path, RestPathCheckinExchange):
			if reg.isSubscriber(rsvp) {
				reg.checkinExchange(rsvp)
			}
		case path == RestPathCheckinGuest:
			if reg.isSubscriber(rsvp) {
				reg.checkinGuest(rsvp)
			}
		case strings.HasPrefix(path, RestPathDeny):
			if reg.isAdmin(rsvp) {
				reg.deny(rsvp)
			}
		case path == RestPathDumpSubscribers:
			if reg.isSubscriber(rsvp) {
				reg.dumpSubscribers(rsvp)
			}
		case strings.HasPrefix(path, RestPathInvite):
			if reg.isSubscriber(rsvp) {
				reg.invite(rsvp)
			}
		case path == RestPathReload:
			if reg.isAdmin(rsvp) {
				reg.reload(rsvp)
			}
		case path == RestPathSubscribe:
			reg.subscribe(rsvp)
		case strings.HasPrefix(path, RestPathUnsubscribe):
			if reg.isSubscriber(rsvp) {
				reg.unsubscribe(rsvp)
			}
		default:
			http.Error(rsvp, path, http.StatusNotFound)
		}
	default:
		http.Error(rsvp, method, http.StatusMethodNotAllowed)
	}
}

func (reg *registry) restsvc() {
	xlog.Trace.Println("start rest", reg.http.Addr)
	err := reg.http.ListenAndServeTLS(vpnCertFile, vpnSigFile)
	err = xerrors.Suppress(err, http.ErrServerClosed)
	if err == nil {
		xlog.Trace.Println("stopped rest", reg.http.Addr)
	} else {
		xlog.Errata.Println("quit rest", reg.http.Addr, err)
	}
}

func (reg *registry) showAddress(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathShowAddress)
	if len(s) > 0 {
		sub, ok := reg.named[s]
		if !ok {
			http.Error(rsvp, s, http.StatusNotFound)
		} else {
			fmt.Fprintln(rsvp, sub.Addr)
		}
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
	s := rsvp.trimPrefix(RestPathShowHosts)
	if len(s) > 0 {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			http.Error(rsvp, err.Error(), http.StatusBadRequest)
		} else if name, ok := reg.hosts.name[addr]; !ok {
			http.Error(rsvp, addr.String(), http.StatusNotFound)
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
		http.Error(rsvp, err.Error(), http.StatusInternalServerError)
	} else {
		t.Execute(rsvp, reg.pending)
	}
}

func (reg *registry) showSubscriber(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathShowSubscriber)
	if len(s) == 0 {
		names := xmaps.Keys(reg.named)
		slices.Sort(names)
		for _, name := range names {
			fmt.Fprintln(rsvp, name)
		}
	} else if sub, ok := reg.named[s]; !ok {
		http.Error(rsvp, s, http.StatusNotFound)
	} else {
		fmt.Fprintln(rsvp, sub)
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
	if !reg.hasCertificate(rsvp) {
		return
	}
	c := rsvp.req.TLS.PeerCertificates[0]
	cn := c.Subject.CommonName
	if sub, found := reg.named[cn]; found {
		if c.Equal(sub.cert) {
			http.Error(rsvp, "already subscribed", http.StatusGone)
		} else {
			http.Error(rsvp, "name in use", http.StatusForbidden)
		}
	} else if _, found = reg.pending[cn]; found {
		http.Error(rsvp, "subscription pending approval",
			http.StatusConflict)
	} else {
		reg.pending[cn] = NewSubscriber(c)
	}
}

func (reg *registry) unsubscribe(rsvp *rsvp) {
	if !reg.hasCertificate(rsvp) {
		return
	}
	peer0 := rsvp.req.TLS.PeerCertificates[0]
	name := rsvp.trimPrefix(RestPathUnsubscribe)
	if len(name) == 0 {
		name = peer0.Subject.CommonName
	}
	if sub, found := reg.named[name]; !found {
		http.Error(rsvp, name, http.StatusNotFound)
	} else if !peer0.Equal(sub.cert) && !reg.isAdmin(rsvp) {
		http.Error(rsvp, fmt.Sprint("unsubscribe ", name),
			http.StatusForbidden)
	} else {
		delete(reg.addressed, sub.Addr)
		delete(reg.admin, name)
		delete(reg.named, name)
		reg.indexed[int(sub.Id)] = nil
		os.Remove(sub.stateFileName())
	}
}

func (reg *registry) whoisAddressed(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathWhoisAddressed)
	addr, err := netip.ParseAddr(s)
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
		return
	} else if sub, ok := reg.addressed[addr]; !ok || sub == nil {
		http.Error(rsvp, s, http.StatusNotFound)
	} else {
		reg.marshalSub(rsvp, sub)
	}
}

func (reg *registry) whoisId(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathWhoisId)
	id, err := ParseId(s)
	if err != nil {
		http.Error(rsvp, err.Error(), http.StatusBadRequest)
	} else if i := id.Index(); i >= len(reg.indexed) {
		http.Error(rsvp, s, http.StatusNotFound)
	} else {
		reg.marshalSub(rsvp, reg.indexed[i])
	}
}

func (reg *registry) whoisNamed(rsvp *rsvp) {
	s := rsvp.trimPrefix(RestPathWhoisNamed)
	sub, ok := reg.named[s]
	if !ok || sub == nil {
		http.Error(rsvp, s, http.StatusNotFound)
	} else {
		reg.marshalSub(rsvp, sub)
	}
}

func (reg *registry) marshalSub(rsvp *rsvp, sub *Subscriber) {
	if b, err := json.MarshalIndent(sub, "", "  "); err != nil {
		http.Error(rsvp, err.Error(), http.StatusInternalServerError)
	} else {
		rsvp.Write(b)
	}
}
