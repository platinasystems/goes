// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/ed25519"
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
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmaps"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"gopkg.in/yaml.v3"
)

type Config *struct {
	// Required network prefix.
	Prefix netip.Prefix
	// Optional static VPN address assignment. All others are dynamically
	// assigned starting at the next available address from the highest
	// entry. (default: Prefix.Addr)
	Address map[string]netip.Addr
	// Optional list of subscriber's that are enabled to
	// approve/deny/unscribe others.
	Admins []string
	// A required list of packet exchange names.
	Exchanges []string
}

// [Registry] unmarshals the [Flags.FN.Cfg] YAML file into ConfigByName.
var ConfigByName map[string]Config

type Registration struct {
	Name,
	// Via subscriber CommonName
	Via string
	Id Id
	// VPN tunnel address
	Addr netip.Addr
	// VPN service address
	Service netip.AddrPort
	DNSNames,
	IPAddresses,
	URIs []string
	SigAlg x509.SignatureAlgorithm
	SigDER,
	PubKey,
	CipherText []byte
}

func (reg *Registration) verifier() (func([]byte) bool, error) {
	pub, err := x509.ParsePKIXPublicKey(reg.SigDER)
	if err != nil {
		return nil, err
	}
	switch reg.SigAlg {
	case x509.UnknownSignatureAlgorithm:
		return nil, fmt.Errorf("%q: %w signature algorithm",
			reg.Name, xerrors.ErrIncomplete)
	case x509.PureEd25519:
		k, ok := pub.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%T isn't %v", pub, reg.SigAlg)
		}
		return pureEd25519{k}.verify, nil
	}
	return nil, fmt.Errorf("%q: %w signature algorithm: %v",
		reg.Name, xerrors.ErrUnsupported, reg.SigAlg)
}

