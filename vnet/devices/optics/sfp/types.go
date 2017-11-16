// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"github.com/platinasystems/go/elib"
)

// SFP Ids from eeprom.
type Id reg8

const (
	IdUnknown Id = iota
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
	IdQsfpDD
	// - 0x7F: Reserved
	// 0x80 - 0xFF: Vendor Specific
)

func (i Id) String() string {
	var t = [...]string{
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
		0x18: "QSFP-DD",
	}
	return elib.Stringer(t[:], int(i))
}

type ConnectorType reg8

const (
	ConnectorUnknown ConnectorType = iota
	ConnectorSubscriber
	ConnectorFibreChannelStyle1
	ConnectorFibreChannelStyle2
	ConnectorBNCTNC
	ConnectorFibreChannelCoax
	ConnectorFiberJack
	ConnectorLucent
	ConnectorMTRJ
	ConnectorMU
	ConnectorSG
	ConnectorOpticalPigtail
	ConnectorMPO1x12
	ConnectorMPO2x16
	ConnectorHSSDC2               ConnectorType = 0x20
	ConnectorCopperPigtail        ConnectorType = 0x21
	ConnectorRJ45                 ConnectorType = 0x22
	ConnectorNoSeparableConnector ConnectorType = 0x23
	ConnectorMXC2x16              ConnectorType = 0x24
)

