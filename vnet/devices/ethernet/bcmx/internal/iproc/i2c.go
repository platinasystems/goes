package iproc

// BCM I2C SMBUS Registers
// NOTE: All Register Bits/Fields are R/W, unless specified otherwise.
type I2cRegs struct {
	// [31] RESET
	//      When set, reset SMBUS block to its default state.
	//
	// [30] SMB_EN
	//      When set to '1', the SMBUS block is enabled for operation.
	//      When set the SMBUS block will abort the current transaction in
	//      compliance with the SMBUS master and slave behavior at the end
	//      of the current transaction and stop responding to the SMBUS
	//      master/slave transactions.
	//
	// [29] BIT_BANG_EN
	//      R/W When set to '1', the SMBUS block is placed into bit-bang mode.
	//      SMBUS interface pins are controlled using Bit-Bang Control Register.
	//
	// [28] EN_NIC_SMB_ADDR_0
	//      When set to '1', the SMBUS block responds to slave address 7b0.
	//
	// [27] PROMISCOUS_MODE
	//      When set to '1', the SMBUS block responds to all SMBUS
	//      transactions, regardless of the slave address.
	//
	// [26] TIMESTAMP_CNT_EN
	//      When set to '1', the TIMESTAMP counter is enabled.
	//      When set to ‘0’, the TIMESTAMP counter holds its value.
	//
	// [19:16] MASTER_RTRY_CNT (default: 0xf)
	//         This bit indicates a number of retries. In the case where the
	//         SMBUS block acted as a master and lost an SMBUS arbitration.
	//         HW will retry the transaction as many times as specified in this
	//         register BEFORE it reports loss of arbitration status to the
	//         firmware. When this field is 0, the firmware will not do any
	//         retries, but on the initial attempt, it reports loss of
	//         arbitration.
	Config Reg32

	// [31] MODE_400
	//      When set, the SMBUS block operates in 400KHz mode.
	//      When cleared, SMBUS operates in 100KHz mode.
	//
	// [30:24] RANDOM_SLAVE_STRETCH (default: 0x19)
	//         These bits specify time for which clock low time will be
	//         stretched after each byte (that is ACK bit) when the SMBUS
	//         block acts as a slave.
	//         << This is useful in "legacy mode” to allow firmware time
	//            to handle the data. >>
	//         Note that this time contributes to the slave TLOW:SEXT time,
	//         that is combined random and periodic slave stretch should not
	//         exceed 25ms.
	//         Register has 1ms resolution. Default is 25ms.
	//
	// [23:16] PERIODIC_SLAVE_STRETCH (default: 0x0)
	//         These bits specify time for which each clock period low time will
	//         be stretched when the SMBUS block acts as a slave.
	//         Note that a cumulative clock low extend time (TLOW:SEXT) for
	//         which slave device is allowed to stretch the clock from the
	//         beginning to end of the message (that is from START to STOP) is
	//         25ms. For example, if the message is a Block Write transaction,
	//         then the 36B long allowed periodic stretch would be:
	//         25ms/(9*36) ~= 77us.
	//         This is assuming that random slave stretching is not used.
	//         Register has 1us resolution.
	//
	// [15:8] SMBUS_IDLE_TIME (default 0x32)
	//        These bits specify the time for which both SMBCLK and SMBDAT
	//        must be high before a master can assume that bus is free.
	//        Register has 1us resolution.
	//        Default is 50us
	Timing_config Reg32

	// [7]   Enable
	// [6:0] Address
	//       4 slave addesses to listen on when not in promiscuous mode.
	Slave_bus_address Reg32

	// [31] RX_FIFO_FLUSH
	//      When set, HW will flush the Rx fifo when the current Rx
	//      transaction completes.
	// [30] TX_FIFO_FLUSH
	//      When set, HW will flush the Tx fifo when the current Tx
	//      transaction completes.
	// [22:16] RX_PKT_COUNT (RO)
	//         Number of pkts in the Slave Rx fifo
	// [13:8] RX_FIFO_THRESHOLD
	//        When the Rx fifo hits this threshold, the SMBUS block will
	//        generate an event for the control processor. When set to 0x0,
	//        this event generation is disabled.
	//        Threshold is specified with the byte resolution.
	Fifo_control [2]Reg32 // master/slave

	Bit_bang_control Reg32

	_ [0x30 - 0x18]byte

	// [31] START_BUSY_COMMAND
	//      This bit is self clearing. When written to a '1', the currently
	//      programmed SMBUS transaction will activate. Writing this bit as
	//      a '0' has no effect. This bit must be read as a '0' before setting
	//      it to prevent un-predictable results.
	// [30] ABORT
	//      Transaction Abort. This bit can be set at any time by the firmware
	//      or the driver in order to abort the transaction. The HW will abort
	//      transaction in compliance with the SMBUS rules and clear the bit
	//      when done.
	// [27:25] STATUS (R)
	//         These bits encode status of the last master transaction.
	//         Valid when START_BUSY is cleared after it was set.
	//         000 = Transaction completed successfully
	//         001 = Lost Arbitration. Firmware should restart transaction, if required.
	//         010 = NACK detected after (slave address) first byte. Indicates that
	//                slave device is off-line.
	//         011 = NACK detected after byte other than first. Indicates that slave device maybe busy.
	//         100 = TIMEOUT_MIN exceeded. Indicates that slave device held bus for more then 25ms
	//                (assuming that our clock driver is functioning properly).
	//         101 = TLOW:MEXT exceeded. Used for non-standard transactions where payload is longer
	//                than the TX FIFO. Indicates FIFO under-run (firmware didn’t provide additional data within 10ms).
	//         110 = TLOW:MEXT exceeded. Indicates that read transaction failed due the Master
	//                Rx FIFO being full for more than 10ms.
	//         111 = Reserved
	// [12:9] SMBUS_PROTOCOL
	//         see Operation defined above.
	// [8] PEC
	//     Parity Error should be checked or calculated for this transaction.
	// [7:0] RD_BYTE_COUNT
	//       Number of bytes that SMBUS block should read from the slave in
	//       Block Write/Block Read Process Call or Block Read. If this field
	//       is 0 the SMBUS block will assume that first byte received from the
	//       slave is a Byte Count. If different than 0 HW will use this value
	//       as an indication as where to insert NACK and STOP that is end the
	//       transaction.
	Command [2]Reg32 // master/slave

	// Master:
	// [31] rx_fifo_full
	//       This bit is set when the master receive FIFO become full.
	//       Writing a '1' to this position will clear this bit. When this bit
	//       is '1', the SMB0_EVENT bit will be '1' in each processor.
	// [30] rx_threshold_hit
	//       This bit is set when the master receive FIFO is equal to or 0x0
	//       larger than the Master RX_FIFO_THRESHOLD. Writing a '1'
	//       to this position will clear this bit. When this bit is '1', the
	//       SMB0_EVENT bit will be '1' in each processor.
	// [29] rx_event
	//       This bit is set when the master receive FIFO holds at least 0x0
	//       one valid transaction. Writing a '1' to this position will clear
	//       this bit. When this bit is '1', the SMB0_EVENT bit will be '1' in
	//       each processor.
	// [28] start_busy
	//       This bit is set when master START_BUSY transitions from 1 to 0.
	//       Writing a '1' to this position will clear this bit. When this
	//       bit is '1', the SMB0_EVENT bit will be '1' in each processor.
	// [27] tx underrun
	//	     This bit is set when Master Tx FIFO becomes empty and less 0x0
	//       then PKT_LENGTH bytes were output on the SMBUS.
	//
	// Slave:
	// [26] rx_fifo_full
	//       Same description as Master Bit above.
	// [25] rx_threshold_hit
	//       Same description as Master Bit above.
	// [24] rx_event
	//       Same description as Master Bit above.
	// [23] start_busy
	//       Same description as Master Bit above.
	// [22] tx underrun
	//       Same description as Master Bit above.
	// [21] read event
	//       This bit is set when the slave hardware detected read transaction
	//       directed toward the SMBUS block. Writing a '1' to this position
	//       will clear this bit. When this bit is '1', the SMB0_EVENT bit will
	//       be '1' in each processor.
	Interrupt_enable Reg32

	// Interrupt status write 1 to clear interrupt.
	Interrupt_status_write_1_to_clear Reg32

	// write:
	//  [31] MASTER_WRITE_STATUS
	//       '0' - byte other than last in an SMBus transaction
	//       '1' - End of SMBus transaction
	//  [7:0] DATA
	//
	// read:
	//  [31:30] MASTER_READ_STATUS
	//          { 0 => fifo empty, 1 => start, 2 => middle, 3 => end }
	//  [29:28] PEC_ERROR
	//           This bit indicates status of the PEC checking. HW will check
	//           the PEC only in case where PEC bit in SMBUS Master Command
	//           Register was set for the transaction. This field is valid
	//           only when RD_STATUS = 2’b11
	//  [7:0] DATA
	Data_fifo [2]struct{ Write, Read Reg32 }
}
