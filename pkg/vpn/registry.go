// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"gopkg.in/yaml.v3"
)

type registry struct {
	mutex sync.RWMutex
	http  *http.Server
	vpn   map[string]*regVpn
	wg    sync.WaitGroup
}

type regVpn struct {
	mutex sync.RWMutex
	name  string

	prefix netip.Prefix

	addr struct {
		name  map[netip.Addr]string
		named map[string]netip.Addr
		top   netip.Addr
	}

	admin       map[string]bool
	subscribers *Certificates
	pending     pending

	block struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[int]*pem.Block
		named      map[string]*pem.Block
	}

	idbook IdBook

	exchange IdRing
}

// [Registry] unmarshals the [Flags.FN.Cfg] YAML file into ConfigByName.
var ConfigByName map[string]*struct {
	// Required network prefix.
	Prefix netip.Prefix
	// Optional static VPN address assignment. All others are dynamically
	// assigned starting at the next available address from the highest
	// entry. (default: Prefix.Addr)
	Address map[string]netip.Addr
	// Optional list of subscriber's that are enabled to
	// approve/deny/unscribe others.
	Admins []string
	// Optional directory or ``.pem'' extensioned file containing
	// certificates of approved subscribers.
	// (default: StateHome/VPN-subscribers/)
	Subscribers string
}

