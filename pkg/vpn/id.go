// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type Id uint32

const (
	SizeofId = 4

	SizeofLabel = 2 * SizeofId

	IdBits = SizeofId * 8

	IdVersionBit = IdBits - 4
	IdIndexMask  = (1 << IdVersionBit) - 1

	InvalidId Id = 1<<IdBits - 1

	NextIdVersion = 1 << IdVersionBit
)

var MyId Id
var MyLabel []byte

func MakeId(i int, v uint8) Id {
	return Id(i) | Id(v)<<IdVersionBit
}

func MakeLabel(from, to Id) []byte {
	lbl := make([]byte, SizeofLabel)
	binary.BigEndian.PutUint32(lbl, uint32(from))
	binary.BigEndian.PutUint32(lbl[SizeofId:], uint32(to))
	return lbl
}

func ParseId(s string) (Id, error) {
	u, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	return Id(u), err
}

func Relabel(buf []byte, local, remote Id) {
	i := len(buf) - SizeofId
	binary.BigEndian.PutUint32(buf[i:], uint32(remote))
	i -= SizeofId
	binary.BigEndian.PutUint32(buf[i:], uint32(local))
}

func ScanLabel(buf []byte) (to, from Id) {
	i := len(buf) - SizeofId
	to = Id(binary.BigEndian.Uint32(buf[i:]))
	i -= SizeofId
	from = Id(binary.BigEndian.Uint32(buf[i:]))
	return
}

func TruncLabel(buf []byte) []byte {
	return buf[:len(buf)-SizeofLabel]
}

func (id Id) Index() int {
	return int(id & IdIndexMask)
}

func (id Id) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(id), 16)), nil
}

func (id *Id) Revise() {
	*id += NextIdVersion
}

func (id Id) String() string {
	return fmt.Sprint(id.Index(), ".", id.Version())
}

func (p *Id) UnmarshalJSON(b []byte) error {
	id, err := ParseId(string(b))
	if err == nil {
		*p = id
	}
	return err
}

func (id Id) Version() uint8 {
	return uint8(id >> IdVersionBit)
}
