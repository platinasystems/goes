// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build linux,amd64

package biosupdate

import (
	"fmt"
	"github.com/platinasystems/elib/hw"
	"io/ioutil"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var sysBusPciPath string = "/sys/bus/pci/devices"

const (
	intelC244ChipsetVendorId uint16 = 0x8086
	intelC244ChipsetDeviceId uint16 = 0x8c54

	ich9RegHSFS uintptr = 0x04

	ich9RegHSFC uintptr = 0x06

	ich9RegFAddr  uintptr = 0x08
	ich9RegFData0 uintptr = 0x10

	ich9RegFrap  uintptr = 0x50
	ich9RegFreg0 uintptr = 0x54
)

const (
	hsfsFDoneOff  uint16 = 0x00
	hsfsFDone     uint16 = (0x01 << hsfsFDoneOff)
	hsfsFCErrOff  uint16 = 0x01
	hsfsFCErr     uint16 = (0x01 << hsfsFCErrOff)
	hsfsBEraseOff uint16 = 0x03
	hsfsBErase    uint16 = (0x03 << hsfsBEraseOff)
)

const (
	hsfcFgoOff    uint16 = 0x00
	hsfcFgo       uint16 = (0x01 << hsfcFgoOff)
	hsfcFCycleOff uint16 = 0x01
	hsfcFCycle    uint16 = (0x03 << hsfcFCycleOff)
	hsfcFDBCOff   uint16 = 0x08
	hsfcFDBC      uint16 = (0x3f << hsfcFDBCOff)
	hsfcSmeOff    uint16 = 0x0F
	hsfcSme       uint16 = (0x01 << hsfcSmeOff)
)

const (
	sys_iopl   = 172 //amd64
	sys_ioperm = 173 //amd64
)

type Programmer struct {
	rcra          uintptr
	rcrb_mem_file *os.File
	rcrb_virt_mem []byte
	rcrb          unsafe.Pointer
	spibar        unsafe.Pointer

	biosBase, biosLimit uint32
	biosImageSize       uint32
	totalFlashSize      uint32

	currentSPINum uint
}

func (p *Programmer) BiosBase() uint32       { return p.biosBase }
func (p *Programmer) BiosLimit() uint32      { return p.biosLimit }
func (p *Programmer) BiosImageSize() uint32  { return p.biosImageSize }
func (p *Programmer) TotalFlashSize() uint32 { return p.totalFlashSize }
func (p *Programmer) CurrentSPINum() uint    { return p.currentSPINum }

func (p *Programmer) spiBarReadReg32(offset uintptr) uint32 {
	//	return *(*uint32)(unsafe.Pointer(uintptr(p.spibar)+offset))
	return hw.LoadUint32(uintptr(p.spibar) + offset)
}

func (p *Programmer) spiBarReadReg16(offset uintptr) uint16 {
	//	return *(*uint16)(unsafe.Pointer(uintptr(p.spibar)+offset))
	return hw.LoadUint16(uintptr(p.spibar) + offset)
}

func (p *Programmer) spiBarWriteReg32(offset uintptr, d uint32) {
	//	*(*uint32)(unsafe.Pointer(uintptr(p.spibar)+offset)) = d
	hw.StoreUint32(uintptr(p.spibar)+offset, d)
}

func (p *Programmer) spiBarWriteReg16(offset uintptr, d uint16) {
	//	*(*uint16)(unsafe.Pointer(uintptr(p.spibar)+offset)) = d
	hw.StoreUint16(uintptr(p.spibar)+offset, d)
}

func ichFregBase(flreg uint32) uint32 {
	return ((flreg) << 12) & 0x01fff000
}

func ichFregLimit(flreg uint32) uint32 {
	return (((flreg) >> 4) & 0x01fff000) | 0xfff
}

func FindProgrammerRCRA() (rcra uintptr, err error) {
	fis, err := ioutil.ReadDir(sysBusPciPath)
	if err != nil {
		return
	}

	for _, fi := range fis {
		config_file_name := sysBusPciPath + "/" + fi.Name() + "/config"

		found := func() bool {
			f, ferr := os.OpenFile(config_file_name, os.O_RDONLY, 0)
			if ferr != nil {
				return false
			}

			defer f.Close()

			var buf [4]byte
			var vendor_id, device_id uint16

			// Read and check Vendor ID
			if _, ferr = f.Seek(0, os.SEEK_SET); ferr != nil {
				return false
			}

			if _, ferr = f.Read(buf[:2]); ferr != nil {
				return false
			}

			vendor_id = uint16(buf[0]) + (uint16(buf[1]) << 8)
			if vendor_id != intelC244ChipsetVendorId {
				return false
			}

			// Read and check Device ID
			if _, ferr = f.Seek(2, os.SEEK_SET); ferr != nil {
				return false
			}

			if _, ferr = f.Read(buf[:2]); ferr != nil {
				return false
			}

			device_id = uint16(buf[0]) + (uint16(buf[1]) << 8)
			if device_id != intelC244ChipsetDeviceId {
				return false
			}

			// Read and return RCRA
			if _, ferr = f.Seek(0xF0, os.SEEK_SET); ferr != nil {
				return false
			}

			if _, ferr = f.Read(buf[:4]); ferr != nil {
				return false
			}

			rcra = uintptr(buf[0]) + (uintptr(buf[1]) << 8)
			rcra += (uintptr(buf[2]) << 16) + (uintptr(buf[3]) << 24)

			return true
		}()

		if found {
			return
		}
	}

	err = fmt.Errorf("Unable to find Intel C224 Chipset Device.")
	return
}

func (p *Programmer) Open(spinum uint) (err error) {
	defer func() {
		if err != nil {
			p.Close()
		}
	}()

	err = p.selectSPI(spinum)
	if err != nil {
		return
	}

	rcra, err := FindProgrammerRCRA()
	if err != nil {
		return
	}

	p.rcra = rcra & uintptr(0xffffc000)

	p.rcrb_mem_file, err = os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return
	}

	/* TODO: Adjust rcra and length to 0x1000 page size */

	p.rcrb_virt_mem, err = syscall.Mmap(int(p.rcrb_mem_file.Fd()), int64(p.rcra), int(0x4000),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return
	}

	// Find RCRB and SPIBAR
	p.rcrb = unsafe.Pointer(&p.rcrb_virt_mem[0])
	p.spibar = unsafe.Pointer(uintptr(p.rcrb) + uintptr(0x3800))

	// Read freg and determine BIOS Base and Limit in SPI
	freg := p.spiBarReadReg32(ich9RegFreg0 + 4)
	p.biosBase = ichFregBase(freg)
	p.biosLimit = ichFregLimit(freg)
	p.biosImageSize = p.biosLimit - p.biosBase + 1

	/* TODO: Find total flash size */

	p.totalFlashSize = 16384 * 1024

	return
}

