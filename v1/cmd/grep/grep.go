// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grep

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "grep" }

func (Command) Usage() string {
	return "grep [REGEXP]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print lines matching regular expression pattern",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print lines matching regexp.`,
	}
}

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		return errors.New("Missing regexp")
	}
	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args)
	}
	regex, err := regexp.Compile(args[0])
	if err != nil {
		return fmt.Errorf("Unable to compile %s: %s", args[0], err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if regex.MatchString(scanner.Text()) {
			fmt.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading standard input: %s", err)
	}
	return nil
}
