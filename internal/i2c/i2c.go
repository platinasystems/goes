// Package i2c supports Linux I2C devices.
package i2c

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// /dev/i2c-X ioctl commands.  The ioctl's parameter is always an unsigned long, except for:
//   - I2C_FUNCS, takes pointer to an unsigned long
//   - I2C_RDWR, takes pointer to struct i2c_rdwr_ioctl_data
//	 - I2C_SMBUS, takes pointer to struct i2c_smbus_ioctl_data
type IoctlOp uintptr

const (
	I2C_RETRIES     IoctlOp = 0x0701 /* number of times a device address should be polled when not acknowledging */
	I2C_TIMEOUT     IoctlOp = 0x0702 /* Set timeout in units of 10 ms */
	I2C_SLAVE       IoctlOp = 0x0703 /* Use this slave address */
	I2C_SLAVE_FORCE IoctlOp = 0x0706 /* Use this slave address, even if it is already in use by a driver. */
	I2C_TENBIT      IoctlOp = 0x0704 /* 0 for 7 bit addrs, != 0 for 10 bit */
	I2C_FUNCS       IoctlOp = 0x0705 /* Get the adapter functionality mask */
	I2C_RDWR        IoctlOp = 0x0707 /* Combined R/W transfer (one STOP only) */
	I2C_PEC         IoctlOp = 0x0708 /* != 0 to use PEC with SMBus */
	I2C_SMBUS       IoctlOp = 0x0720 /* SMBus transfer */
)

type Bus struct {
	// Bus index X for corresponding device /dev/i2c-X
	index int

	// File descriptor for /dev/i2c-INDEX
	fd int

	features FeatureFlag
}

func New(index, address int) (*Bus, error) {
	bus := new(Bus)
	err := bus.Open(index)
	if err == nil {
		err = bus.ForceSlaveAddress(address)
	}
	return bus, err
}

func (b *Bus) Open(index int) (err error) {
	path := fmt.Sprintf("/dev/i2c-%d", index)
	fd, err := syscall.Open(path, syscall.O_RDWR, 0)
	if err != nil {
		err = fmt.Errorf("open %s: %s", path, err)
		return
	}
	b.fd = fd
	b.index = index
	defer func() {
		if err != nil {
			syscall.Close(b.fd)
		}
	}()

	b.features, err = b.GetFeatures()
	if err != nil {
		err = fmt.Errorf("ioctl FUNCS %s: %s", path, err)
		return
	}

	return
}

func (b *Bus) Close() (err error) {
	err = syscall.Close(b.fd)
	return
}

// Do calls function f with given bus and slave device selected.
func Do(index, slave int, f func(bus *Bus) error) (err error) {
	var bus Bus
	if err = bus.Open(index); err != nil {
		return
	}
	defer bus.Close()
	if err = bus.ForceSlaveAddress(slave); err != nil {
		return
	}
	return f(&bus)
}

func ioctlInt(b *Bus, op IoctlOp, arg int) (err error) {
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(b.fd), uintptr(op), uintptr(arg))
	if e != 0 {
		err = e
	}
	return
}

func chk(tag string, err error) error {
	if err != nil {
		err = fmt.Errorf("%s: %s", tag, err)
	}
	return err
}

func (b *Bus) SetRetries(n int) error      { return chk("set retries", ioctlInt(b, I2C_RETRIES, n)) }
func (b *Bus) SetTimeout(n int) error      { return chk("set timeout", ioctlInt(b, I2C_TIMEOUT, n)) }
func (b *Bus) SetSlaveAddress(n int) error { return chk("set slave address", ioctlInt(b, I2C_SLAVE, n)) }
func (b *Bus) ForceSlaveAddress(n int) error {
	return chk("force slave address", ioctlInt(b, I2C_SLAVE_FORCE, n))
}
func (b *Bus) Set10BitAddressing() error {
	return chk("set 10bit addressing", ioctlInt(b, I2C_TENBIT, 1))
}
func (b *Bus) Set7BitAddressing() error {
	return chk("set 7bit addressing", ioctlInt(b, I2C_TENBIT, 0))
}

