// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"sort"
	"sync"
)

type cache struct {
	sync.Mutex

	builtins map[string]func(...string) error
	names    []string
	path     []string
}

func (g *Goes) Builtins() map[string]func(...string) error {
	if g.cache.builtins == nil || len(g.cache.builtins) == 0 {
		g.cache.Lock()
		defer g.cache.Unlock()
		g.cache.builtins = map[string]func(...string) error{
			"apropos":   g.apropos,
			"complete":  g.complete,
			"copyright": g.copyright,
			"help":      g.help,
			"license":   g.license,
			"man":       g.man,
			"patents":   g.patents,
			"usage":     g.usage,
			"version":   g.version,
		}
	}
	return g.cache.builtins
}

func (g *Goes) Names() []string {
	if len(g.cache.names) < len(g.ByName) {
		g.cache.Lock()
		defer g.cache.Unlock()
		if got, want := len(g.cache.names), len(g.ByName); got < want {
			if got == 0 {
				g.cache.names = make([]string, 0, want)
			} else {
				g.cache.names = g.cache.names[:0]
			}
			for k := range g.ByName {
				g.cache.names = append(g.cache.names, k)
			}
			sort.Strings(g.cache.names)
		}
	}
	return g.cache.names
}

// set Path of sub-goes. e.g. "ip address"
func (g *Goes) Path() []string {
	if g.parent != nil && len(g.cache.path) == 0 {
		g.cache.Lock()
		defer g.cache.Unlock()
		for p := g; p != nil; p = p.parent {
			g.cache.path = append([]string{p.String()},
				g.cache.path...)
		}
	}
	return g.cache.path
}