func (p *Programmer) Close() (err error) {
	if p.rcrb_mem_file != nil {
		p.rcrb_mem_file.Close()
		p.rcrb_mem_file = nil
	}

	if p.rcrb_virt_mem != nil {
		syscall.Munmap(p.rcrb_virt_mem)
		p.rcrb_mem_file.Close()
		p.rcrb_virt_mem = nil
	}

	// Always select SPI0 and enable watchdog
	if p.currentSPINum != 0 {
		err = p.selectSPI(0)
	}

	return
}

func (p *Programmer) ReadAt(b []byte, off int64, c chan float32) (n int, err error) {
	// clear FDONE, FCERR, AEL by writing 1 to them (if they are set)
	p.spiBarWriteReg16(ich9RegHSFS, p.spiBarReadReg16(ich9RegHSFS))

	var blocklen = 64
	n = 0
	l := len(b)

	for l > 0 {
		if blocklen > l {
			blocklen = l
		}

		page := 256 - int(off&0xFF)
		if blocklen > page {
			blocklen = page
		}

		if c != nil {
			c <- (float32(n) / float32(len(b))) * 100
		}

		p.ichSetAddress(uint32(off))

		hsfc := p.spiBarReadReg16(ich9RegHSFC)
		// Set Read Operation and clear byte count
		hsfc &= ^hsfcFCycle
		hsfc &= ^hsfcFDBC

		// Set byte Count
		hsfc |= (uint16(blocklen-1) << hsfcFDBCOff) & hsfcFDBC

		// Start
		hsfc |= hsfcFgo
		p.spiBarWriteReg16(ich9RegHSFC, hsfc)

		err = p.ichWaitForOperationToComplete("read", 5*time.Second)
		if err != nil {
			return
		}

		p.ichReadData(b[n:(n+blocklen)], ich9RegFData0)
		n += blocklen
		l -= blocklen
		off += int64(blocklen)
	}

	return
}

func (p *Programmer) EraseFlash(addr, size uint32, c chan float32) (err error) {
	blocksize, err := p.getEraseBlockSize(addr)

	if (addr % blocksize) != 0 {
		return fmt.Errorf("Erase address not aligned.\n")
	}

	if (size % blocksize) != 0 {
		return fmt.Errorf("Erase length not aligned.\n")
	}

	var a float32
	d := float32(blocksize) / float32(size)

	for size > 0 {

		if c != nil {
			c <- (a) * 100
		}

		err = p.eraseFlashBlock(addr, blocksize)
		if err != nil {
			return
		}
		addr += blocksize
		size -= blocksize
		a += d
	}

	return
}

