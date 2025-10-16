// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net/http"
	"strings"
	"sync"
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

func (rsvp *rsvp) reqargs(cmd string) []string {
	s := strings.TrimPrefix(rsvp.req.URL.Path, cmd)
	s = strings.TrimPrefix(s, "/")
	if len(s) == 0 {
		return nil
	}
	return strings.Split(s, "/")
}
