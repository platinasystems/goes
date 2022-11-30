// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package address

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/netip"
	"sort"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

const UnixPrefix = "unix://"

var addresses = cache.New[map[string]string](func(
	p *map[string]string,
) error {
	*p = make(map[string]string)
	b, err := ioutil.ReadFile(filename.Addresses())
	if err == nil {
		err = json.Unmarshal(b, p)
	} else if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	return err
})

func Load(ski string) (addr net.Addr, ok bool) {
	addresses.Ref(func(p *map[string]string) error {
		var s string
		s, ok = (*p)[ski]
		if strings.HasPrefix(s, UnixPrefix) {
			addr = &net.UnixAddr{
				Name: strings.TrimPrefix(s, UnixPrefix),
				Net:  "unix",
			}
			return nil
		}
		if ap, err := netip.ParseAddrPort(s); err != nil {
			return err
		} else {
			addr = net.TCPAddrFromAddrPort(ap)
		}
		return nil
	})
	return
}

func Range(f func(ski, addr string) bool) {
	addresses.Ref(func(p *map[string]string) error {
		var i int
		keys := make([]string, len(*p))
		for k := range *p {
			keys[i] = k
			i += 1
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !f(k, (*p)[k]) {
				break
			}
		}
		return nil
	})
}

func Store(ski, addr string) error {
	if !strings.HasPrefix(addr, UnixPrefix) {
		if _, err := netip.ParseAddrPort(addr); err != nil {
			return err
		}
	}
	return addresses.Ref(func(p *map[string]string) error {
		(*p)[ski] = addr
		data, err := json.MarshalIndent(*p, "", "  ")
		if err == nil {
			err = ioutil.WriteFile(filename.Addresses(), data, 0644)
		}
		return err
	})
}

func Unaddressed(ski, name string) error {
	return fmt.Errorf("%s[%s]: not found in %s",
		ski, name, filename.Addresses())
}