func (i ConnectorType) String() string {
	var t = [...]string{
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
	return elib.Stringer(t[:], int(i))
}

type Compliance byte

const (
	Log2Compliance40GXLPPI, Compliance40GXLPPI = iota, 1 << iota
	Log2Compliance40G_LR, Compliance40G_LR
	Log2Compliance40G_SR, Compliance40G_SR
	Log2Compliance40G_CR, Compliance40G_CR
	Log2Compliance10G_SR, Compliance10G_SR
	Log2Compliance10G_LR, Compliance10G_LR
	Log2Compliance10G_LRM, Compliance10G_LRM
	Log2ComplianceExtendedValid, ComplianceExtendedValid
)

func (i Compliance) String() string {
	var t = [...]string{
		Log2Compliance40GXLPPI:      "40G XLPPI",
		Log2Compliance40G_LR:        "40G LR",
		Log2Compliance40G_SR:        "40G SR",
		Log2Compliance40G_CR:        "40G CR",
		Log2Compliance10G_SR:        "10G SR",
		Log2Compliance10G_LR:        "10G LR",
		Log2Compliance10G_LRM:       "10G LRM",
		Log2ComplianceExtendedValid: "extended",
	}
	return elib.FlagStringer(t[:], elib.Word(i))
}

type ExtendedCompliance byte

const (
	ExtendedComplianceUnspecified       = iota
	ExtendedCompliance100G_AOC_BER_5e5  // 01h 100G_AOC (Active Optical Cable) or 25GAUI C2M AOC. Providing a worst BER of 5 × 10^(-5)
	ExtendedCompliance100G_SR           // 02h 100GBASE-SR4 or 25GBASE-SR
	ExtendedCompliance100G_LR           // 03h 100GBASE-LR4 or 25GBASE-LR
	ExtendedCompliance100G_ER           // 04h 100GBASE-ER4 or 25GBASE-ER
	ExtendedCompliance100G_SR10         // 05h 100GBASE-SR10
	ExtendedCompliance100G_CWDM4        // 06h 100G CWDM4
	ExtendedCompliance100G_PSM4         // 07h 100G PSM4 Parallel SMF
	ExtendedCompliance100G_ACC_BER_5e5  // 08h 100G ACC (Active Copper Cable) or 25GAUI C2M ACC. Providing a worst BER of 5 × 10^(-5)
	_                                   // 09h Obsolete (assigned before 100G CWDM4 MSA required FEC)
	_                                   // 0Ah Reserved
	ExtendedCompliance100G_CR           // 0Bh 100GBASE-CR4 or 25GBASE-CR CA-L
	ExtendedCompliance25G_CR_CA_S       // 0Ch 25GBASE-CR CA-S
	ExtendedCompliance25G_CR_CA_N       // 0Dh 25GBASE-CR CA-N
	_                                   // 0Eh Reserved
	_                                   // 0Fh Reserved
	ExtendedCompliance40G_ER            // 10h 40GBASE-ER4
	ExtendedCompliance4x10G_SR          // 11h 4 x 10GBASE-SR
	ExtendedCompliance40G_PSM4          // 12h 40G PSM4 Parallel SMF
	ExtendedComplianceG959_1_P1I1_2D1   // 13h G959.1 profile P1I1-2D1 (10709 MBd, 2km, 1310nm SM)
	ExtendedComplianceG959_1_P1S1_2D2   // 14h G959.1 profile P1S1-2D2 (10709 MBd, 40km, 1550nm SM)
	ExtendedComplianceG959_1_P1L1_2D2   // 15h G959.1 profile P1L1-2D2 (10709 MBd, 80km, 1550nm SM)
	ExtendedCompliance10GBASE_T         // 16h 10GBASE-T with SFI electrical interface
	ExtendedCompliance100G_CLR4         // 17h 100G CLR4
	ExtendedCompliance100G_AOC_BER_1e12 // 18h 100G AOC or 25GAUI C2M AOC. Providing a worst BER of 10^(-12) or below
	ExtendedCompliance100G_ACC_BER_1e12 // 19h 100G ACC or 25GAUI C2M ACC. Providing a worst BER of 10^(-12) or below
	ExtendedCompliance100GE_DWDM2       // 1Ah 100GE-DWDM2 (DWDM transceiver using 2 wavelengths on a 1550nm DWDM grid with a reach up to 80km)
)

func (i ExtendedCompliance) String() string {
	var t = [...]string{
		ExtendedComplianceUnspecified:       "unspecified",
		ExtendedCompliance100G_AOC_BER_5e5:  "100G AOC BER < 5e-5",
		ExtendedCompliance100G_SR:           "100GBASE-SR4",
		ExtendedCompliance100G_LR:           "100GBASE-LR4",
		ExtendedCompliance100G_ER:           "100GBASE-ER4",
		ExtendedCompliance100G_SR10:         "100GBASE-SR10",
		ExtendedCompliance100G_CWDM4:        "100G CWDM4",
		ExtendedCompliance100G_PSM4:         "100G PSM4 Parallel SMF",
		ExtendedCompliance100G_ACC_BER_5e5:  "100G ACC BER < 5e-5",
		ExtendedCompliance100G_CR:           "100GBASE-CR4 or 25GBASE-CR CA-L",
		ExtendedCompliance25G_CR_CA_S:       "25GBASE-CR CA-S",
		ExtendedCompliance25G_CR_CA_N:       "25GBASE-CR CA-N",
		ExtendedCompliance40G_ER:            "40GBASE-ER4",
		ExtendedCompliance4x10G_SR:          "4 x 10GBASE-SR",
		ExtendedCompliance40G_PSM4:          "40G PSM4",
		ExtendedComplianceG959_1_P1I1_2D1:   "G959.1 profile P1I1-2D1 (10709 MBd, 2km, 1310nm SM)",
		ExtendedComplianceG959_1_P1S1_2D2:   "G959.1 profile P1S1-2D2 (10709 MBd, 40km, 1550nm SM)",
		ExtendedComplianceG959_1_P1L1_2D2:   "G959.1 profile P1L1-2D2 (10709 MBd, 80km, 1550nm SM)",
		ExtendedCompliance10GBASE_T:         "10GBASE-T with SFI electrical interface",
		ExtendedCompliance100G_CLR4:         "100G CLR4",
		ExtendedCompliance100G_AOC_BER_1e12: "100G AOC BER < 1e-12",
		ExtendedCompliance100G_ACC_BER_1e12: "100G ACC BER < 1e-12",
		ExtendedCompliance100GE_DWDM2:       "100GE-DWDM2",
	}
	return elib.Stringer(t[:], int(i))
}

type ChannelStatusInterrupt uint16

func (i ChannelStatusInterrupt) String() string {
	var t = [...]string{
		0x0001: "L-Rx1 LOS",
		0x0002: "L-Rx2 LOS",
		0x0004: "L-Rx3 LOS",
		0x0008: "L-Rx4 LOS",
		0x0010: "L-Tx1 LOS",
		0x0020: "L-Tx2 LOS",
		0x0040: "L-Tx3 LOS",
		0x0080: "L-Tx4 LOS",
		0x0100: "L-Tx1 Fault",
		0x0200: "L-Tx2 Fault",
		0x0400: "L-Tx3 Fault",
		0x0800: "L-Tx4 Fault",
		0x1000: "L-Tx1 Adapt EQ Fault",
		0x2000: "L-Tx2 Adapt EQ Fault",
		0x4000: "L-Tx3 Adapt EQ Fault",
		0x8000: "L-Tx4 Adapt EQ Fault",
	}
	return elib.Stringer(t[:], int(i))
}

type ModuleInterrupts uint16

func (i ModuleInterrupts) String() string {
	var t = [...]string{
		0x0010: "L-Temp Low Warning",
		0x0020: "L-Temp High Warning",
		0x0040: "L-Temp Low Alarm",
		0x0080: "L-Temp High Alarm",
		0x1000: "L-Vcc Low Warning",
		0x2000: "L-Vcc High Warning",
		0x4000: "L-Vcc Low Alarm",
		0x8000: "L-Vcc High Alarm",
	}
	return elib.Stringer(t[:], int(i))
}

type ChannelStatusLOL byte

func (i ChannelStatusLOL) String() string {
	var t = [...]string{
		0x01: "L-Rx1 LOL",
		0x02: "L-Rx2 LOL",
		0x04: "L-Rx3 LOL",
		0x08: "L-Rx4 LOL",
		0x10: "L-Tx1 LOL",
		0x20: "L-Tx2 LOL",
		0x40: "L-Tx3 LOL",
		0x80: "L-Tx4 LOL",
	}
	return elib.Stringer(t[:], int(i))
}

type ChannelRxPowerInterrupts uint16
type ChannelTxBiasInterrupts uint16
type ChannelTxPowerInterrupts uint16

func (i ChannelRxPowerInterrupts) String() string {
	var t = [...]string{
		0x0001: "L-Rx2 Power Low Warning",
		0x0002: "L-Rx2 Power High Warning",
		0x0004: "L-Rx2 Power Low Alarm",
		0x0008: "L-Rx2 Power High Alarm",
		0x0010: "L-Rx1 Power Low Warning",
		0x0020: "L-Rx1 Power High Warning",
		0x0040: "L-Rx1 Power Low Alarm",
		0x0080: "L-Rx1 Power High Alarm",
		0x0100: "L-Rx4 Power Low Warning",
		0x0200: "L-Rx4 Power High Warning",
		0x0400: "L-Rx4 Power Low Alarm",
		0x0800: "L-Rx4 Power High Alarm",
		0x1000: "L-Rx3 Power Low Warning",
		0x2000: "L-Rx3 Power High Warning",
		0x4000: "L-Rx3 Power Low Alarm",
		0x8000: "L-Rx3 Power High Alarm",
	}
	return elib.Stringer(t[:], int(i))
}

func (i ChannelTxBiasInterrupts) String() string {
	var t = [...]string{
		0x0001: "L-Tx2 Bias Low Warning",
		0x0002: "L-Tx2 Bias High Warning",
		0x0004: "L-Tx2 Bias Low Alarm",
		0x0008: "L-Tx2 Bias High Alarm",
		0x0010: "L-Tx1 Bias Low Warning",
		0x0020: "L-Tx1 Bias High Warning",
		0x0040: "L-Tx1 Bias Low Alarm",
		0x0080: "L-Tx1 Bias High Alarm",
		0x0100: "L-Tx4 Bias Low Warning",
		0x0200: "L-Tx4 Bias High Warning",
		0x0400: "L-Tx4 Bias Low Alarm",
		0x0800: "L-Tx4 Bias High Alarm",
		0x1000: "L-Tx3 Bias Low Warning",
		0x2000: "L-Tx3 Bias High Warning",
		0x4000: "L-Tx3 Bias Low Alarm",
		0x8000: "L-Tx3 Bias High Alarm",
	}
	return elib.Stringer(t[:], int(i))
}

func (i ChannelTxPowerInterrupts) String() string {
	var t = [...]string{
		0x0001: "L-Tx2 Power Low Warning",
		0x0002: "L-Tx2 Power High Warning",
		0x0004: "L-Tx2 Power Low Alarm",
		0x0008: "L-Tx2 Power High Alarm",
		0x0010: "L-Tx1 Power Low Warning",
		0x0020: "L-Tx1 Power High Warning",
		0x0040: "L-Tx1 Power Low Alarm",
		0x0080: "L-Tx1 Power High Alarm",
		0x0100: "L-Tx4 Power Low Warning",
		0x0200: "L-Tx4 Power High Warning",
		0x0400: "L-Tx4 Power Low Alarm",
		0x0800: "L-Tx4 Power High Alarm",
		0x1000: "L-Tx3 Power Low Warning",
		0x2000: "L-Tx3 Power High Warning",
		0x4000: "L-Tx3 Power Low Alarm",
		0x8000: "L-Tx3 Power High Alarm",
	}
	return elib.Stringer(t[:], int(i))
}

var StaticRedisFields = []string{
	"qsfp.id",
	"qsfp.compliance",
	"qsfp.partnumber",
	"qsfp.presence",
	"qsfp.serialnumber",
	"qsfp.vendor",
	"qsfp.connectortype",
}

var StaticMonitoringRedisFields = []string{
	"qsfp.rx.power.highAlarmThreshold.units.mW",
	"qsfp.rx.power.highWarnThreshold.units.mW",
	"qsfp.rx.power.lowAlarmThreshold.units.mW",
	"qsfp.rx.power.lowWarnThreshold.units.mW",
	"qsfp.temperature.highAlarmThreshold.units.C",
	"qsfp.temperature.highWarnThreshold.units.C",
	"qsfp.temperature.lowAlarmThreshold.units.C",
	"qsfp.temperature.lowWarnThreshold.units.C",
	"qsfp.tx.bias.highAlarmThreshold.units.mA",
	"qsfp.tx.bias.highWarnThreshold.units.mA",
	"qsfp.tx.bias.lowAlarmThreshold.units.mA",
	"qsfp.tx.bias.lowWarnThreshold.units.mA",
	"qsfp.tx.power.highAlarmThreshold.units.mW",
	"qsfp.tx.power.highWarnThreshold.units.mW",
	"qsfp.tx.power.lowAlarmThreshold.units.mW",
	"qsfp.tx.power.lowWarnThreshold.units.mW",
	"qsfp.vcc.highAlarmThreshold.units.V",
	"qsfp.vcc.highWarnThreshold.units.V",
	"qsfp.vcc.lowAlarmThreshold.units.V",
	"qsfp.vcc.lowWarnThreshold.units.V",
}

var DynamicMonitoringRedisFields = []string{
	"qsfp.alarms.module",
	"qsfp.alarms.channels",
	"qsfp.rx1.power.units.mW",
	"qsfp.rx2.power.units.mW",
	"qsfp.rx3.power.units.mW",
	"qsfp.rx4.power.units.mW",
	"qsfp.temperature.units.C",
	"qsfp.vcc.units.V",
	"qsfp.tx1.bias.units.mA",
	"qsfp.tx1.power.units.mW",
	"qsfp.tx2.bias.units.mA",
	"qsfp.tx2.power.units.mW",
	"qsfp.tx3.bias.units.mA",
	"qsfp.tx3.power.units.mW",
	"qsfp.tx4.bias.units.mA",
	"qsfp.tx4.power.units.mW",
}
