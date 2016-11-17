// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package nsid

import "strings"

func (*nsid) Help(args ...string) string {
	help := "no help"
	if strings.HasSuffix(args[0], "nsid") {
		args = args[1:]
	}
	if args[0] == "set" || args[0] == "unset" {
		switch len(args) {
		case 1:
			help = "NAME"
		case 2:
			help = "ID"
		}
	}
	return help
}
