// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type rsvp struct {
	http.ResponseWriter
	req   *http.Request
	doneC chan empty
}

var rsvpPool = &sync.Pool{
	New: func() any {
		return &rsvp{
			doneC: make(chan empty),
		}
	},
}

func (rsvp *rsvp) done() { rsvp.doneC <- done }

// Just return if errors is nil; otherwise,
// [http.Error] with error string
// and appropriate status code.
func (rsvp *rsvp) reporterr(err error) {
	var code int
	if err == nil {
		return
	}
	if xerrors.IsIncomplete(err) ||
		xerrors.IsInvalid(err) ||
		xerrors.IsRange(err) {
		code = http.StatusBadRequest
	} else if xerrors.IsNotFound(err) {
		code = http.StatusNotFound
	} else if xerrors.IsUnavailable(err) {
		code = http.StatusConflict
	} else if errors.Is(err, fs.ErrPermission) {
		code = http.StatusForbidden
	} else {
		code = http.StatusInternalServerError
	}
	http.Error(rsvp, err.Error(), code)
}

// Print “OK” if nil error; otherwise, [reporterr].
func (rsvp *rsvp) reportok(err error) {
	if err == nil {
		fmt.Fprintln(rsvp, "OK")
	} else {
		rsvp.reporterr(err)
	}
}

func (rsvp *rsvp) reqargs(cmd string) []string {
	s := strings.TrimPrefix(rsvp.req.URL.Path, cmd)
	s = strings.TrimPrefix(s, "/")
	if len(s) == 0 {
		return nil
	}
	return strings.Split(s, "/")
}
