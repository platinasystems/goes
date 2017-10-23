// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "github.com/platinasystems/go/internal/nl"

var If struct {
	NameByIndex  map[int32]string
	IndexByName  map[string]int32
	IndicesByGid map[uint32][]int32
	Master       map[int32]int32
}

func MakeIfMaps(sr *nl.SockReceiver) error {
	If.NameByIndex = make(map[int32]string)
	If.IndexByName = make(map[string]int32)
	If.IndicesByGid = make(map[uint32][]int32)
	If.Master = make(map[int32]int32)
	req, err := nl.NewMessage(
		nl.Hdr{
			Type:  RTM_GETLINK,
			Flags: nl.NLM_F_REQUEST | nl.NLM_F_DUMP,
		},
		IfInfoMsg{
			Family: AF_UNSPEC,
		},
	)
	if err == nil {
		err = sr.UntilDone(req, func(b []byte) {
			if nl.HdrPtr(b).Type != RTM_NEWLINK {
				return
			}
			var ifla Ifla
			ifla.Write(b)
			msg := IfInfoMsgPtr(b)
			name := nl.Kstring(ifla[IFLA_IFNAME])
			gid := nl.Uint32(ifla[IFLA_GROUP])
			If.NameByIndex[msg.Index] = name
			If.IndexByName[name] = msg.Index
			If.IndicesByGid[gid] = append(If.IndicesByGid[gid],
				msg.Index)
			if val := ifla[IFLA_MASTER]; len(val) > 0 {
				If.Master[msg.Index] = nl.Int32(val)
			}
		})
	}
	return err
}
