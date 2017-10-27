// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package genl

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/sizeof"
)

const GENL_NAMSIZ = 16 // length of family name

const GENL_MIN_ID = nl.NLMSG_MIN_TYPE
const GENL_MAX_ID uint16 = 1023

const SizeofMsg = (2 * sizeof.Byte) + sizeof.Short

const MSG nl.Align = SizeofMsg

type Msg struct {
	Cmd     uint8
	Version uint8
	_       uint16
}

func MsgPtr(b []byte) *Msg {
	if len(b) < nl.SizeofHdr+MSG.Size() {
		return nil
	}
	return (*Msg)(unsafe.Pointer(&b[nl.SizeofHdr]))
}

func (msg Msg) Read(b []byte) (int, error) {
	*(*Msg)(unsafe.Pointer(&b[0])) = msg
	return MSG.Size(), nil
}

const (
	GENL_ADMIN_PERM uint8 = 1 << iota
	GENL_CMD_CAP_DO
	GENL_CMD_CAP_DUMP
	GENL_CMD_CAP_HASPOL
	GENL_UNS_ADMIN_PERM
)

// List of reserved static generic netlink identifiers:
const (
	GENL_ID_CTRL = nl.NLMSG_MIN_TYPE + iota
	GENL_ID_VFS_DQUOT
	GENL_ID_PMCRAID
	GENL_START_ALLOC
)

// Controller
const (
	CTRL_CMD_UNSPEC uint8 = iota
	CTRL_CMD_NEWFAMILY
	CTRL_CMD_DELFAMILY
	CTRL_CMD_GETFAMILY
	CTRL_CMD_NEWOPS
	CTRL_CMD_DELOPS
	CTRL_CMD_GETOPS
	CTRL_CMD_NEWMCAST_GRP
	CTRL_CMD_DELMCAST_GRP
	CTRL_CMD_GETMCAST_GRP /* unused */

	N_CTRL_CMD
)

const CTRL_CMD_MAX = N_CTRL_CMD - 1

const (
	CTRL_ATTR_UNSPEC uint16 = iota
	CTRL_ATTR_FAMILY_ID
	CTRL_ATTR_FAMILY_NAME
	CTRL_ATTR_VERSION
	CTRL_ATTR_HDRSIZE
	CTRL_ATTR_MAXATTR
	CTRL_ATTR_OPS
	CTRL_ATTR_MCAST_GROUPS

	N_CTRL_ATTR
)

const CTRL_ATTR_MAX = N_CTRL_ATTR - 1

type Gea [N_CTRL_ATTR][]byte

func (gea *Gea) Write(b []byte) (int, error) {
	i := nl.NLMSG.Align(nl.SizeofHdr + MSG.Size())
	if i >= len(b) {
		nl.IndexAttrByType(gea[:], nl.Empty)
		return 0, nil
	}
	nl.IndexAttrByType(gea[:], b[i:])
	return len(b) - i, nil
}

const (
	CTRL_ATTR_OP_UNSPEC uint16 = iota
	CTRL_ATTR_OP_ID
	CTRL_ATTR_OP_FLAGS

	N_CTRL_ATTR_OP
)

const CTRL_ATTR_OP_MAX = N_CTRL_ATTR_OP - 1

const (
	CTRL_ATTR_MCAST_GRP_UNSPEC uint16 = iota
	CTRL_ATTR_MCAST_GRP_NAME
	CTRL_ATTR_MCAST_GRP_ID

	N_CTRL_ATTR_MCAST_GRP
)

const CTRL_ATTR_MCAST_GRP_MAX = N_CTRL_ATTR_MCAST_GRP - 1

func GetFamily(sr *nl.SockReceiver, name string) (uint16, error) {
	seq := nl.Seq()
	req, err := nl.NewMessage(nl.Hdr{
		Type:  GENL_ID_CTRL,
		Flags: nl.NLM_F_REQUEST,
		Seq:   seq,
	}, Msg{
		Cmd: CTRL_CMD_GETFAMILY,
	}, nl.Attr{CTRL_ATTR_FAMILY_NAME, nl.KstringAttr(name)})
	if err != nil {
		return 0, fmt.Errorf("GetFamily(%q).Request: %v", name, err)
	}
	if err := sr.Sock.Send(req); err != nil {
		return 0, fmt.Errorf("GetFamily(%q).Send: %v", name, err)
	}
	ercv := func(err error) (uint16, error) {
		return 0, fmt.Errorf("GetFamily(%q).Receive: %v", name, err)
	}
	for {
		b, err := sr.Recv()
		if err != nil {
			return ercv(err)
		}
		if len(b) < nl.SizeofHdr {
			return ercv(syscall.ENOMSG)
		}
		h := nl.HdrPtr(b)
		if h == nil {
			return ercv(nl.Ehdr(len(b)))
		}
		if int(h.Len) != len(b) {
			return ercv(h.Elen())
		}
		if sr.Sock.Pid != h.Pid {
			return ercv(h.Epid())
		}
		if seq != h.Seq {
			return ercv(h.Eseq())
		}
		switch h.Type {
		case nl.NLMSG_ERROR:
			e := nl.NlmsgerrPtr(b)
			return ercv(syscall.Errno(-e.Errno))
		case GENL_ID_CTRL:
			msg := MsgPtr(b)
			if msg.Cmd == CTRL_CMD_NEWFAMILY {
				var gea Gea
				gea.Write(b)
				return nl.Uint16(gea[CTRL_ATTR_FAMILY_ID]), nil
			}
		}
	}
	return ercv(syscall.ENOTRECOVERABLE)
}
