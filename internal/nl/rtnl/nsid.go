// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"os"
	"path/filepath"

	"github.com/platinasystems/go/internal/nl"
)

// Send a RTM_GETNSID request to netlink and return the response attribute.
func Nsid(sr *nl.SockReceiver, name string) (int32, error) {
	nsid := int32(-1)

	f, err := os.Open(filepath.Join(VarRunNetns, name))
	if err != nil {
		return -1, err
	}
	defer f.Close()

	if req, err := nl.NewMessage(nl.Hdr{
		Type:  RTM_GETNSID,
		Flags: nl.NLM_F_REQUEST | nl.NLM_F_ACK,
	}, NetnsMsg{
		Family: AF_UNSPEC,
	},
		nl.Attr{NETNSA_FD, nl.Uint32Attr(f.Fd())},
	); err != nil {
		return -1, err
	} else if err = sr.UntilDone(req, func(b []byte) {
		var netnsa Netnsa
		if nl.HdrPtr(b).Type != RTM_NEWNSID {
			return
		}
		n, err := netnsa.Write(b)
		if err != nil || n == 0 {
			return
		}
		if val := netnsa[NETNSA_NSID]; len(val) > 0 {
			nsid = nl.Int32(val)
		}
	}); err != nil {
		return -1, err
	}
	return nsid, nil
}
