// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package cmdline maps /proc/cmdline
package cmdline

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Cmdline map[string]string

var File = "/proc/cmdline"

func New() (keys []string, m Cmdline, err error) {
	bf, err := ioutil.ReadFile(File)
	if err != nil {
		return
	}
	clre, err := regexp.Compile("\\S+='.+'|\\S+=\".+\"|\\S+=\\S+|\\S+")
	if err != nil {
		return
	}
	m = make(Cmdline)
	for _, bl := range bytes.Split(bf, []byte{'\n'}) {
		if len(bl) == 0 {
			continue
		}
		for _, b := range clre.FindAll(bl, -1) {
			m.Set(string(b))
		}
	}
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

// Set() is m[KEY] = VALUE if kv has '=' and m[KEY] = "true" otherwise
func (m Cmdline) Set(kv string) {
	eq := strings.Index(kv, "=")
	if eq <= 1 || eq == len(kv)-1 {
		m[kv] = "true"
	} else if kv[eq+1] == '\'' || kv[eq+1] == '"' {
		m[kv[:eq]] = kv[eq+2 : len(kv)-1]
	} else {
		m[kv[:eq]] = kv[eq+1:]
	}
}

// String() reformats the command line
func (m Cmdline) String() string {
	buf := make([]byte, 0, os.Getpagesize())
	for k, v := range m {
		if len(buf) > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, []byte(k)...)
		if v != "true" {
			buf = append(buf, '=')
			buf = append(buf, []byte(v)...)
		}
	}
	return string(buf)
}
