// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import (
	"fmt"
	"log"
	"os"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	Muteln = func(args ...any) {}
	Mutef  = func(format string, args ...any) {}

	PlainStderr = log.New(os.Stderr, Prefix(), 0)

	ShortFileStderr = log.New(os.Stderr, Prefix(), log.Lshortfile)
	ShortFileStdout = log.New(os.Stdout, Prefix(), log.Lshortfile)
)

func Prefix() string {
	return fmt.Sprint(program.Base, ": ")
}
