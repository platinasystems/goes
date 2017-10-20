// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "github.com/platinasystems/go/internal/nl"

func IfNameByIndex(sr *nl.SockReceiver) (map[int32]string, error) {
	ifnames := make(map[int32]string)
	if req, err := nl.NewMessage(
		nl.Hdr{
			Type:  RTM_GETLINK,
			Flags: nl.NLM_F_REQUEST | nl.NLM_F_DUMP,
		},
		IfInfoMsg{
			Family: AF_UNSPEC,
		},
	); err != nil {
		return nil, err
	} else if err = sr.UntilDone(req, func(b []byte) {
		if nl.HdrPtr(b).Type != RTM_NEWLINK {
			return
		}
		var ifla Ifla
		ifla.Write(b)
		msg := IfInfoMsgPtr(b)
		if val := ifla[IFLA_IFNAME]; len(val) > 0 {
			ifnames[msg.Index] = nl.Kstring(val)
		}
	}); err != nil {
		return nil, err
	}
	return ifnames, nil
}
