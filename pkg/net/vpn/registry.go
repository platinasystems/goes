// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509certs"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

//go:embed registry.txt
var RegistryHelp string

// Registry is a [goes] daemon that provides a [RESTful] API to a VPN's
// ephemeral cipher key database and its persistent authentication
// certificates.
// Each VPN has one and only one active Registry that should use replicated
// storage volumes for prerequisite configuration and dynamic state files.
//
// Usage: goes start vpn registry [flags]
//
// Flags:
//
//	-q	Silence most logs.
//
//	-service <addr>:<port> (default 0.0.0.0:8003)
//		If <addr> is 0.0.0.0 or [::], this will lookup the first ipv4
//		or ipv6 address of certificate's primary DNS name.
//
//	-v	Log everything.
//
// Prerequisite Configuration Files:
//
//   - [KeyFileName]
//   - [CrtFileName]
//   - [PrefixFileName]
//   - [HostFileName]
//
// [RESTful]: https://en.wikipedia.org/wiki/REST
func Registry(ctx context.Context, args []string) error {
	var wg sync.WaitGroup
	var reg registry

	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		fmt.Print(RegistryHelp)
		return nil
	}

	svc, err := DaemonFlags(ctx, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Print(RegistryHelp)
		}
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	verbose.Println("start", svc)
	defer verbose.Println("stopped", svc)
	defer wg.Wait()
	defer cancel()
	defer verbose.Println("stopping", svc, "...")

	c, err := vpnCrtFile()
	if err != nil {
		return err
	}
	k, err := vpnKeyFile()
	if err != nil {
		return err
	}

	// reg.subjectCommonName = first.Subject.CommonName

	reg.vpn = make(map[string]*regVpn)

	reg.http = &http.Server{
		Addr:    svc.String(),
		Handler: &reg,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAnyClientCert,
		},
		BaseContext: func(net.Listener) context.Context {
			return cctx
		},
	}

	err = filepath.WalkDir(program.ConfigHome(),
		func(path string, de fs.DirEntry, err error) error {
			return reg.walker(ctx, path, de, err)
		})
	if err != nil {
		return err
	}

	wg.Add(1)
	go reg.shutdown(cctx, &wg)

	verbose.Print("start ", reg.http.Addr)
	defer verbose.Print("stopped ", reg.http.Addr)

	err = reg.http.ListenAndServeTLS(c.Path, k.Path)
	cancel()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

type registry struct {
	mutex sync.RWMutex
	http  *http.Server
	vpn   map[string]*regVpn
	wg    sync.WaitGroup
}

type regVpn struct {
	mutex sync.RWMutex
	dir,
	name string
	prefix netip.Prefix
	hosts  *hostsFile
	admins,
	mirrors,
	subscribers *x509certs.File
	pending pending

	block struct {
		addressed  map[netip.Addr]*pem.Block
		identified map[int]*pem.Block
		named      map[string]*pem.Block
	}

	idbook IdBook

	exchange IdRing
}

func (reg *registry) vpnNamed(name string) *regVpn {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return reg.vpn[name]
}

func (reg *registry) add(vpn *regVpn) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.vpn[vpn.name] = vpn
}

