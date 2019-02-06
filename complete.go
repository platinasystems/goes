// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type completer interface {
	Complete(...string) []string
}

func (g *Goes) Complete(args ...string) (completions []string) {
	n := len(args)
	if n == 0 || len(args[0]) == 0 {
		completions = g.Names()
	} else if v, found := g.ByName[args[0]]; found {
		if method, found := v.(completer); found {
			completions = method.Complete(args[1:]...)
		} else {
			completions, _ = filepath.Glob(args[n-1] + "*")
		}
	} else if _, found := g.Builtins()[args[0]]; found {
		if n == 1 || len(args[n-1]) == 0 {
			completions = g.Names()
		} else {
			for _, name := range g.Names() {
				if strings.HasPrefix(name, args[n-1]) {
					completions = append(completions, name)
				}
			}
		}
	} else if n == 1 {
		for _, name := range g.Names() {
			if strings.HasPrefix(name, args[0]) {
				completions = append(completions, name)
			}
		}
		for builtin := range g.Builtins() {
			if strings.HasPrefix(builtin, args[0]) {
				completions = append(completions, builtin)
			}
		}
		if len(completions) > 0 {
			sort.Strings(completions)
		}
	}
	return
}

// This may be used for bash completion of goes commands like this.
//
//	_goes() {
//		if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
//			COMPREPLY=($(goes complete ${COMP_WORDS[@]:1} ''))
//		else
//			COMPREPLY=($(goes complete ${COMP_WORDS[@]:1}))
//		fi
//		return 0
//	}
//
//	type -p goes >/dev/null && complete -F _goes -o filenames goes
func (g *Goes) complete(args ...string) error {
	for _, s := range g.Complete(args...) {
		fmt.Println(s)
	}
	return nil
}
