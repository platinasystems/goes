// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import ()

type platform struct {
}

func (p *platform) Init() (err error) {
	if err = p.boardInit(); err != nil {
		return err
	}
	return nil
}

func (p *platform) boardInit() (err error) {
	return nil
}