func (reg *registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if !req.TLS.HandshakeComplete || len(req.TLS.PeerCertificates) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	peer0 := req.TLS.PeerCertificates[0]
	qv := req.URL.Query()
	op := qv.Get("op")
	path := strings.TrimLeft(req.URL.Path, "/")
	vpn := reg.vpnNamed(path)
	if vpn == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, path)
		return
	}
	switch op {
	case "approve":
		if !vpn.isAuthorized(peer0, vpn.admins) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.adminApprove(req); err != nil {
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
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.checkin(w, req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "deny":
		if !vpn.isAuthorized(peer0, vpn.admins) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.adminDeny(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "disable":
		if !vpn.isAuthorized(peer0, vpn.admins) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.adminDisable(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "dump-mirrors":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.mirrors.Dump(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "dump-subscribers":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.subscribers.Dump(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "enable":
		if !vpn.isAuthorized(peer0, vpn.admins) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.adminEnable(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "ping":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "rescan":
		if path != "" {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Fprint(w, path, ": unacceptable")
		} else if !vpn.isAuthorized(peer0) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := filepath.WalkDir(program.ConfigHome(),
			func(path string, de fs.DirEntry, err error) error {
				ctx := req.Context()
				return reg.walker(ctx, path, de, err)
			}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "show-active":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.showActive(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	case "show-admins":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprint(w, vpn.admins)
		}
	case "show-mirrors":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprint(w, vpn.mirrors)
		}
	case "show-pending":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprint(w, &vpn.pending)
		}
	case "show-subscribers":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else {
			fmt.Fprint(w, vpn.subscribers)
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
		if !vpn.isAuthorized(peer0, vpn.admins) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.adminUnsubscribe(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		} else {
			fmt.Fprintln(w, "OK")
		}
	case "whois":
		if !vpn.isAuthorized(peer0, vpn.subscribers) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, peer0.Subject.CommonName)
		} else if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, req.Method)
		} else if err := vpn.whois(w, req); err != nil {
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

func (reg *registry) walker(
	ctx context.Context, path string, de fs.DirEntry, err error,
) error {
	if err != nil {
		verbose.Println("skipping", path, "b/c", err)
		return fs.SkipDir
	}
	if de.IsDir() && strings.HasPrefix(de.Name(), ".") {
		verbose.Println("skipping", path)
		return fs.SkipDir
	}
	if de.Name() != vpnPrefixFileName {
		return nil
	}
	if err = ctx.Err(); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	name := vpnName(dir)
	if reg.vpnNamed(name) != nil {
		verbose.Println(name, "exists")
		return nil
	}

	vpn := &regVpn{
		dir:  filepath.Join(program.StateHome(), name),
		name: name,
	}

	vpn.prefix, err = egress.MarkResult(vpnPrefix(path))
	if err != nil {
		return err
	}

	vpn.block.addressed = make(map[netip.Addr]*pem.Block)
	vpn.block.identified = make(map[int]*pem.Block)
	vpn.block.named = make(map[string]*pem.Block)

	afn := filepath.Join(name, vpnAdminsFileName)
	vpn.admins, err = egress.MarkResult(vpnCertsFile(afn))
	if err != nil {
		return err
	}

	mfn := filepath.Join(name, vpnMirrorsFileName)
	vpn.mirrors, err = egress.MarkResult(vpnCertsFile(mfn))
	if err != nil {
		return err
	}

	sfn := filepath.Join(name, vpnSubscribersFileName)
	vpn.subscribers, err = egress.MarkResult(vpnCertsFile(sfn))
	if err != nil {
		return err
	}

	vpn.hosts, err = egress.MarkResult(newHostsFile(name, vpn.prefix))
	if err != nil {
		return err
	}

	reg.add(vpn)

	return nil
}

func (*regVpn) reqsub(req *http.Request) (string, error) {
	qv := req.URL.Query()
	if !qv.Has("subscriber") {
		return "", ErrUnspecifiedSub
	}
	return qv.Get("subscriber"), nil
}

func (vpn *regVpn) lookup(cn string) (netip.Addr, error) {
	addr, err := vpn.hosts.named(cn)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			addr, err = vpn.hosts.lease(cn)
		} else {
			return addr, err
		}
	}
	if !addr.IsValid() {
		err = fmt.Errorf("%s[%s]: %w",
			vpn.hosts.path, cn, ErrInvalid)
	}
	return addr, err
}

func (vpn *regVpn) adminApprove(req *http.Request) error {
	sub, err := vpn.reqsub(req)
	if err != nil {
		return err
	}
	blk, cert, err := vpn.pending.pull(sub)
	if err != nil {
		return err
	}
	return vpn.subscribers.Add(blk, cert)
}

func (vpn *regVpn) adminDeny(req *http.Request) error {
	sub, err := vpn.reqsub(req)
	if err == nil {
		_, _, err = vpn.pending.pull(sub)
	}
	return err
}

func (vpn *regVpn) adminDisable(req *http.Request) error {
	sub, err := vpn.reqsub(req)
	if err != nil {
		return err
	}
	return vpn.admins.Remove(sub)
}

func (vpn *regVpn) adminEnable(req *http.Request) error {
	sub, err := vpn.reqsub(req)
	if err != nil {
		return err
	}
	blk, cert, err := vpn.subscribers.Named(sub)
	if err != nil {
		return err
	}
	return vpn.admins.Add(blk, cert)
}

func (vpn *regVpn) adminUnsubscribe(req *http.Request) error {
	sub, err := vpn.reqsub(req)
	if err != nil {
		return err
	}
	return vpn.admins.Remove(sub)
}

func (vpn *regVpn) checkin(w http.ResponseWriter, req *http.Request) error {
	var id, via box.Id
	var addr netip.Addr
	var svc string

	defer req.Body.Close()
	qv := req.URL.Query()

	name := req.TLS.PeerCertificates[0].Subject.CommonName

	if qv.Has("service") {
		svc = qv.Get("service")
		if ap, err := netip.ParseAddrPort(svc); err != nil {
			return fmt.Errorf("service: %w", err)
		} else {
			verbose.Printf("new service %s @ %v", name, ap)
		}
	} else {
		const period = 5 * time.Second
		for try := 1; true; try++ {
			var err error
			via, err = vpn.exchange.Next()
			if err == nil {
				verbose.Printf("new quest %s via %d",
					name, IdIndex(via))
				break
			}
			if try == 3 {
				return fmt.Errorf("exchange %w", err)
			}
			verbose.Printf("retry %s exchange assignment in %v",
				name, period)
			time.Sleep(period)
		}
	}

	data, err := egress.MarkResult(io.ReadAll(req.Body))
	if err != nil {
		return err
	}

	blk, _ := pem.Decode(data)
	if blk == nil {
		return ErrNotPEM
	} else if blk.Type != "PUBLIC KEY" {
		return ErrInvalidBlockType
	}
	if _, ok := blk.Headers["nonce"]; !ok {
		return ErrNoNonce
	}

	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	entry, exists := vpn.block.named[name]
	if exists {
		entry.Headers["nonce"] = blk.Headers["nonce"]
		entry.Bytes = blk.Bytes
		if addr, err = addressHeader(entry); err != nil {
			return fmt.Errorf("existing address header: %w", err)
		}
		entry.Headers["address"] = addr.String()
		if id, err = idHeader(entry); err != nil {
			return fmt.Errorf("existing id header: %w", err)
		}
		id = BumpIdVersion(id)
		entry.Headers["id"] = fmt.Sprint(id)
		if len(svc) > 0 {
			err = vpn.exchange.Update(id)
			if err != nil {
				return fmt.Errorf("exchange id:", err)
			}
			entry.Headers["service"] = svc
		} else {
			via, err = vpn.exchange.Next()
			if err != nil {
				return fmt.Errorf("exchange id:", err)
			}
			entry.Headers["via"] = fmt.Sprint(via)
		}
	} else {
		blk.Headers["name"] = name
		addr, err = vpn.lookup(name)
		if err != nil {
			return fmt.Errorf("address: %w", err)
		}
		blk.Headers["address"] = addr.String()
		id = vpn.idbook.New()
		blk.Headers["id"] = fmt.Sprint(id)
		if len(svc) > 0 {
			vpn.exchange.Append(id)
			blk.Headers["service"] = svc
		} else {
			via, err = vpn.exchange.Next()
			if err != nil {
				return fmt.Errorf("exchange id:", err)
			}
			blk.Headers["via"] = fmt.Sprint(via)
		}

		vpn.block.named[name] = blk
		vpn.block.addressed[addr] = blk
		vpn.block.identified[IdIndex(id)] = blk
	}

	fmt.Fprintln(w, "id:", id)
	fmt.Fprintln(w, "address:", addr)
	fmt.Fprintln(w, "prefix:", vpn.prefix)
	if len(svc) == 0 {
		fmt.Fprintln(w, "via:", via)
	}
	return nil
}

func (vpn *regVpn) checkout(w http.ResponseWriter, req *http.Request) error {
	vpn.mutex.Lock()
	defer vpn.mutex.Unlock()

	name := req.TLS.PeerCertificates[0].Subject.CommonName
	blk, ok := vpn.block.named[name]
	if !ok {
		return egress.Mark(ErrNotFound)
	}
	delete(vpn.block.named, name)
	id, err := egress.MarkResult(idHeader(blk))
	if err != nil {
		return err
	}
	delete(vpn.block.identified, IdIndex(id))
	addr, err := egress.MarkResult(addressHeader(blk))
	if err != nil {
		return err
	}
	delete(vpn.block.addressed, addr)
	if _, ok := blk.Headers["service"]; ok {
		vpn.exchange.Remove(id)
	}
	vpn.idbook.Put(id)
	return nil
}

func (*regVpn) isAuthorized(peer *x509.Certificate, permitted ...interface {
	Has(*x509.Certificate) bool
}) bool {
	if c, err := vpnCrtFile(); err != nil {
		return false
	} else if c.Has(peer) {
		return true
	}
	for _, v := range permitted {
		if v.Has(peer) {
			return true
		}
	}
	return false
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

// This has an empty response.  The client will retrieve the server cert
// through its TLS negotiation; then prompt the user to ise as root certificate
// authority.
func (vpn *regVpn) subscribe(req *http.Request) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	blk, _ := pem.Decode(data)
	cert := req.TLS.PeerCertificates[0]
	cn := cert.Subject.CommonName
	if _, _, err = vpn.subscribers.Named(cn); err == nil {
		return fmt.Errorf("%s: %w", cn, ErrSubscribed)
	}
	vpn.pending.add(blk, cert)
	return nil
}

func (vpn *regVpn) whois(w http.ResponseWriter, req *http.Request) error {
	var (
		blk *pem.Block
		ok  bool
	)
	qv := req.URL.Query()
	vpn.mutex.RLock()
	defer vpn.mutex.RUnlock()
	if qv.Has("name") {
		blk, ok = vpn.block.named[qv.Get("name")]
	} else if qv.Has("id") {
		id, err := egress.MarkResult(ParseId(qv.Get("id")))
		if err != nil {
			return err
		}
		blk, ok = vpn.block.identified[IdIndex(id)]
	} else if qv.Has("address") {
		addr, err := egress.MarkResult(netip.
			ParseAddr(qv.Get("address")))
		if err != nil {
			return err
		}
		blk, ok = vpn.block.addressed[addr]
	} else {
		return errors.New("no <name>, <address> or <id>")
	}
	if blk == nil || !ok {
		return ErrNotFound
	}
	return pem.Encode(w, blk)
}
