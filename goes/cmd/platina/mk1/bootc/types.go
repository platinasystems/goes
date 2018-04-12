// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

type RegInfo struct {
	Mac    string
	IP     string
	Images []string
}
type RegReply struct {
	ReplyType int
	TorName   string
	ImageName string
	Script    string
	Error     error
}

var regInfo RegInfo
var regReply RegReply
