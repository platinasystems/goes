// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"io/ioutil"
	"sort"
	"strings"
)

func CompleteName(s string) (list []string) {
	names := func() (list []string) {
		if dir, err := ioutil.ReadDir("/var/run/netns"); err == nil {
			for _, fi := range dir {
				list = append(list, fi.Name())
			}
		}
		return
	}
	if len(s) == 0 {
		list = names()
	} else {
		for _, name := range names() {
			if strings.HasPrefix(name, s) {
				list = append(list, name)
			}
		}
	}
	if len(list) > 0 {
		sort.Strings(list)
	}
	return
}
