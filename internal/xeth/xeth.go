/* XETH driver sideband control.
 *
 * Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

const DefaultSizeofCh = 4
const netname = "unixpacket"

type SizeofRxchOpt int
type SizeofTxchOpt int
type DialOpt bool

// This provides a buffered interface to an XETH driver side band channel.  Set
// and ExceptionFrame operations are queued to a channel of configurable depth.
// A service go-routine de-queues this channel and, if necessary, dials the
// respctive socket then starts a receive go-routine.
type Xeth struct {
	name string
	addr *net.UnixAddr
	sock *net.UnixConn

	RxCh <-chan []byte
	rxch chan<- []byte
	TxCh chan<- []byte
	txch <-chan []byte

	sockch chan *net.UnixConn
}

// New(driver[, options...]]])
// driver :: XETH driver name (e.g. "platina-mk1")
// Options:
//	SizeofRxchOpt	override DefaultSizeofCh for rxch
//			could be minimal if xeth.Rx() from another go-routine
//	SizeofTxchOpt	override DefaultSizeofCh for txch
//			for maximum buffering, should be
//				number of devices * number of stats
//	DialOpt		if false, don't dial until Assert()
func New(driver string, opts ...interface{}) (*Xeth, error) {
	addr, err := net.ResolveUnixAddr(netname,
		fmt.Sprintf("@%s/xeth", driver))
	if err != nil {
		return nil, err
	}
	sizeofTxch := DefaultSizeofCh
	sizeofRxch := DefaultSizeofCh
	shouldDial := true
	for _, opt := range opts {
		switch t := opt.(type) {
		case SizeofRxchOpt:
			sizeofRxch = int(t)
		case SizeofTxchOpt:
			sizeofTxch = int(t)
		case DialOpt:
			shouldDial = bool(t)
		}
	}
	rxch := make(chan []byte, sizeofRxch)
	txch := make(chan []byte, sizeofTxch)
	xeth := &Xeth{
		name: driver,
		addr: addr,
		RxCh: rxch,
		rxch: rxch,
		TxCh: txch,
		txch: txch,

		sockch: make(chan *net.UnixConn),
	}
	if shouldDial {
		for {
			err = xeth.dial()
			if operr, ok := err.(*net.OpError); ok {
				fmt.Println("OpError:", operr)
				if operr.Timeout() {
					continue
				}
			}
			break
		}
	}
	return xeth, err
}

func (xeth *Xeth) String() string { return xeth.name }

func (xeth *Xeth) Close() error {
	if xeth.sock == nil {
		return nil
	}
	close(xeth.TxCh)
	for _ = range xeth.RxCh {
		// txgo closes sockch after sock shutdown
		// rxgo closes rxch after sockch close
	}
	return nil
}

func (xeth *Xeth) EthtoolDump() {
	buf := Pool.Get(SizeofBreakMsg)
	msg := (*BreakMsg)(unsafe.Pointer(&buf[0]))
	msg.Hdr.Op = uint8(XETH_ETHTOOL_DUMP_OP)
	xeth.TxCh <- buf
}

func (xeth *Xeth) ExceptionFrame(buf []byte) error {
	b := Pool.Get(len(buf))
	copy(b, buf)
	xeth.TxCh <- b
	return nil
}

func (xeth *Xeth) SetStat(ifname, stat string, count uint64) error {
	var statindex uint64
	var op uint8
	if linkstat, found := LinkStatOf(stat); found {
		op = uint8(XETH_LINK_STAT_OP)
		statindex = uint64(linkstat)
	} else if ethtoolstat, found := EthtoolStatOf(stat); found {
		op = uint8(XETH_ETHTOOL_STAT_OP)
		statindex = uint64(ethtoolstat)
	} else {
		return fmt.Errorf("%q unknown", stat)
	}
	buf := Pool.Get(SizeofStatMsg)
	msg := (*StatMsg)(unsafe.Pointer(&buf[0]))
	msg.Hdr.Op = op
	copy(msg.Ifname[:], ifname)
	msg.Stat.Index = statindex
	msg.Stat.Count = count
	xeth.TxCh <- buf
	return nil
}

func (xeth *Xeth) UntilBreak(f func([]byte) error) error {
	breakhdr := make([]byte, SizeofMsgHdr)
	for buf := range xeth.RxCh {
		if bytes.Compare(buf, breakhdr) == 0 {
			Pool.Put(buf)
			break
		}
		err := f(buf)
		Pool.Put(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

// panic if Xeth sock dial fails
func (xeth *Xeth) Assert() {
	if xeth.sock == nil {
		if err := xeth.dial(); err != nil {
			panic(err)
		}
	}
}

func IsEAGAIN(err error) bool {
	if operr, ok := err.(*net.OpError); ok {
		if oserr, ok := operr.Err.(*os.SyscallError); ok {
			if oserr.Err == syscall.EAGAIN {
				return true
			}
		}
	}
	return false
}

func (xeth *Xeth) dial() error {
	for {
		sock, err := net.DialUnix(netname, nil, xeth.addr)
		if err == nil {
			xeth.sock = sock
			break
		}
		if !IsEAGAIN(err) {
			return err
		}
	}
	go xeth.rxgo()
	go xeth.txgo()
	xeth.sockch <- xeth.sock
	return nil
}

func (xeth *Xeth) rxgo() {
	buf := Pool.Get(PageSize)
	oob := Pool.Get(PageSize)
	defer Pool.Put(buf)
	defer Pool.Put(oob)
	defer close(xeth.rxch)
	for sock := range xeth.sockch {
		for {
			n, noob, flags, addr, err := sock.ReadMsgUnix(buf, oob)
			_ = noob
			_ = flags
			_ = addr
			if n == 0 {
				break
			}
			if err != nil {
				if xeth.sock != nil {
					fmt.Fprint(os.Stderr, xeth.name, ": ",
						err, "\n")
				}
				break
			}
			msg := Pool.Get(n)
			copy(msg, buf[:n])
			xeth.rxch <- msg
		}
	}
}

func (xeth *Xeth) shutdown() {
	const (
		SHUT_RD = iota
		SHUT_WR
		SHUT_RDWR
	)
	if xeth.sock != nil {
		if f, err := xeth.sock.File(); err == nil {
			syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
		}
		xeth.sock.Close()
		xeth.sock = nil
	}
}

func (xeth *Xeth) txgo() {
	var err error
	defer close(xeth.sockch)
	defer xeth.shutdown()
	oob := []byte{}
	for buf := range xeth.txch {
		if xeth.sock == nil {
			xeth.sock, err = net.DialUnix(netname, nil, xeth.addr)
			if err != nil {
				Pool.Put(buf)
				continue
			}
		}
		_, _, err = xeth.sock.WriteMsgUnix(buf, oob, nil)
		Pool.Put(buf)
		if err != nil {
			xeth.shutdown()
		}
	}
}