func (b *Bus) GetFeatures() (mask FeatureFlag, err error) {
	var flags [1]uintptr
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(b.fd), uintptr(I2C_FUNCS), uintptr(unsafe.Pointer(&flags[0])))
	if e != 0 {
		err = e
		return
	}
	mask = FeatureFlag(flags[0])
	return
}

// Functionality (features)
type FeatureFlag uint32

const (
	I2C                    FeatureFlag = 0x00000001
	TenBit_Address         FeatureFlag = 0x00000002
	Protocol_Mangling      FeatureFlag = 0x00000004 /* I2C_M_IGNORE_NAK etc. */
	SMBUS_PEC              FeatureFlag = 0x00000008
	No_Start               FeatureFlag = 0x00000010 /* I2C_M_NOSTART */
	Slave                  FeatureFlag = 0x00000020
	SMBUS_Block_Proc_Call  FeatureFlag = 0x00008000 /* SMBus 2.0 */
	SMBUS_Quick            FeatureFlag = 0x00010000
	SMBUS_Read_Byte        FeatureFlag = 0x00020000
	SMBUS_Write_Byte       FeatureFlag = 0x00040000
	SMBUS_Read_Byte_Data   FeatureFlag = 0x00080000
	SMBUS_Write_Byte_Data  FeatureFlag = 0x00100000
	SMBUS_Read_Word_Data   FeatureFlag = 0x00200000
	SMBUS_Write_Word_Data  FeatureFlag = 0x00400000
	SMBUS_Proc_Call        FeatureFlag = 0x00800000
	SMBUS_Read_Block_Data  FeatureFlag = 0x01000000
	SMBUS_Write_Block_Data FeatureFlag = 0x02000000
	SMBUS_Read_I2C_Block   FeatureFlag = 0x04000000 /* I2C-like block xfer with 1-byte register address. */
	SMBUS_Write_I2C_Block  FeatureFlag = 0x08000000
)

func (x FeatureFlag) String() string {
	return map[FeatureFlag]string{
		0:  "I2C",
		1:  "10 Bit Addresses",
		2:  "Protocol Mangling",
		3:  "SMBUS PEC",
		4:  "No Start",
		5:  "Slave",
		15: "SMBUS Block Proc Call",
		16: "SMBUS Quick",
		17: "SMBUS Read Byte",
		18: "SMBUS Write Byte",
		19: "SMBUS Read Byte Data",
		20: "SMBUS Write Byte Data",
		21: "SMBUS Read Word Data",
		22: "SMBUS Write Word Data",
		23: "SMBUS Proc Call",
		24: "SMBUS Read Block Data",
		25: "SMBUS Write Block Data",
		26: "SMBUS Read I2C Block",
		27: "SMBUS Write I2C Block",
	}[x]
}

/* SMBus transaction types (size parameter in the above functions)
   Note: these no longer correspond to the (arbitrary) PIIX4 internal codes! */
type SMBusSize uint32

const (
	// This sends a single bit to the device, at the place of the Rd/Wr bit.
	Quick SMBusSize = 0

	// This reads a single byte from a device, without specifying a device register.
	Byte SMBusSize = 1

	// This reads a single byte from a device, from a designated register.
	// The register is specified through the Comm byte.
	ByteData SMBusSize = 2

	// As above but 2 bytes of data.
	WordData SMBusSize = 3

	// This command selects a device register (through the Comm byte), sends
	// 16 bits of data to it, and reads 16 bits of data in return.
	ProcCall SMBusSize = 4

	// This command reads/writes a block of up to 32 bytes from a device, from a
	// designated register that is specified through the Comm byte. The amount
	// of data is specified by first byte of data.
	BlockData SMBusSize = 5

	I2CBlockBroken SMBusSize = 6

	// This command selects a device register (through the Comm byte), sends
	// 1 to 31 bytes of data to it, and reads 1 to 31 bytes of data in return.
	// SMBus 2.0
	BlockProcCall SMBusSize = 7

	I2CBlockData SMBusSize = 8
)

