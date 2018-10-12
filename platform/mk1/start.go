// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mk1

import (
	"os"

	"github.com/platinasystems/go/internal/machine"
)

func Start(name string) (err error) {
	machine.Name = name
	return Goes.Main(os.Args...)
}
