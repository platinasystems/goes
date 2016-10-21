// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

const QsfpNChannel = 4

type QsfpSignal uint8

const (
	QsfpLowPowerMode QsfpSignal = iota
	QsfpInterruptL
	QsfpModulePresentL
	QsfpModuleSelectL
	QsfpResetL
	QsfpNSignal
)

type reg8 byte
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
}

// Lower memory map.
// Everything in network byte order.
// Bytes 0-85 are read only; 86-128 are read/write.
type qsfpRegs struct {
	id SfpId

	// [0] Data not ready.  Indicates transceiver has not yet achieved power up and monitor data is
	// not ready.  Bit remains high until data is ready to be read at which time the device sets the bit low.
	// [1] interrupt active low pin value
	status reg16

	// [0] [3:0] per channel latched rx loss of signal
	//     [7:4] per channel latched tx loss of signal (optional)
	// [1] [3:0] per channel latched tx fault
	//   All else is reserved.
	channelStatusInterrupt reg16
	_                      [1]byte

	monitorInterruptStatus monitorInterruptRegs
	_                      [9]byte

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
	}

	_ [86 - 50]byte

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
	_                   [2]reg8

	monitorInterruptDisable monitorInterruptRegs
	_                       [12]byte

	passwordEntryChange [4]reg8
	passwordEntry       [4]reg8

	upperMemoryMapPageSelect reg8

	upperMemory [128]reg8
}

// Upper memory map (page select 0)
// Read only.
type SfpRegs struct {
	Id                           SfpId
	ExtendedId                   byte
	ConnectorType                SfpConnectorType
	Compatibility                [8]byte
	Encoding                     byte
	NominalBitRate100MbitsPerSec byte
	_                            byte
	LinkLength                   [5]byte
	_                            byte
	VendorName                   [16]byte
	_                            byte
	VendorOui                    [3]byte
	VendorPartNumber             [16]byte
	VendorRevision               [4]byte
	LaserWavelengthInNm          [2]byte
	_                            byte
	checksum_0_to_62             byte
	Options                      [2]byte
	MaxBitRateMarginPercent      byte
	MinBitRateMarginPercent      byte
	VendorSerialNumber           [16]byte
	VendorDateCode               [8]byte
	_                            [3]byte
	checksum_63_to_94            byte
	VendorSpecific               [32]byte
}

type qsfpHighLow struct{ hi, lo reg16 }
type qsfpThreshold struct{ alarm, warning struct{ hi, lo reg16 } }

// Upper memory map (page select 3)
type qsfpThresholdRegs struct {
	temperature   qsfpThreshold
	_             [8]byte
	supplyVoltage qsfpThreshold
	_             [176 - 152]byte
	rxPower       qsfpThreshold
	txBiasCurrent qsfpThreshold
	_             [226 - 192]byte

	// Bytes 226-255 are read/write.
	vendorChannelControls     [14]byte
	optionalChannelControlls  [2]byte
	thresholdInterruptDisable [4]byte
	_                         [256 - 246]byte
}
