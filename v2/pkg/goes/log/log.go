// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package log

import (
	"fmt"
	"log"
	"os"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Flags = log.Lshortfile

var Fatal = log.Fatal

func Plain() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprint(program.Base, ": "))
}

func Style() {
	log.SetOutput(os.Stdout)
	log.SetFlags(Flags)
	log.SetPrefix(fmt.Sprint(program.Base, ":"))
}
