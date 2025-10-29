// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "encoding/binary"

var ByteOrder = binary.BigEndian

// [binary.Append] [ByteOrder] to end of buf.
func ByteOrderAppend(buf []byte, val any) ([]byte, error) {
	return binary.Append(buf, ByteOrder, val)
}

// [binary.Decode] [ByteOrder] from beginning of buf.
func ByteOrderDecode(buf []byte, ptr any) (int, error) {
	return binary.Decode(buf, ByteOrder, ptr)
}

// [binary.Encode] [ByteOrder] to beginning of buf.
func ByteOrderEncode(buf []byte, val any) (int, error) {
	return binary.Encode(buf, ByteOrder, val)
}

// [binary.Decode] [ByteOrder] from beginning of buf and,
// if successful, return remainder.
func ByteOrderRemove(data []byte, ptr any) ([]byte, error) {
	n, err := binary.Decode(data, ByteOrder, ptr)
	if err == nil {
		data = data[n:]
	}
	return data, err
}