func (p *Programmer) CheckErased(base, size uint32, c chan float32) (erased bool, err error) {

	d := make([]byte, size)
	n, err := p.ReadAt(d, int64(base), c)
	if err != nil {
		return
	}
	if uint32(n) != size {
		return false, fmt.Errorf("%d bytes read, expecting %d bytes", n, size)
	}

	for i := range d {
		if d[i] != 0xFF {
			return false, nil
		}
	}

	return true, nil
}

func (p *Programmer) WriteAt(b []byte, off int64, c chan float32) (n int, err error) {
	// clear FDONE, FCERR, AEL by writing 1 to them (if they are set)
	p.spiBarWriteReg16(ich9RegHSFS, p.spiBarReadReg16(ich9RegHSFS))

	var blocklen = 64
	n = 0
	l := len(b)

	for l > 0 {
		if blocklen > l {
			blocklen = l
		}

		page := 256 - int(off&0xFF)
		if blocklen > page {
			blocklen = page
		}

		if c != nil {
			c <- (float32(n) / float32(len(b))) * 100
		}

		p.ichSetAddress(uint32(off))

		p.ichFillData(b[n:(n+blocklen)], ich9RegFData0)

		hsfc := p.spiBarReadReg16(ich9RegHSFC)
		// Set Write Operation and clear byte count
		hsfc &= ^hsfcFCycle
		hsfc |= (0x02 << hsfcFCycleOff)
		hsfc &= ^hsfcFDBC

		// Set byte Count
		hsfc |= (uint16(blocklen-1) << hsfcFDBCOff) & hsfcFDBC

		// Start
		hsfc |= hsfcFgo
		p.spiBarWriteReg16(ich9RegHSFC, hsfc)

		err = p.ichWaitForOperationToComplete("write", 5*time.Second)
		if err != nil {
			return
		}

		n += blocklen
		l -= blocklen
		off += int64(blocklen)
	}

	return
}

func (p *Programmer) ichReadData(b []byte, reg0 uintptr) {
	var t uint32

	for i := range b {
		if i%4 == 0 {
			t = p.spiBarReadReg32(reg0 + uintptr(i))
		}
		b[i] = byte((t >> (uint(i%4) * 8)) & 0xFF)
	}
}

func (p *Programmer) ichFillData(b []byte, reg0 uintptr) {
	var t uint32

	if len(b) == 0 {
		return
	}

	var i int
	for i = range b {
		if i%4 == 0 {
			t = 0
		}

		t |= (uint32(b[i]) << (uint(i%4) * 8))
		if (i % 4) == 3 {
			p.spiBarWriteReg32(reg0+uintptr(i-(i%4)), t)
		}
	}
	i--
	// Write remaining data
	if (i % 4) == 3 {
		p.spiBarWriteReg32(reg0+uintptr(i-(i%4)), t)
	}
}

func (p *Programmer) ichSetAddress(a uint32) {
	v := p.spiBarReadReg32(ich9RegFAddr) & ^uint32(0x01FFFFFF)
	v = (a & 0x01FFFFFF) | v
	p.spiBarWriteReg32(ich9RegFAddr, v)
}

func (p *Programmer) ichWaitForOperationToComplete(opname string, d time.Duration) (err error) {
	var hsfs uint16
	timedout := false

	timeout := time.After(d)

wait:
	for {
		select {
		case <-timeout:
			timedout = true
			break wait
		default:
			hsfs = p.spiBarReadReg16(ich9RegHSFS)
			if (hsfs & (hsfsFDone | hsfsFCErr)) != 0 {
				break wait
			}
		}
	}

	// Clear FDONE, FCERR, AEL by writing 1 to them (if they are set)
	p.spiBarWriteReg16(ich9RegHSFS, p.spiBarReadReg16(ich9RegHSFS))

	if (hsfs & hsfsFCErr) != 0 {
		addr := p.spiBarReadReg32(ich9RegFAddr) & 0x01FFFFFF
		return fmt.Errorf("Transaction error @ %08X during %s.", addr, opname)
	}

	if timedout {
		addr := p.spiBarReadReg32(ich9RegFAddr) & 0x01FFFFFF
		return fmt.Errorf("Transaction timed out @ %08X during %s.", addr, opname)
	}

	return
}

func (p *Programmer) getEraseBlockSize(addr uint32) (size uint32, err error) {
	decBErase := []uint32{256, 4 * 1024, 8 * 1024, 64 * 1024}

	p.ichSetAddress(addr)

	encBErase := (p.spiBarReadReg16(ich9RegHSFS) & hsfsBErase) >> hsfsBEraseOff

	return decBErase[encBErase], nil
}

