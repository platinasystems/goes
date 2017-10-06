// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netns

import (
	"strings"
)

func CompleteName(s string) (list []string) {
	list = List()
	if len(s) != 0 {
		for i := 0; i < len(list); {
			if strings.HasPrefix(list[i], s) {
				i++
			} else {
				copy(list[i:len(list)], list[i+1:])
				list = list[:len(list)-1]
			}
		}
	}
	return
}
