// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"github.com/platinasystems/go/elib"
)

// SFP Ids from eeprom.
type SfpId uint8

const (
	IdUnknown SfpId = iota
	IdGbic
	IdOnMotherboard
	IdSfp
	IdXbi
	IdXenpak
	IdXfp
	IdXff
	IdXfpE
	IdXpak
	IdX2
	IdDwdmSfp
	IdQsfp
	IdQsfpPlus
	IdCxp
	IdShieldedMiniMultilaneHD4X
	IdShieldedMiniMultilaneHD8X
	IdQsfp28
	IdCxp2
	IdCdfpStyle12
	IdShieldedMiniMultilaneHD4XFanout
	IdShieldedMiniMultilaneHD8XFanoutCable
	IdCdfpStyle3
	IdMicroQsfp
	// - 0x7F: Reserved
	// 0x80 - 0xFF: Vendor Specific
)

var sfpIdNames = [...]string{
	0x00: "Unknown or unspecified",
	0x01: "GBIC",
	0x02: "Module/connector soldered to motherboard",
	0x03: "SFP/SFP+/SFP28",
	0x04: "300 pin XBI",
	0x05: "XENPAK",
	0x06: "XFP",
	0x07: "XFF",
	0x08: "XFP-E",
	0x09: "XPAK",
	0x0A: "X2",
	0x0B: "DWDM-SFP/SFP+",
	0x0C: "QSFP",
	0x0D: "QSFP+",
	0x0E: "CXP",
	0x0F: "Shielded Mini Multilane HD 4X",
	0x10: "Shielded Mini Multilane HD 8X",
	0x11: "QSFP28",
	0x12: "CXP2/CXP28",
	0x13: "CDFP (Style 1/Style2)",
	0x14: "Shielded Mini Multilane HD 4X Fanout Cable",
	0x15: "Shielded Mini Multilane HD 8X Fanout Cable",
	0x16: "CDFP (Style 3)",
	0x17: "Micro QSFP",
}

func (i SfpId) String() string { return elib.Stringer(sfpIdNames[:], int(i)) }

type SfpConnectorType uint8

const (
	SfpConnectorUnknown SfpConnectorType = iota
	SfpConnectorSubscriber
	SfpConnectorFibreChannelStyle1
	SfpConnectorFibreChannelStyle2
	SfpConnectorBNCTNC
	SfpConnectorFibreChannelCoax
	SfpConnectorFiberJack
	SfpConnectorLucent
	SfpConnectorMTRJ
	SfpConnectorMU
	SfpConnectorSG
	SfpConnectorOpticalPigtail
	SfpConnectorMPO1x12
	SfpConnectorMPO2x16
	SfpConnectorHSSDC2               SfpConnectorType = 0x20
	SfpConnectorCopperPigtail        SfpConnectorType = 0x21
	SfpConnectorRJ45                 SfpConnectorType = 0x22
	SfpConnectorNoSeparableConnector SfpConnectorType = 0x23
	SfpConnectorMXC2x16              SfpConnectorType = 0x24
)

var sfpConnectorTypeNames = [...]string{
	0x00: "Unknown or unspecified",
	0x01: "SC (Subscriber Connector)",
	0x02: "Fibre Channel Style 1 copper connector",
	0x03: "Fibre Channel Style 2 copper connector",
	0x04: "BNC/TNC (Bayonet/Threaded Neill-Concelman)",
	0x05: "Fibre Channel coax headers",
	0x06: "Fiber Jack",
	0x07: "LC (Lucent Connector)",
	0x08: "MT-RJ (Mechanical Transfer - Registered Jack)",
	0x09: "MU (Multiple Optical)",
	0x0A: "SG",
	0x0B: "Optical Pigtail",
	0x0C: "MPO 1x12 (Multifiber Parallel Optic)",
	0x0D: "MPO 2x16",
	0x20: "HSSDC II (High Speed Serial Data Connector)",
	0x21: "Copper pigtail",
	0x22: "RJ45 (Registered Jack)",
	0x23: "No separable connector",
	0x24: "MXC 2x16",
}

func (i SfpConnectorType) String() string { return elib.Stringer(sfpConnectorTypeNames[:], int(i)) }
