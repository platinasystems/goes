// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mbr

const (
	MediaTypeOff = 0x15
	MagicOffL    = 0x1fe
	MagicOffM    = 0x1ff
	MagicValL    = 0x55
	MagicValM    = 0xaa
)

func isFatValidMedia(s []byte) bool {
	return 0xf8 <= s[MediaTypeOff] || s[MediaTypeOff] == 0xf0
}

func Probe(s []byte) bool {
	if s[MagicOffL] == MagicValL && s[MagicOffM] == MagicValM &&
		!isFatValidMedia(s) {
		return true
	}
	return false
}
