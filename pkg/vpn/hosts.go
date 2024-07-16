// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bufio"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

type hostsFile struct {
	sync.RWMutex
	path   string
	mode   os.FileMode
	prefix netip.Prefix
	addr   map[string]netip.Addr
	name   map[netip.Addr]string
	top    netip.Addr
}

func newHostsFile(vpn string, prefix netip.Prefix) (*hostsFile, error) {
	vpnhosts := filepath.Join(vpn, HostsFileName)
	h := &hostsFile{
		path:   filepath.Join(xos.StateHome(), vpnhosts),
		prefix: prefix,
		addr:   make(map[string]netip.Addr),
		name:   make(map[netip.Addr]string),
	}
	r, err := os.Open(h.path)
	if err != nil {
		config := filepath.Join(xos.ConfigHome(), vpnhosts)
		if !os.IsNotExist(err) {
			return h, err
		} else if r, err = os.Open(config); err != nil {
			return h, err
		}
	}
	defer r.Close()
	for sc := bufio.NewScanner(r); sc.Scan(); {
		s := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(s, "#") {
			continue
		}
		f := strings.Fields(s)
		if len(f) < 2 {
			continue
		}
		addr, e := netip.ParseAddr(f[0])
		if e != nil {
			return h, fmt.Errorf("%s: %w", h.path, e)
		}
		if !h.prefix.Contains(addr) {
			return h, xerrors.
				Range(h.path, addr.String(), h.prefix.String())
		}
		if !h.top.IsValid() || h.top.Less(addr) {
			h.top = addr
		}
		h.addr[f[1]] = addr
		h.name[addr] = f[1]
		for _, aka := range f[2:] {
			h.addr[aka] = addr
			verbose.Println(addr, aka)
		}
	}
	return h, nil
}

func (h *hostsFile) addressed(addr netip.Addr) (string, error) {
	h.RLock()
	defer h.RUnlock()
	name, ok := h.name[addr]
	if !ok {
		return name, xerrors.NotFound(h.path, addr.String())
	}
	return name, nil
}

func (h *hostsFile) lease(name string) (netip.Addr, error) {
	h.Lock()
	defer h.Unlock()
	addr := h.top.Next()
	if !addr.IsValid() {
		return addr, xerrors.ErrInvalid
	}
	h.addr[name] = addr
	h.name[addr] = name
	h.top = addr
	const fflags = os.O_RDWR | os.O_CREATE | os.O_APPEND
	w, err := os.OpenFile(h.path, fflags, h.mode)
	if err != nil {
		return addr, err
	}
	defer w.Close()
	_, err = fmt.Fprintln(w, addr, name)
	return addr, err
}

func (h *hostsFile) named(hn string) (netip.Addr, error) {
	h.RLock()
	defer h.RUnlock()
	var err error
	addr, ok := h.addr[hn]
	if !ok {
		err = xerrors.NotFound(h.path, hn)
	}
	return addr, err
}
