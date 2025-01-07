// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xregexp

import (
	"regexp"
	"sync"
)

var compiled struct {
	sync.RWMutex
	re map[string]*regexp.Regexp
}

// Cache regexp compilation.
func CompileOnce(expr string) (*regexp.Regexp, error) {
	if compiled.re != nil {
		compiled.RLock()
		re, ok := compiled.re[expr]
		compiled.RUnlock()
		if ok {
			return re, nil
		}
	}
	compiled.Lock()
	defer compiled.Unlock()
	if compiled.re == nil {
		compiled.re = make(map[string]*regexp.Regexp)
	} else if re, ok := compiled.re[expr]; ok {
		return re, nil
	}
	re, err := regexp.Compile(expr)
	if err == nil {
		compiled.re[expr] = re
	}
	return re, err
}
