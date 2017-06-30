// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mkdir

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "mkdir"
	Apropos = "make directories"
	Usage   = "mkdir [OPTION]... DIRECTORY..."
	Man     = `
DESCRIPTION
	Create the DIRECTORY(ies), if they do not already exist.

OPTIONS
	-m MODE
		set mode (as in chmod) to the given octal value;
		default, 0777 & umask()

	-p
		don't print error if existing and make parent directories
		as needed

	-v
		print a message for each created directory`

	DefaultMode = 0755
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var perm os.FileMode = DefaultMode

	flag, args := flags.New(args, "-p", "-v")
	parm, args := parms.New(args, "-m")

	if len(parm.ByName["-m"]) == 0 {
		parm.ByName["-m"] = "0755"
	}

	mode, err := strconv.ParseUint(parm.ByName["-m"], 8, 64)
	if err != nil {
		return err
	}

	if mode != DefaultMode {
		old := syscall.Umask(0)
		defer syscall.Umask(old)
		perm = os.FileMode(mode)
	}

	f := os.Mkdir
	if flag.ByName["-p"] {
		f = os.MkdirAll
	}

	for _, dn := range args {
		if err := f(dn, perm); err != nil {
			return err
		}
		if flag.ByName["-v"] {
			fmt.Printf("mkdir: created directory ‘%s’\n", dn)
		}
	}
	return nil
}
