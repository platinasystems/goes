// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"log"
	"os"
)

const LogFlags = log.Lshortfile

var Fatal = log.Fatal

func PlainLog() {
	log.SetFlags(0)
	log.SetPrefix(Prog + ": ")
}

func StyleLog() {
	log.SetOutput(os.Stdout)
	log.SetFlags(LogFlags)
	log.SetPrefix(Prog + ":")
}
