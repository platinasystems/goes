// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/log/style"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		style.System()
	}
	style.Plain.Errata.Print("plain errata")
	style.Plain.Notice.Print("plain notice")
	style.ShortFile.Errata.Print("short file errata")
	style.ShortFile.Notice.Print("short file notice")
}