// Registry is a web server providing a REST interface to persistent
// files and ephemeral tables.
func Registry(ctx context.Context, args []string) error {
	var reg registry

	xlog.SetPrefixes("registry/")

	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
A RESTful WWW server.

{{flags .}}`)

	defineCert()
	defineConfig()
	defineConfigDir()
	defineDataDir()
	defineSig()
	defineStateDir()

	enableQuiet()
	enableVerbose()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	cfg := filepath.Join(vpnConfigDir, vpnConfig)
	if _, err = os.Stat(cfg); err != nil {
		return err
	}

	cfn := filepath.Join(vpnConfigDir, vpnCert)
	if cs, err := certificates(cfn); err != nil {
		return err
	} else if len(cs) == 0 {
		return xerrors.Invalid(cfn)
	} else {
		reg.crt = cs[0]
	}

	svc := fmt.Sprint(":", defaultRegistryPort)
	if len(reg.crt.URIs) > 0 {
		if s := reg.crt.URIs[0].Port(); len(s) > 0 {
			svc = ":" + s
		}
	}

	reg.vpn = make(map[string]*regVpn)

	if err = reg.reload(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	reg.http = &http.Server{
		Addr:    svc,
		Handler: &reg,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			ClientAuth: tls.RequireAnyClientCert,
		},
		BaseContext: func(net.Listener) context.Context {
			return cctx
		},
	}

	now := time.Now().UnixMicro()
	reg.start = strconv.FormatInt(now, 10)
	reg.vcsRev = xprogram.VcsRevision.String()

	xlog.Info.Println("start", svc, "vcsrev", reg.vcsRev)
	defer xlog.Info.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer xlog.Info.Println("stopping", svc, "...")

	wg.Go(func() { go xlog.AlarmHandler(cctx) })
	wg.Go(func() { reg.shutdown(cctx) })

	err = reg.http.ListenAndServeTLS(cfn, PrivSigFileName())
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

type registry struct {
	mutex sync.RWMutex

	start,
	vcsRev string

	crt  *x509.Certificate
	url  *url.URL
	http *http.Server
	vpn  map[string]*regVpn
}

type regVpn struct {
	mutex sync.RWMutex
	name  string
	cfg   Config
	crt   *x509.Certificate

	addr struct {
		name  map[netip.Addr]string
		named map[string]netip.Addr
		top   netip.Addr
	}

	admin map[string]bool

	pending,
	subscribers []*x509.Certificate

	subscriberNamed map[string]*x509.Certificate

	idbook IdBook

	addressed map[netip.Addr]*Registration
	indexed   map[int]*Registration
	named     map[string]*Registration
}

func (reg *registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if !req.TLS.HandshakeComplete {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("incomplete handshake"))
		return
	}
	if len(req.TLS.PeerCertificates) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("no certificates"))
		return
	}
	peer0 := req.TLS.PeerCertificates[0]
	cn := peer0.Subject.CommonName

	qv := req.URL.Query()
	op := qv.Get(RestOp)

	if len(op) == 0 {
		reg.getFileOrDir(w, req)
		return
	}

	name := strings.TrimLeft(req.URL.Path, "/")
	if len(name) == 0 {
		name = "vpn"
	}
	vpn, found := reg.vpnNamed(name)
	if !found || vpn == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(name))
		return
	}

	w.Header().Add(RestUnixMicroStart, reg.start)
	w.Header().Add(RestVcsRevision, reg.vcsRev)

	var err error
	switch op {
	case RestOpApprove:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if err = vpn.selfOrAdmin(cn); err != nil {
		} else if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else if err = vpn.approve(req); err == nil {
			fmt.Fprintln(w, "OK")
		}
	case RestOpCertify:
		if req.Method != http.MethodGet {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else {
			// empty response so that client may retrieve this
			// certificate from TLS negotiation.
		}
	case RestOpCheckin:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else {
			switch ci := qv.Get(RestOpCheckin); ci {
			case "":
				err = xerrors.Incomplete(RestOpCheckin)
			case RestOpCheckinExchange:
				err = vpn.checkinExchange(w, req)
			case RestOpCheckinGuest:
				err = vpn.checkinGuest(w, req)
			default:
				err = xerrors.Invalid(RestOpCheckin, ci)
			}
			if errors.Is(err, xerrors.ErrUnavailable) {
				err = NewStatusError(http.StatusTooEarly)
			}
		}
	case RestOpDeny:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if err = vpn.selfOrAdmin(cn); err != nil {
		} else if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else if err = vpn.deny(req); err != nil {
		} else {
			fmt.Fprintln(w, "OK")
		}
	case RestOpDump:
		switch dumpreq := qv.Get(RestOpDump); dumpreq {
		case RestOpDumpSubscribers:
			if err = vpn.selfOrSubscriber(peer0); err != nil {
			} else if req.Method != http.MethodGet {
				err = NewStatusError(http.
					StatusMethodNotAllowed, req.Method)
			} else {
				err = vpn.dumpSubscribers(w)
			}
		default:
			err = NewStatusError(http.StatusBadRequest, dumpreq)
		}
	case RestOpPing:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if req.Method != http.MethodGet {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case RestOpReload:
		if name != "vpn" {
			err = NewStatusError(http.StatusNotAcceptable,
				name)
		} else if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else if err := reg.reload(); err == nil {
			fmt.Fprintln(w, "OK")
		}
	case RestOpShow:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if req.Method != http.MethodGet {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else {
			switch subject := qv.Get(RestOpShow); subject {
			case RestOpShowActive:
				err = vpn.showActive(w)
			case RestOpShowAddress:
				err = vpn.showAddress(w, qv)
			case RestOpShowAdmins:
				for _, s := range vpn.cfg.Admins {
					fmt.Fprintln(w, "-", s)
				}
			case RestOpShowExchanges:
				for _, s := range vpn.cfg.Exchanges {
					fmt.Fprintln(w, "-", s)
				}
			case RestOpShowHosts:
				err = vpn.showHosts(w, qv)
			case RestOpShowPending:
				err = vpn.showPending(w)
			case RestOpShowPrefix:
				fmt.Fprintln(w, vpn.cfg.Prefix)
			case RestOpShowStart:
				// no output, use response trailer
			case RestOpShowSubscriber:
				err = vpn.showSubscriber(w, qv)
			case RestOpShowTenant:
				err = vpn.showTenant(w, qv)
			case RestOpShowVcs:
				switch qv.Get(RestOpShowVcs) {
				case RestOpShowVcsModified:
					fmt.Fprintln(w, xprogram.VcsModified)
				case RestOpShowVcsRevision:
					// no output, use response trailer
				}
			default:
				err = NewStatusError(http.StatusBadRequest,
					op, " ", subject)
			}
		}
	case RestOpSubscribe:
		if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else if err = vpn.subscribe(req); err == nil {
			fmt.Fprintln(w, "OK")
		}
	case RestOpUnsubscribe:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
		} else if err = vpn.selfOrAdmin(cn); err != nil {
		} else if req.Method != http.MethodPut {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else if err := vpn.unsubscribe(req); err == nil {
			fmt.Fprintln(w, "OK")
		}
	case RestOpWhois:
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			err = NewStatusError(http.StatusForbidden,
				cn, ", ", err)
		} else if req.Method != http.MethodGet {
			err = NewStatusError(http.StatusMethodNotAllowed,
				req.Method)
		} else {
			var b []byte
			if b, err = vpn.whois(req); err == nil {
				xlog.Info.Printf("registration%s", b)
				w.Write(b)
			}
		}
	default:
		err = NewStatusError(http.StatusBadRequest, op)
	}
	if err == nil {
		w.(http.Flusher).Flush()
	} else if se, isStatusError := err.(*StatusError); isStatusError {
		w.Header().Add(RestError, se.s)
		w.WriteHeader(se.Code)
		xlog.Errata.Println(se.Code, se.s)
	} else {
		w.Header().Add(RestError, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		xlog.Errata.Print(err)
	}
}

// virtual-link, this program may be fetched as MAIN-GOOS-GOARCH
var vlink = sync.OnceValue(func() string {
	return fmt.Sprintf("%s-%s-%s", xprogram.MainName(),
		runtime.GOOS, runtime.GOARCH)
})

func (reg *registry) getFileOrDir(w http.ResponseWriter, req *http.Request) {
	name := strings.TrimLeft(req.URL.Path, "/")
	if len(name) == 0 {
		names := []string{vlink()}
		if entries, err := os.ReadDir(vpnDataDir); err == nil {
			for _, entry := range entries {
				names = append(names, entry.Name())
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
		if entries, err := os.ReadDir(name); err != nil {
			if errors.Is(err, fs.ErrPermission) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			fmt.Fprint(w, err)
		} else {
			var names []string
			for _, entry := range entries {
				names = append(names, entry.Name())
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

func (reg *registry) shutdown(ctx context.Context) {
	<-ctx.Done()
	cctx, cancel := context.
		WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	xlog.Info.Print("shutdown ", reg.http.Addr, "...")
	reg.http.Shutdown(cctx)
}

func (reg *registry) reload() error {
	data, err := os.ReadFile(filepath.Join(vpnConfigDir, vpnConfig))
	if err != nil {
		return err
	}
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	if err = yaml.Unmarshal(data, &ConfigByName); err != nil {
		return err
	}
	for name, cfg := range ConfigByName {
		if !cfg.Prefix.IsValid() {
			return xerrors.Invalid(name, "prefix")
		}

		vpn, found := reg.vpn[name]
		if !found {
			named := make(map[string]*x509.Certificate)
			vpn = &regVpn{
				name: name,
				cfg:  cfg,
				crt:  reg.crt,

				subscriberNamed: named,
			}

			vpn.addr.name = make(map[netip.Addr]string)
			vpn.addr.named = make(map[string]netip.Addr)

			vpn.addressed = make(map[netip.Addr]*Registration)
			vpn.indexed = make(map[int]*Registration)
			vpn.named = make(map[string]*Registration)

			reg.vpn[name] = vpn
		} else {
			vpn.cfg = cfg
		}

		vpn.addr.top = cfg.Prefix.Addr()
		for k, v := range cfg.Address {
			if a, found := vpn.addr.named[k]; found {
				delete(vpn.addr.name, a)
			}
			vpn.addr.name[v] = k
			vpn.addr.named[k] = v
			if vpn.addr.top.Compare(v) < 0 {
				vpn.addr.top = v
			}
		}

		var sources []string
		if name == "vpn" {
			sources = append(sources, vpnConfigDir)
		} else {
			sources = append(sources,
				filepath.Join(vpnConfigDir, vpnCert),
				filepath.Join(vpnConfigDir, vpnRegistry))
		}
		sources = append(sources, vpn.StateDir())
		for _, src := range sources {
			subs, err := certificates(src)
			if err == nil {
				vpn.subscribers = append(vpn.subscribers,
					subs...)
				for _, sub := range subs {
					cn := sub.Subject.CommonName
					vpn.subscriberNamed[cn] = sub
				}
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func (reg *registry) vpnNamed(name string) (*regVpn, bool) {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	vpn, ok := reg.vpn[name]
	return vpn, ok
}

func reqsub(req *http.Request) (string, error) {
	qv := req.URL.Query()
	if !qv.Has("subscriber") {
		return "", xerrors.Incomplete("subscriber")
	}
	return qv.Get(RestSubscriber), nil
}

func (vpn *regVpn) approve(req *http.Request) error {
	sub, err := reqsub(req)
	if err != nil {
		return NewStatusError(http.StatusInternalServerError, err)
	}
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	sd := vpn.StateDir()
	if _, err = os.Stat(sd); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(sd, 0755)
		}
		if err != nil {
			return NewStatusError(http.StatusInternalServerError,
				err)
		}
	}
	for i, c := range vpn.pending {
		if c.Subject.CommonName == sub {
			vpn.pending = slices.Delete(vpn.pending, i, i+1)
			err = addCertificate(sd, c)
			if err == nil {
				vpn.subscribers = append(vpn.subscribers, c)
				vpn.subscriberNamed[c.Subject.CommonName] = c
			}
			return NewStatusError(http.StatusInternalServerError,
				err)
		}
	}
	return NewStatusError(http.StatusNotFound, sub)
}

func (vpn *regVpn) checkin(peer *x509.Certificate) (*Registration, error) {
	der, err := x509.MarshalPKIXPublicKey(peer.PublicKey)
	if err != nil {
		return nil, err
	}

	cn := peer.Subject.CommonName
	reg, ok := vpn.named[cn]
	if ok {
		reg.Id.Revise()
	} else {
		reg = &Registration{
			Name: cn,
			Id:   vpn.idbook.New(),
		}
	}

	reg.DNSNames = peer.DNSNames
	for _, ip := range peer.IPAddresses {
		reg.IPAddresses = append(reg.IPAddresses, ip.String())
	}
	for _, uri := range peer.URIs {
		reg.URIs = append(reg.URIs, uri.String())
	}

	reg.SigAlg = peer.SignatureAlgorithm
	reg.SigDER = der

	vpn.indexed[reg.Id.Index()] = reg
	vpn.named[cn] = reg
	xlog.Info.Println("checkin:", cn, reg.Id, reg.SigAlg)
	return reg, nil
}

func (vpn *regVpn) checkinExchange(
	w http.ResponseWriter, req *http.Request,
) error {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	defer req.Body.Close()

	peer := req.TLS.PeerCertificates[0]
	cn := peer.Subject.CommonName
	reg, err := vpn.checkin(peer)
	if err != nil {
		return err
	}
	if qv := req.URL.Query(); qv.Has(RestOpCheckinExchangeVia) {
		reg.Via = qv.Get(RestOpCheckinExchangeVia)
	}
	if _, err = fmt.Fprint(w, uint(reg.Id)); err != nil {
		return err
	}
	xlog.Info.Printf("exchange %s %v", cn, reg.Id)
	return nil
}

func (vpn *regVpn) checkinGuest(
	w http.ResponseWriter, req *http.Request,
) error {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	defer req.Body.Close()

	pk, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	peer := req.TLS.PeerCertificates[0]
	cn := peer.Subject.CommonName
	reg, err := vpn.checkin(peer)
	if err != nil {
		return err
	}
	if qv := req.URL.Query(); qv.Has(RestOpCheckinGuestVia) {
		reg.Via = qv.Get(RestOpCheckinGuestVia)
	}

	reg.PubKey = pk

	if !reg.Addr.IsValid() {
		if addr, ok := vpn.cfg.Address[cn]; ok {
			reg.Addr = addr
		} else if reg.Addr, err = vpn.lease(cn); err != nil {
			return err
		}
		vpn.addressed[reg.Addr] = reg
	}
	w.Header().Set("Content-Type", "application/json")
	bits := ConfigByName[vpn.name].Prefix.Bits()
	prefix := netip.PrefixFrom(reg.Addr, bits)
	b, err := json.MarshalIndent(GuestReceipt{
		Id:     reg.Id,
		Prefix: prefix,
	}, "", "  ")
	if err != nil {
		return err
	}
	if _, err = w.Write(b); err != nil {
		return err
	}
	if true {
		xlog.Info.Printf("guest %s %v @ %v, receipt%s",
			cn, reg.Id, prefix, b)
	}
	return nil
}

func (vpn *regVpn) deny(req *http.Request) error {
	sub, err := reqsub(req)
	if err != nil {
		return NewStatusError(http.StatusInternalServerError, err)
	}
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	for i, c := range vpn.pending {
		if c.Subject.CommonName == sub {
			vpn.pending = slices.Delete(vpn.pending, i, i+1)
			return nil
		}
	}
	return NewStatusError(http.StatusNotFound, sub)
}

func (vpn *regVpn) dumpSubscribers(w io.Writer) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	blk := pem.Block{
		Type: "CERTIFICATE",
	}
	for _, c := range vpn.subscribers {
		blk.Bytes = c.Raw
		if err := pem.Encode(w, &blk); err != nil {
			return err
		}
	}
	return nil
}

func (vpn *regVpn) lease(name string) (netip.Addr, error) {
	addr := vpn.addr.top.Next()
	if !addr.IsValid() {
		return addr, xerrors.Invalid("top")
	}
	if !ConfigByName[vpn.name].Prefix.Contains(addr) {
		return addr, xerrors.Range("lease")
	}
	vpn.addr.named[name] = addr
	vpn.addr.name[addr] = name
	vpn.addr.top = addr
	return addr, nil
}

func (vpn *regVpn) selfOrAdmin(cn string) error {
	if vpn.crt.Subject.CommonName == cn {
		return nil
	}
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	if t, found := vpn.admin[cn]; found && t {
		return nil
	}
	return NewStatusError(http.StatusForbidden, cn)
}

func (vpn *regVpn) selfOrSubscriber(peer *x509.Certificate) error {
	if peer == nil {
		return NewStatusError(http.StatusForbidden,
			"no client certificate")
	}
	if peer.Equal(vpn.crt) {
		return nil
	}

	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	for _, c := range vpn.subscribers {
		if peer.Equal(c) {
			return nil
		}
	}
	return NewStatusError(http.StatusForbidden, peer.Subject.CommonName)
}

func (vpn *regVpn) showActive(w http.ResponseWriter) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	for name, r := range vpn.named {
		fmt.Fprint(w, name, " (", r.Id)
		if r.Addr.IsValid() {
			fmt.Fprintf(w, ", %v", r.Addr)
		}
		fmt.Fprint(w, ")")
		if len(r.Via) > 0 {
			fmt.Fprintf(w, " via %v", r.Via)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func (vpn *regVpn) showAddress(w http.ResponseWriter, qv url.Values) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	if qv.Has("arg0") {
		name := qv.Get("arg0")
		addr, ok := vpn.addr.named[name]
		if !ok {
			return xerrors.Unknown(name)
		}
		fmt.Fprintln(w, addr)
	} else {
		names := make([]string, 0, len(vpn.addr.named))
		for name := range vpn.addr.named {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(w, "%s: %v\n", name, vpn.addr.named[name])
		}
	}
	return nil
}

func (vpn *regVpn) showHosts(w http.ResponseWriter, qv url.Values) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	if qv.Has("arg0") {
		addr, err := netip.ParseAddr(qv.Get("arg0"))
		if err != nil {
			return err
		}
		name, ok := vpn.addr.name[addr]
		if !ok {
			return xerrors.Unknown(addr.String())
		}
		fmt.Fprintln(w, name)
	} else {
		addrs := make([]netip.Addr, 0, len(vpn.addr.name))
		for addr := range vpn.addr.name {
			addrs = append(addrs, addr)
		}
		slices.SortFunc(addrs, func(a, b netip.Addr) int {
			return a.Compare(b)
		})
		for _, addr := range addrs {
			fmt.Fprintf(w, "%v\t%s\n", addr, vpn.addr.name[addr])
		}
	}
	return nil
}

func (vpn *regVpn) showPending(w io.Writer) error {
	t, err := CertificatesTemplate()
	if err != nil {
		return err
	}
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	return t.Execute(w, vpn.pending)
}

func (vpn *regVpn) showSubscriber(w http.ResponseWriter, qv url.Values) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	if !qv.Has("arg0") {
		names := xmaps.Keys(vpn.subscriberNamed)
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintln(w, name)
		}
		return nil
	}

	arg0 := qv.Get("arg0")

	sub, ok := vpn.subscriberNamed[arg0]
	if !ok {
		return xerrors.NotFound(arg0)
	}

	t, err := CertificatesTemplate()
	if err != nil {
		return err
	}

	return t.Execute(w, []*x509.Certificate{sub})
}

func (vpn *regVpn) showTenant(w http.ResponseWriter, qv url.Values) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	if !qv.Has("arg0") {
		return xerrors.Incomplete("address")
	}
	addr, err := netip.ParseAddr(qv.Get("arg0"))
	if err != nil {
		return xerrors.Label(err, "address")
	}
	name, ok := vpn.addr.name[addr]
	if !ok {
		return xerrors.NotFound(addr.String())
	}
	fmt.Fprintln(w, name)
	return nil
}

// This has an empty response.  The client will retrieve the server cert
// through its TLS negotiation; then prompt the user to ise as root certificate
// authority.
func (vpn *regVpn) subscribe(req *http.Request) error {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	c := req.TLS.PeerCertificates[0]
	cn := c.Subject.CommonName
	if _, present := vpn.subscriberNamed[cn]; present {
		return xerrors.Unavailable(cn)
	}
	for _, p := range vpn.pending {
		if p.Subject.CommonName == cn {
			return xerrors.Unavailable(cn)
		}
	}
	vpn.pending = append(vpn.pending, c)
	return nil
}

func (vpn *regVpn) StateDir() string {
	dir := vpnStateDir
	if len(vpn.name) > 0 && vpn.name != "vpn" {
		dir = filepath.Join(dir, vpn.name)
	}
	return dir
}

func (vpn *regVpn) unsubscribe(req *http.Request) error {
	sub, err := reqsub(req)
	if err != nil {
		return err
	}
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	delete(vpn.admin, sub)
	return removeCertificate(vpn.StateDir(), sub, vpn.subscribers)
}

func (vpn *regVpn) whois(req *http.Request) ([]byte, error) {
	var k, s string
	var reg *Registration
	var ok bool

	qv := req.URL.Query()

	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	switch {
	case qv.Has(RestWhoisAddress):
		k = RestWhoisAddress
		s = qv.Get(RestWhoisAddress)
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return nil, err
		}
		reg, ok = vpn.addressed[addr]
	case qv.Has(RestWhoisId):
		k = RestWhoisId
		s = qv.Get(RestWhoisId)
		id, err := ParseId(s)
		if err != nil {
			return nil, err
		}
		reg, ok = vpn.indexed[id.Index()]
	case qv.Has(RestWhoisName):
		k = RestWhoisName
		s = qv.Get(RestWhoisName)
		reg, ok = vpn.named[s]
	default:
		return nil, NewStatusError(http.StatusBadRequest,
			"no <address>, <id>, or <name>: ", req.URL)
	}
	if !ok || reg == nil {
		return nil, NewStatusError(http.StatusNotFound, k+" "+s)
	}
	return json.MarshalIndent(reg, "", "  ")
}
