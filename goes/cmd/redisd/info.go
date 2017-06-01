// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package redisd

import (
	"fmt"
	"io"
)

func (redisd *Redisd) infoServer(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoClients(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoMemory(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoStats(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}

func (redisd *Redisd) infoCpu(w io.Writer) error {
	_, err := fmt.Fprintln(w, "FIXME")
	return err
}
