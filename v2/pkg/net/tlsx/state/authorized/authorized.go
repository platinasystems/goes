// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package authorized

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var ErrUnauthorized = errors.New("unauthorized")

var authorized = cache.New[map[string]bool](func(
	p *map[string]bool,
) error {
	*p = make(map[string]bool)
	fn, err := filename.Authorized.Value()
	if err != nil {
		return err
	}
	r, err := os.Open(fn)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return err
	}
	defer r.Close()
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		ski := sc.Text()
		if _, err = hex.DecodeString(ski); err == nil {
			(*p)[ski] = true
		}
	}
	return nil
})

func Keys() (keys []string) {
	authorized.Ref(func(p *map[string]bool) error {
		var i int
		keys = make([]string, len(*p))
		for k := range *p {
			keys[i] = k
			i += 1
		}
		sort.Strings(keys)
		return nil
	})
	return
}

func Load(ski string) (value bool, ok bool) {
	authorized.Ref(func(p *map[string]bool) error {
		value, ok = (*p)[ski]
		return nil
	})
	return
}

func Preload(skis ...string) {
	authorized.Preload(func(p *map[string]bool) {
		*p = make(map[string]bool)
		for _, ski := range skis {
			(*p)[ski] = true
		}
	})
}

func Range(f func(ski string) bool) {
	authorized.Ref(func(p *map[string]bool) error {
		var keys []string
		for k, v := range *p {
			if v {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !f(k) {
				break
			}
		}
		return nil
	})
}

func Store(ski string, value bool) error {
	return authorized.Ref(func(p *map[string]bool) error {
		if value {
			(*p)[ski] = true
		} else if !(*p)[ski] {
			return fmt.Errorf("%s: %w", ski, ErrUnauthorized)
		} else {
			delete(*p, ski)
		}
		f, err := os.Create(filename.Authorized.String())
		if err != nil {
			return err
		}
		defer f.Close()
		for ski := range *p {
			fmt.Fprintln(f, ski)
		}
		return nil
	})
}
