// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package alias

import (
	"encoding/json"
	"errors"
	"io/fs"
	"io/ioutil"
	"sort"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var ErrNotFound = errors.New("not found")

var aliases = cache.New[map[string]string](func(
	p *map[string]string,
) error {
	*p = make(map[string]string)
	fn, err := filename.Aliases.ValErr()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(fn)
	if err == nil {
		err = json.Unmarshal(b, p)
	} else if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	return err
})

func Keys() (keys []string) {
	aliases.Ref(func(p *map[string]string) error {
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

func Load(aka string) (ski string, ok bool) {
	aliases.Ref(func(p *map[string]string) error {
		ski, ok = (*p)[aka]
		return nil
	})
	return
}

func Range(f func(aka, ski string) bool) {
	aliases.Ref(func(p *map[string]string) error {
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

func Store(aka, ski string) error {
	return aliases.Ref(func(p *map[string]string) error {
		fn, err := filename.Aliases.ValErr()
		if err == nil {
			(*p)[aka] = ski
			data, err := json.Marshal(*p)
			if err == nil {
				err = ioutil.WriteFile(fn, data, 0644)
			}
		}
		return err
	})
}
