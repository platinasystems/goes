// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fou

const FOU_GENL_NAME = "fou"
const FOU_GENL_VERSION uint8 = 1

const (
	FOU_ATTR_UNSPEC            uint16 = iota
	FOU_ATTR_PORT                     // u16
	FOU_ATTR_AF                       // u8
	FOU_ATTR_IPPROTO                  // u8
	FOU_ATTR_TYPE                     // u8
	FOU_ATTR_REMCSUM_NOPARTIAL        // flag

	N_FOU_ATTR
)

const FOU_ATTR_MAX = N_FOU_ATTR - 1

const (
	FOU_CMD_UNSPEC uint8 = iota
	FOU_CMD_ADD
	FOU_CMD_DEL
	FOU_CMD_GET

	N_FOU_CMD
)

const FOU_CMD_MAX = N_FOU_CMD - 1

const (
	FOU_ENCAP_UNSPEC uint8 = iota
	FOU_ENCAP_DIRECT
	FOU_ENCAP_GUE
)