// Registry is a web server providing a REST interface to persistent
// files and ephemeral tables.
func Registry(ctx context.Context, args []string) error {
	var wg sync.WaitGroup
	var reg registry

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
A RESTful WWW server.

{{flags .}}`)

	Flags.FN.Cfg = filepath.Join(ConfigHome(), DefaultCfg)
	Flags.FN.Crt = filepath.Join(ConfigHome(), DefaultCrt)
	Flags.FN.Key = filepath.Join(ConfigHome(), DefaultKey)
	port := flag.Uint("p", 8003, "Port number.")

	err := AddAndParseFlags(ctx, args)
	if err != nil {
		return err
	} else {
		args = flag.Args()
	}

	if _, err = os.Stat(Flags.FN.Cfg); err != nil {
		return err
	}
	if _, err = os.Stat(Flags.FN.Crt); err != nil {
		return err
	}
	if _, err = os.Stat(Flags.FN.Key); err != nil {
		return err
	}

	reg.vpn = make(map[string]*regVpn)

	if err = reg.reload(); err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	svc := fmt.Sprint(":", *port)

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

	verbose.Println("start", svc)
	defer verbose.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer verbose.Println("stopping", svc, "...")

	wg.Add(1)
	go reg.shutdown(cctx, &wg)

	err = reg.http.ListenAndServeTLS(Flags.FN.Crt, Flags.FN.Key)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

func (reg *registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if !req.TLS.HandshakeComplete {
		verbose.Println("incomplete handshake")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if len(req.TLS.PeerCertificates) == 0 {
		verbose.Println("no certificates")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	peer0 := req.TLS.PeerCertificates[0]
	cn := peer0.Subject.CommonName
	qv := req.URL.Query()
	op := qv.Get("op")
	name := strings.TrimLeft(req.URL.Path, "/")
	if len(name) == 0 {
		name = "vpn"
	}
	vpn, found := reg.vpnNamed(name)
	if !found || vpn == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, name)
		return
	}
	var err error
	defer func() {
		if err != nil {
			verbose.Print(op, ": ", err, "\n")
		}
	}()
	switch op {
	case "approve":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if err = vpn.selfOrAdmin(cn); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err = vpn.approve(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "certify":
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			// empty response so that client may retrieve this
			// certificate from TLS negotiation.
			w.WriteHeader(http.StatusOK)
		}
	case "checkin":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err = vpn.checkin(w, req); err != nil {
			if errors.Is(err, xerrors.ErrUnavailable) {
				w.WriteHeader(http.StatusTooEarly)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
			}
		}
	case "deny":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if err = vpn.selfOrAdmin(cn); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err = vpn.deny(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "dump-subscribers":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err = vpn.dumpSubscribers(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "ping":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "reload":
		if name != "vpn" {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprint(w, name, ": unacceptable")
		} else if err := vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := reg.reload(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "show-active":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.showActive(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "show-admins":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			for _, s := range ConfigByName[name].Admins {
				fmt.Println(w, s)
			}
		}
	case "show-pending":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			vpn.showPending(w)
		}
	case "show-subscribers":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.subscribers.Show(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "subscribe":
		if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.subscribe(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "unsubscribe":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if err = vpn.selfOrAdmin(cn); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.unsubscribe(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "whois":
		if err = vpn.selfOrSubscriber(peer0); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, cn)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err = vpn.whois(w, req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (reg *registry) shutdown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	cctx, cancel := context.
		WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	verbose.Print("shutdown ", reg.http.Addr, "...")
	reg.http.Shutdown(cctx)
}

func (reg *registry) reload() error {
	data, err := os.ReadFile(Flags.FN.Cfg)
	if err != nil {
		return err
	}
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	if err = yaml.Unmarshal(data, &ConfigByName); err != nil {
		return err
	}
	for name, cfg := range ConfigByName {
		vpn, found := reg.vpn[name]
		if !found {
			vpn = &regVpn{
				name: name,
			}

			vpn.addr.name = make(map[netip.Addr]string)
			vpn.addr.named = make(map[string]netip.Addr)

			vpn.block.addressed = make(map[netip.Addr]*pem.Block)
			vpn.block.identified = make(map[int]*pem.Block)
			vpn.block.named = make(map[string]*pem.Block)

			reg.vpn[name] = vpn
		}

		if !cfg.Prefix.IsValid() {
			return xerrors.Invalid(name, "prefix")
		}

		// FIXME how to accept changed prefix?
		vpn.prefix = cfg.Prefix

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

		dfn := cfg.Subscribers
		if len(dfn) == 0 {
			dfn = filepath.Join(xdg.StateHome(),
				fmt.Sprint(name, "-subscribers.pem"))
		}
		vpn.subscribers, err = NewCertificates(dfn)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		verbose.Println("subscribers...")
		for _, bc := range vpn.subscribers.BCs {
			if bc.Cert != nil {
				verbose.Println(bc.Cert.Subject.CommonName)
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
	return qv.Get("subscriber"), nil
}

func (vpn *regVpn) approve(req *http.Request) error {
	sub, err := reqsub(req)
	if err != nil {
		return err
	}
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	blk, cert, err := vpn.pending.pull(sub)
	if err != nil {
		return err
	}
	return vpn.subscribers.Add(blk, cert)
}

func (vpn *regVpn) checkin(w http.ResponseWriter, req *http.Request) error {
	var (
		id,
		via box.Id
		addr netip.Addr
		ap   netip.AddrPort
		err  error
		ok   bool
		svc  string
	)

	defer req.Body.Close()
	qv := req.URL.Query()

	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	cn := req.TLS.PeerCertificates[0].Subject.CommonName

	if qv.Has("service") {
		svc = qv.Get("service")
		if ap, err = netip.ParseAddrPort(svc); err != nil {
			return xerrors.Label(err, "service")
		} else {
			verbose.Printf("new service %s @ %v", cn, ap)
		}
	} else if via, err = vpn.exchange.Next(); err != nil {
		return err
	} else {
		verbose.Printf("new quest %s via %d", cn, IdIndex(via))
	}

	data, err := xerrors.MarkResult(io.ReadAll(req.Body))
	if err != nil {
		return err
	}

	blk, _ := pem.Decode(data)
	if blk == nil {
		return xerrors.Invalid("encoding")
	} else if blk.Type != "PUBLIC KEY" {
		return xerrors.Invalid("block_type", blk.Type)
	}
	if _, ok := blk.Headers["nonce"]; !ok {
		return xerrors.Incomplete("nonce")
	}

	entry, exists := vpn.block.named[cn]
	if exists {
		entry.Headers["nonce"] = blk.Headers["nonce"]
		entry.Bytes = blk.Bytes
		if addr, err = addressHeader(entry); err != nil {
			return xerrors.Label(err, "existing_address_header")
		}
		entry.Headers["address"] = addr.String()
		if id, err = idHeader(entry); err != nil {
			return xerrors.Label(err, "existing_id_header")
		}
		id = BumpIdVersion(id)
		entry.Headers["id"] = fmt.Sprint(id)
		if len(svc) > 0 {
			err = vpn.exchange.Update(id)
			if err != nil {
				return xerrors.Label(err, "exchange_id")
			}
			entry.Headers["service"] = svc
		} else {
			via, err = vpn.exchange.Next()
			if err != nil {
				return xerrors.Label(err, "next_exchange")
			}
			entry.Headers["via"] = fmt.Sprint(via)
		}
	} else {
		blk.Headers["name"] = cn

		if addr, ok = vpn.addr.named[cn]; !ok {
			if addr, err = vpn.lease(cn); err != nil {
				return err
			}
		}
		blk.Headers["address"] = addr.String()

		id = vpn.idbook.New()
		blk.Headers["id"] = fmt.Sprint(id)
		if len(svc) > 0 {
			vpn.exchange.Append(id)
			blk.Headers["service"] = svc
		} else if via, err = vpn.exchange.Next(); err != nil {
			return xerrors.Label(err, "exchange_id")
		} else {
			blk.Headers["via"] = fmt.Sprint(via)
		}

		vpn.block.named[cn] = blk
		vpn.block.addressed[addr] = blk
		vpn.block.identified[IdIndex(id)] = blk
	}

	fmt.Fprintln(w, "id:", id)
	fmt.Fprintln(w, "address:", addr)
	fmt.Fprintln(w, "prefix:", ConfigByName[vpn.name].Prefix)
	if len(svc) == 0 {
		fmt.Fprintln(w, "via:", via)
		verbose.Printf("%s assigned %d @ %v via %v\n",
			cn, id, addr, via)
	} else {
		verbose.Printf("%s assigned %d @ %v\n", cn, id, addr)
	}
	return nil
}

func (vpn *regVpn) deny(req *http.Request) error {
	sub, err := reqsub(req)
	if err == nil {
		vpn.mutex.Lock()
		defer vpn.mutex.Unlock()
		_, _, err = vpn.pending.pull(sub)
	}
	return err
}

func (vpn *regVpn) dumpSubscribers(w io.Writer) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	return vpn.subscribers.Dump(w)
}

func (vpn *regVpn) lease(name string) (netip.Addr, error) {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	addr := vpn.addr.top.Next()
	if !addr.IsValid() {
		return addr, xerrors.Invalid("top")
	}
	if !vpn.prefix.Contains(addr) {
		return addr, xerrors.Range("lease")
	}
	vpn.addr.named[name] = addr
	vpn.addr.name[addr] = name
	vpn.addr.top = addr
	return addr, nil
}

func (vpn *regVpn) selfOrAdmin(cn string) error {
	if me, err := Crt(); err != nil {
		return err
	} else if me.First().Subject.CommonName == cn {
		return nil
	}
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	if t, found := vpn.admin[cn]; found && t {
		return nil
	}
	return xerrors.NotFound("admin", cn)
}

func (vpn *regVpn) selfOrSubscriber(peer *x509.Certificate) error {
	if peer == nil {
		return xerrors.Invalid("peer")
	}
	if me, err := Crt(); err != nil {
		return err
	} else if me.Has(peer) {
		return nil
	}

	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	if vpn.subscribers.Has(peer) {
		return nil
	}
	return xerrors.NotFound(peer.Subject.CommonName)
}

func (vpn *regVpn) showActive(w http.ResponseWriter) error {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	for name, blk := range vpn.block.named {
		id, err := idHeader(blk)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "%s (%d, %d", name, IdIndex(id), IdVersion(id))
		addr, err := addressHeader(blk)
		if err == nil {
			fmt.Fprintf(w, ", %v", addr)
		}
		fmt.Fprint(w, ")")
		via, err := viaHeader(blk)
		if err == nil {
			fmt.Fprintf(w, " via %d", IdIndex(via))
		}
		svc, err := serviceHeader(blk)
		if err == nil {
			fmt.Fprintf(w, " service %v", svc)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func (vpn *regVpn) showPending(w io.Writer) {
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	fmt.Fprint(w, &vpn.pending)
}

// This has an empty response.  The client will retrieve the server cert
// through its TLS negotiation; then prompt the user to ise as root certificate
// authority.
func (vpn *regVpn) subscribe(req *http.Request) error {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	blk, _ := pem.Decode(data)
	cert := req.TLS.PeerCertificates[0]
	cn := cert.Subject.CommonName
	if _, present := vpn.subscribers.Named[cn]; present {
		return xerrors.Unavailable("n")
	}
	vpn.pending.add(blk, cert)
	return nil
}

func (vpn *regVpn) unsubscribe(req *http.Request) error {
	sub, err := reqsub(req)
	if err != nil {
		return err
	}
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()
	delete(vpn.admin, sub)
	return vpn.subscribers.Remove(sub)
}

func (vpn *regVpn) whois(w http.ResponseWriter, req *http.Request) error {
	var (
		k   string
		blk *pem.Block
		ok  bool
	)

	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()

	qv := req.URL.Query()
	if qv.Has("name") {
		k = qv.Get("name")
		blk, ok = vpn.block.named[k]
	} else if qv.Has("id") {
		id, err := xerrors.MarkResult(ParseId(qv.Get("id")))
		if err != nil {
			return xerrors.Label(err, "id")
		}
		k = fmt.Sprint(id)
		blk, ok = vpn.block.identified[IdIndex(id)]
	} else if qv.Has("address") {
		addr, err := xerrors.MarkResult(netip.
			ParseAddr(qv.Get("address")))
		if err != nil {
			return xerrors.Label(err, "address")
		}
		k = fmt.Sprint(addr)
		blk, ok = vpn.block.addressed[addr]
	} else {
		return xerrors.Incomplete("no <name>, <address> or <id>")
	}
	if blk == nil || !ok {
		return xerrors.NotFound(k)
	}
	return pem.Encode(w, blk)
}
