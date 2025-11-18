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

var (
	ErrAlreadySubscribed      = errors.New("already subscribed")
	ErrAlreadyPendingApproval = errors.New("already pending approval")
	ErrNameInUse              = errors.New("name in use")
	ErrNone                   = errors.New("none")
	ErrNotAllowed             = errors.New("not allowed")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrUnprocessable          = errors.New("unprocessable")
	ErrVcsMatch               = errors.New("upgrade required")
)

func AlreadySubscribed(args ...any) error {
	return xerrors.Label(ErrAlreadySubscribed, args...)
}

func AlreadyPendingApproval(args ...any) error {
	return xerrors.Label(ErrAlreadyPendingApproval, args...)
}

func NameInUse(args ...any) error {
	return xerrors.Label(ErrNameInUse, args...)
}

func NotAllowed(args ...any) error {
	return xerrors.Label(ErrNotAllowed, args...)
}

func Unauthorized(args ...any) error {
	return xerrors.Label(ErrUnauthorized, args...)
}

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
	reporterr(rsvp.ResponseWriter, err)
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

func reporterr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var code int
	if xerrors.IsIncomplete(err) ||
		xerrors.IsInvalid(err) ||
		xerrors.IsRange(err) {
		code = http.StatusBadRequest
	} else if xerrors.IsNotFound(err) {
		code = http.StatusNotFound
	} else if xerrors.IsUnavailable(err) {
		code = http.StatusConflict
	} else if xerrors.IsBroken(err) {
		code = http.StatusInternalServerError
	} else if errors.Is(err, ErrUnauthorized) {
		code = http.StatusUnauthorized
	} else if errors.Is(err, ErrAlreadySubscribed) {
		code = http.StatusGone
	} else if errors.Is(err, ErrAlreadyPendingApproval) {
		code = http.StatusConflict
	} else if errors.Is(err, ErrNameInUse) {
		code = http.StatusForbidden
	} else if errors.Is(err, ErrNone) {
		code = http.StatusNoContent
	} else if errors.Is(err, ErrNotAllowed) {
		code = http.StatusMethodNotAllowed
	} else if errors.Is(err, ErrUnprocessable) {
		code = http.StatusUnprocessableEntity
	} else if errors.Is(err, ErrVcsMatch) {
		code = http.StatusUpgradeRequired
	} else if errors.Is(err, fs.ErrNotExist) {
		code = http.StatusNotFound
	} else if errors.Is(err, fs.ErrPermission) {
		code = http.StatusForbidden
	} else {
		code = http.StatusInternalServerError
	}
	http.Error(w, err.Error(), code)
}