type RW int

const (
	Write RW = iota
	Read
)

const BlockMax = 32

// Byte_Data
type SMBusData [BlockMax + 2]byte

func (b *Bus) Do(rw RW, command uint8, size SMBusSize, data *SMBusData) (err error) {
	err = b.ReadWrite(rw, command, size, data)
	tag := "write"
	if rw == Read {
		tag = "read"
	}
	err = chk(tag, err)
	return
}

func (b *Bus) ReadWrite(rw RW, command uint8, size SMBusSize, data *SMBusData) (err error) {
	var zero SMBusData
	if data == nil {
		data = &zero
	}

	type smbus_cmd struct {
		// 0 => write, 1 => read
		isRead uint8
		// Command byte, a data byte which often selects a register on the device.
		command uint8
		size    SMBusSize
		data    *SMBusData
	}

	cmd := smbus_cmd{
		command: command,
		size:    size,
		data:    data,
	}
	if rw == Read {
		cmd.isRead = 1
	}
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(b.fd), uintptr(I2C_SMBUS), uintptr(unsafe.Pointer(&cmd)))
	if e != 0 {
		err = e
	}
	return
}

func (b *Bus) Read(cmd uint8, size SMBusSize, data *SMBusData) (err error) {
	return b.Do(Read, cmd, size, data)
}

func (bus *Bus) ReadBlock(offset, n int, delay time.Duration) ([]byte, error) {
	// FIXME this should use BlockData
	var err error
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		var data SMBusData
		data[0] = uint8(offset & 0x00ff)
		err = bus.Do(Write, uint8(offset>>8), ByteData, &data)
		if err != nil {
			break
		}
		time.Sleep(delay)
		err = bus.Do(Read, 0, Byte, &data)
		if err != nil {
			break
		}
		buf[i] = data[0]
		offset++
	}
	return buf, err
}

func (b *Bus) Write(cmd uint8, size SMBusSize, data *SMBusData) (err error) {
	return b.Do(Write, cmd, size, data)
}

type MessageFlags uint16

const (
	ReadData     MessageFlags = 0x0001 /* read data, from slave to master; else write */
	TenBit       MessageFlags = 0x0010 /* this is a ten bit chip address */
	Stop         MessageFlags = 0x8000 /* if feature Protocol_Mangling */
	NoStart      MessageFlags = 0x4000 /* if feature No_Start */
	Rev_Dir_Addr MessageFlags = 0x2000 /* if feature Protocol_Mangling */
	Ignore_NAK   MessageFlags = 0x1000 /* if feature Protocol_Mangling */
	No_Read_ACK  MessageFlags = 0x0800 /* if feature Protocol_Mangling */
	Recv_Len     MessageFlags = 0x0400 /* length will be first received byte */
)

type Message struct {
	Address uint16
	Flags   MessageFlags
	Data    []byte
}

const SendMaxMsgs = 42

func (b *Bus) Send(messages []Message) (err error) {
	type msg struct {
		address uint16
		flags   MessageFlags
		len     uint16
		data    unsafe.Pointer
	}
	var ms [SendMaxMsgs]msg
	l := len(messages)
	if l > len(ms) {
		return fmt.Errorf("too many messages: max %d", len(ms))
	}

	for i := 0; i < l; i++ {
		ms[i].address = messages[i].Address
		ms[i].flags = messages[i].Flags
		ms[i].len = uint16(len(messages[i].Data))
		ms[i].data = unsafe.Pointer(&messages[i].Data[0])
	}

	type data struct {
		msgs  unsafe.Pointer
		nmsgs uint32
	}
	d := data{msgs: unsafe.Pointer(&ms[0]), nmsgs: uint32(l)}
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(b.fd), uintptr(I2C_RDWR), uintptr(unsafe.Pointer(&d)))
	if e != 0 {
		err = e
	}
	return
}
