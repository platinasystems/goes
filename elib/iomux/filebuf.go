package iomux

import (
	"github.com/platinasystems/go/elib"

	"fmt"
	"io"
	"sync"
	"syscall"
)

type FileReadWriteCloser interface {
	Filer
	io.Writer
	io.Closer
	Read(advance int) []byte
}

type FileBuf struct {
	File
	name               string
	maxReadBytes       uint
	txBufLock          sync.Mutex
	txBuffer, rxBuffer elib.ByteVec
}

func NewFileBuf(fd int, format string, args ...interface{}) *FileBuf {
	return &FileBuf{
		File: File{Fd: fd},
		name: fmt.Sprintf(format, args...),
	}
}

func (f *FileBuf) String() string { return f.name }

func (s *FileBuf) Read(advance int) []byte {
	if advance >= len(s.rxBuffer) {
		s.rxBuffer = s.rxBuffer[:0]
	} else {
		s.rxBuffer = s.rxBuffer[advance:]
	}
	return s.rxBuffer
}

func (s *FileBuf) ReadReady() (err error) {
	i := len(s.rxBuffer)

	if s.maxReadBytes <= 0 {
		s.maxReadBytes = 4 << 10
	}
	s.rxBuffer.Resize(s.maxReadBytes)

	var n int
	n, err = syscall.Read(s.Fd, s.rxBuffer[i:])
	if n < 0 {
		n = 0
	}
	s.rxBuffer = s.rxBuffer[:i+n]
	if err != nil {
		switch err {
		case syscall.EAGAIN:
			err = nil
			return
		}
		err = tst(err, "read")
		return
	}
	if n == 0 {
		s.Close()
	}
	return
}

func (s *FileBuf) Write(p []byte) (n int, err error) {
	s.txBufLock.Lock()
	defer s.txBufLock.Unlock()
	i := len(s.txBuffer)
	n = len(p)
	if n > 0 {
		s.txBuffer.Resize(uint(n))
		copy(s.txBuffer[i:i+n], p)
		Update(s)
	}
	return
}

func (s *FileBuf) TxBuf() elib.ByteVec { return s.txBuffer }
func (s *FileBuf) TxLen() int          { return len(s.txBuffer) }

func (s *FileBuf) WriteAvailable() bool { return len(s.txBuffer) > 0 }

func (s *FileBuf) WriteReady() (err error) {
	s.txBufLock.Lock()

	if s.Fd < 0 { // closed?
		s.txBufLock.Unlock()
		return
	}

	needUpdate := false
	defer func() {
		s.txBufLock.Unlock()
		if needUpdate {
			Update(s)
		}
	}()

	if len(s.txBuffer) > 0 {
		var n int
		n, err = syscall.Write(s.Fd, s.txBuffer)
		if err != nil {
			switch err {
			case syscall.EAGAIN:
				err = nil
				return
			}
			err = tst(err, "write")
			return
		}
		l := len(s.txBuffer)
		switch {
		case n == l:
			s.txBuffer = s.txBuffer[:0]
		case n > 0:
			copy(s.txBuffer, s.txBuffer[n:])
			s.txBuffer = s.txBuffer[:l-n]
		}
		// Whole buffer written => toggle write available.
		needUpdate = true
	}
	return
}

func (s *FileBuf) ErrorReady() error {
	// FIXME
	panic("error")
	return nil
}

func (s *FileBuf) Close() (err error) {
	s.txBufLock.Lock()
	defer s.txBufLock.Unlock()
	Del(s)
	err = syscall.Close(s.Fd)
	if err != nil {
		err = fmt.Errorf("close: %s", err)
	}
	s.Fd = -1
	return
}

func tst(err error, tag string) error {
	if err != nil {
		err = fmt.Errorf("%s %s", tag, err)
	}
	return err
}