func (p *Programmer) eraseFlashBlock(addr, size uint32) (err error) {
	blocksize, err := p.getEraseBlockSize(addr)

	if (addr % blocksize) != 0 {
		return fmt.Errorf("Erase address not aligned.\n")
	}

	if size != blocksize {
		return fmt.Errorf("Erase length not aligned.\n")
	}

	p.ichSetAddress(uint32(addr))

	// clear FDONE, FCERR, AEL by writing 1 to them (if they are set)
	p.spiBarWriteReg16(ich9RegHSFS, p.spiBarReadReg16(ich9RegHSFS))

	hsfc := p.spiBarReadReg16(ich9RegHSFC)
	// Set Erase Operation
	hsfc &= ^hsfcFCycle
	hsfc |= (0x03 << hsfcFCycleOff)

	// Start
	hsfc |= hsfcFgo
	p.spiBarWriteReg16(ich9RegHSFC, hsfc)

	err = p.ichWaitForOperationToComplete("erase", 5*time.Second)
	if err != nil {
		return
	}

	return
}

func (p *Programmer) selectSPI(spinum uint) (err error) {
	currentSPINum, err := GetSelectedSPI()
	if err != nil {
		return
	}

	if currentSPINum == spinum {
		return
	}

	if err = SelectSPI(spinum); err != nil {
		return
	}

	timedout := false
	timeout := time.After(5 * time.Second)

wait:
	for {
		select {
		case <-timeout:
			timedout = true
			break wait
		default:
			if currentSPINum, err = GetSelectedSPI(); err != nil {
				return
			}

			if currentSPINum == spinum {
				break wait
			}
		}
	}

	if timedout {
		return fmt.Errorf("Unable to select SPI%d\n", spinum)
	}

	p.currentSPINum = spinum

	return
}

func GetSelectedSPI() (spinum uint, err error) {
	// Dump PLD registers for debug
	if err = DumpPLD(); err != nil {
		return
	}

	b, err := ioReadReg8(0x602)
	if err != nil {
		return
	}

	if b & 0x80 == 0 {
		spinum = 0
	} else {
		spinum = 1
	}

	return
}

func SelectSPI(spinum uint) (err error) {
	b, err := ioReadReg8(0x602)
	if err != nil {
		return
	}

	if spinum > 0 {
		// Disable watchdog recovery
		b &= ^byte(0x01)
		if err = ioWriteReg8(0x604, b); err != nil {
			return
		}

		// Set bios_mux_select
		b |= byte(0x80)
		if err = ioWriteReg8(0x604, b); err != nil {
			return
		}

		return
	}

	// Clear bios_mux_select
	b &= ^byte(0x80)
	if err = ioWriteReg8(0x604, b); err != nil {
		return
	}

	// Enable watchdog recovery
	b |= byte(0x01)
	if err = ioWriteReg8(0x604, b); err != nil {
		return
	}

	return
}

func ioReadReg8(addr uintptr) (b byte, err error) {
	if err = setIoperm(addr); err != nil {
		return
	}

	buf := make([]byte, 1)
	f, err := os.Open("/dev/port")
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = f.Seek(int64(addr), 0); err != nil {
		return
	}

	n, err := f.Read(buf)
	if err != nil {
		return
	}
	f.Close()

	if n != 1 {
		return 0, fmt.Errorf("Unable to read port.")
	}

	if err = clrIoperm(addr); err != nil {
		return
	}

	b = buf[0]

	return
}

func ioWriteReg8(addr uintptr, b byte) (err error) {
	if err = setIoperm(addr); err != nil {
		return
	}

	buf := make([]byte, 1)
	buf[0] = b
	f, err := os.OpenFile("/dev/port", os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = f.Seek(int64(addr), 0); err != nil {
		return err
	}

	n, err := f.Write(buf)
	if err != nil {
		return err
	}

	f.Sync()
	f.Close()

	if n != 1 {
		return fmt.Errorf("Unable to write port.")
	}

	if err = clrIoperm(addr); err != nil {
		return
	}

	return
}

func setIoperm(addr uintptr) error {
	level := 3
	if _, _, errno := syscall.Syscall(sys_iopl,
		uintptr(level), 0, 0); errno != 0 {
		return errno
	}
	num := 1
	on := 1
	if _, _, errno := syscall.Syscall(sys_ioperm, addr,
		uintptr(num), uintptr(on)); errno != 0 {
		return errno
	}

	return nil
}

func clrIoperm(addr uintptr) error {
	num := 1
	on := 0
	if _, _, errno := syscall.Syscall(sys_ioperm, addr,
		uintptr(num), uintptr(on)); errno != 0 {
		return errno
	}
	level := 0
	if _, _, errno := syscall.Syscall(sys_iopl,
		uintptr(level), 0, 0); errno != 0 {
		return errno

	}
	return nil
}
