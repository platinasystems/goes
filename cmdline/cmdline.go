// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package cmdline maps /proc/cmdline
package cmdline

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"sort"
)

var File = "/proc/cmdline"

func New() (keys []string, m map[string]string, err error) {
	bf, err := ioutil.ReadFile(File)
	if err != nil {
		return
	}
	clre, err := regexp.Compile("\\S+='.+'|\\S+=\".+\"|\\S+=\\S+|\\S+")
	if err != nil {
		return
	}
	m = make(map[string]string)
	for _, bl := range bytes.Split(bf, []byte{'\n'}) {
		if len(bl) == 0 {
			continue
		}
		for _, b := range clre.FindAll(bl, -1) {
			eq := bytes.Index(b, []byte{'='})
			if eq <= 1 || eq == len(b)-1 {
				m[string(b)] = "true"
				continue
			}
			name := string(b[:eq])
			var value []byte
			if b[eq+1] == '\'' || b[eq+1] == '"' {
				value = b[eq+2 : len(b)-1]
			} else {
				value = b[eq+1:]
			}
			m[name] = string(value)
		}
	}
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}
