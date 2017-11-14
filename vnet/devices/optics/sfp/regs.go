// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

const QsfpNChannel = 4

type QsfpSignal uint8

const (
	QsfpLowPowerMode QsfpSignal = iota
	QsfpInterruptStatus
	QsfpModuleIsPresent
	QsfpModuleSelected
	QsfpResetIsActive
	QsfpNSignal
)

type reg8 uint8
type reg16 [2]reg8
type regi16 reg16

type QsfpAlarmStatus uint8

// 4 bit alarm status for temperature, voltage, ...
const (
	QsfpLoWarning QsfpAlarmStatus = 1 << iota
	QsfpHiWarning
	QsfpLoAlarm
	QsfpHiAlarm
)

type monitorInterruptRegs struct {
	// [0] [7:4] latched temperature alarm status
	// [1] [7:4] latched supply voltage alarm status
	//   All else is reserved.
	module reg16

	_ [1]byte

	// [0] [7:4] rx channel 0 power alarm status
	//     [3:0] rx channel 1 power alarm status
	// [1] same for channels 2 & 3
	channelRxPower reg16

	// [0], [1] rx channel 0-3 tx bias current alarm status
	channelTxBiasCurrent reg16

	// [0], [1] rx channel 0-3 tx power alarm status
	channelTxPower reg16
}

// Lower memory map.
// Everything in network byte order.
// Bytes 0-85 are read only; 86-128 are read/write.
type qsfpRegs struct {
	id Id

	// [0] Data not ready.  Indicates transceiver has not yet achieved power up and monitor data is
	// not ready.  Bit remains high until data is ready to be read at which time the device sets the bit low.
	// [1] interrupt active low pin value
	status reg16

	// [0] [3:0] per channel latched rx loss of signal
	//     [7:4] per channel latched tx loss of signal (optional)
	// [1] [3:0] per channel latched tx fault
	//     [7:4] per channel latched tx adaptive EQ fault (optional)
	//   All else is reserved.
	channelStatusInterrupt reg16
	channelStatusLOL       reg8

	monitorInterruptStatus monitorInterruptRegs
	_                      [7]byte

	// Module Monitoring Values.
	internallyMeasured struct {
		// signed 16 bit, units of degrees Celsius/256
		temperature regi16
		_           [2]byte
		// 16 unsigned; units of 100e-6 Volts
		supplyVoltage reg16
		_             [6]byte
		// Channel Monitoring Values.
		// unsigned 16 bit, units of 1e-7 Watts
		rxPower [QsfpNChannel]reg16
		// unsigned 16 bit, units of 2e-6 Amps
		txBiasCurrent [QsfpNChannel]reg16
		txPower       [QsfpNChannel]reg16
	}

	_ [86 - 58]byte

	// Bytes 86 through 128 are all read/write.

	// [3:0] per channel laser disable
	txDisable reg8

	rxRateSelect        reg8
	txRateSelect        reg8
	rxApplicationSelect [4]reg8

	// [1] low power enable
	// [0] override LP_MODE signal; allows software to set low power mode.
	powerControl reg8

	txApplicationSelect [4]reg8
	_                   [21]byte

	passwordEntryChange [4]reg8
	passwordEntry       [4]reg8

	upperMemoryMapPageSelect reg8

	upperMemory [128]reg8
}

// Upper memory map (page select 0)
// Read only.
type Eeprom struct {
	Id            Id
	ExtendedId    reg8
	ConnectorType ConnectorType

	// Byte 131 Compatibility[0] SfpCompliance
	//   [0] 40G active cable xlppi
	//   [1] 40GBASE-LR4
	//   [2] 40GBASE-SR4
	//   [3] 40GBASE-CR4
	//   [4] 10GBASE-SR
	//   [5] 10GBASE-LR
	//   [6] 10GBASE-LRM
	//   [7] Extended (Options[0] becomes SfpExtendedCompliance)
	Compatibility [8]reg8

	Encoding                     reg8
	NominalBitRate100MbitsPerSec reg8
	_                            reg8
	LinkLength                   [5]reg8
	_                            reg8
	VendorName                   [16]reg8
	_                            reg8
	VendorOui                    [3]reg8
	VendorPartNumber             [16]reg8
	VendorRevision               [4]reg8
	LaserWavelengthInNm          [2]reg8
	_                            reg8
	checksum_0_to_62             reg8
	Options                      [2]reg8
	MaxBitRateMarginPercent      reg8
	MinBitRateMarginPercent      reg8
	VendorSerialNumber           [16]reg8
	VendorDateCode               [8]reg8
	_                            [3]reg8
	checksum_63_to_94            reg8
	VendorSpecific               [32]reg8
}

type qsfpHighLow struct{ hi, lo reg16 }
type qsfpThreshold struct{ alarm, warning struct{ hi, lo reg16 } }

// Upper memory map (page select 3)
type qsfpThresholdRegs struct {
	_             [128]reg8
	temperature   qsfpThreshold
	_             [144 - 136]reg8
	supplyVoltage qsfpThreshold
	_             [176 - 152]reg8
	rxPower       qsfpThreshold
	txBiasCurrent qsfpThreshold
	txPower       qsfpThreshold
	_             [226 - 200]reg8

	// Bytes 226-255 are read/write.
	vendorChannelControls     [14]reg8
	optionalChannelControlls  [2]reg8
	thresholdInterruptDisable [4]reg8
	_                         [256 - 246]reg8
}
