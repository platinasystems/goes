// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IFA_F_SECONDARY uint8 = 1 << iota
	IFA_F_NODAD
	IFA_F_OPTIMISTIC
	IFA_F_DADFAILED
	IFA_F_HOMEADDRESS
	IFA_F_DEPRECATED
	IFA_F_TENTATIVE
	IFA_F_PERMANENT
)

const IFA_F_TEMPORARY = IFA_F_SECONDARY

const (
	IFA_F_MANAGETEMPADDR uint32 = 1 << (iota + 8)
	IFA_F_NOPREFIXROUTE
	IFA_F_MCAUTOJOIN
	IFA_F_STABLE_PRIVACY
)
