// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package echo

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/flags"
)

type echo struct{}

func New() echo { return echo{} }

func (echo) String() string { return "echo" }
func (echo) Usage() string  { return "echo [-n] [STRING]..." }

func (echo) Main(args ...string) error {
	flag, args := flags.New(args, "-n")
	s := strings.Join(args, " ")
	if flag["-n"] {
		fmt.Print(s)
	} else {
		fmt.Println(s)
	}
	return nil
}

func (echo) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a line of text",
	}
}

func (echo) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	echo - print a line of text

SYNOPSIS
	echo [-n] [STRING]...

DESCRIPTION
	Echo the STRING(s) to standard output.

	-n     do not output the trailing newline`,
	}
}
