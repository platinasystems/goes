// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import "fmt"

// Unless already set, use args[0] for "name".
func (opt *Options) Name(args []string) ([]string, error) {
	if len(args) >= 1 {
		if name := opt.Parms.ByName["name"]; len(name) > 0 {
			return args,
				fmt.Errorf("%v unexpected, already named %q",
					args, name)
		}
		opt.Parms.ByName["name"] = args[0]
		args = args[1:]
	}
	return args, nil
}

// Error on args[1:]; otherwise use args[0] for "name".
func (opt *Options) OnlyName(args []string) error {
	a, err := opt.Name(args)
	if err == nil && len(a) > 0 {
		err = fmt.Errorf("%v unexpected", a)
	}
	return err
}
