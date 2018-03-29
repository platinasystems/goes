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
	"fmt"
	"net"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const DefaultSizeofCh = 4
const netname = "unixpacket"

type AssertDialOpt bool
type SizeofRxchOpt int
type SizeofTxchOpt int

// This provides a buffered interface to an XETH driver side band channel.  Set
// and ExceptionFrame operations are queued to a channel of configurable depth.
// A service go-routine de-queues this channel and, if necessary, dials the
// respctive socket then starts a receive go-routine.
type Xeth struct {
	name string
	txch chan []byte
	rxch chan []byte
	stat *strings.Replacer

	closed bool

	IndexofEthtoolStat map[string]uint64
}

// New(driver, stats[, options...]]])
// driver :: XETH driver name (e.g. "platina-mk1")
// stats :: list of driver's ethtool stat names
// Options:
//	AssertDialOpt	if true, panic if can't dial the driver's
//			side-band socket
//	SizeofRxchOpt	override DefaultSizeofCh for rxch
//			could be minimal if xeth.Rx() from another go-routine
//	SizeofTxchOpt	override DefaultSizeofCh for txch
//			for maximum buffering, should be
//				number of devices * number of stats
func New(driver string, stats []string, opts ...interface{}) (*Xeth, error) {
	addr, err := net.ResolveUnixAddr(netname,
		fmt.Sprintf("@%s.xeth", driver))
	if err != nil {
		return nil, err
	}
	assertDial := false
	sizeofTxch := DefaultSizeofCh
	sizeofRxch := DefaultSizeofCh
	for _, opt := range opts {
		switch t := opt.(type) {
		case AssertDialOpt:
			assertDial = bool(t)
		case SizeofRxchOpt:
			sizeofRxch = int(t)
		case SizeofTxchOpt:
			sizeofTxch = int(t)
		}
	}
	xeth := &Xeth{
		name: driver,
		rxch: make(chan []byte, sizeofRxch),
		txch: make(chan []byte, sizeofTxch),
		stat: strings.NewReplacer(" ", "-", ".", "-", "_", "-"),

		IndexofEthtoolStat: make(map[string]uint64),
	}
	for i, s := range stats {
		xeth.IndexofEthtoolStat[xeth.stat.Replace(s)] = uint64(i)
	}
	runtime.SetFinalizer(xeth, (*Xeth).Close)
	go xeth.txgo(addr, assertDial)
	return xeth, nil
}

func (xeth *Xeth) String() string { return xeth.name }

func (xeth *Xeth) Close() error {
	runtime.SetFinalizer(xeth, nil)
	if xeth.closed {
		return nil
	}
	close(xeth.txch)
	for _ = range xeth.rxch {
		// txgo closes rxch after sock shutdown
	}
	xeth.IndexofEthtoolStat = nil
	return nil
}

func (xeth *Xeth) ExceptionFrame(buf []byte) error {
	b := Pool.Get(len(buf))
	copy(b, buf)
	xeth.txch <- b
	return nil
}

func (xeth *Xeth) Set(ifname, stat string, count uint64) error {
	var statindex uint64
	var found bool
	var op uint8
	modstat := xeth.stat.Replace(stat)
	if statindex, found = IndexofNetStat[modstat]; found {
		op = SbOpSetNetStat
	} else if statindex, found = xeth.IndexofEthtoolStat[modstat]; found {
		op = SbOpSetEthtoolStat
	} else {
		return fmt.Errorf("STAT %q unknown", modstat)
	}
	buf := Pool.Get(SizeofSbHdrSetStat)
	sbhdr := (*SbHdr)(unsafe.Pointer(&buf[0]))
	sbhdr.Op = op
	sbstat := (*SbSetStat)(unsafe.Pointer(&buf[SizeofSbHdr]))
	copy(sbstat.Ifname[:], ifname)
	sbstat.Statindex = statindex
	sbstat.Count = count
	xeth.txch <- buf
	return nil
}

func (xeth *Xeth) Rx(buf []byte) (n int, err error) {
	poolbuf := <-xeth.rxch
	n = len(poolbuf)
	if n > len(buf) {
		return 0, syscall.E2BIG
	}
	buf = buf[:n]
	copy(buf, poolbuf)
	Pool.Put(poolbuf)
	return n, nil
}

func (xeth *Xeth) txgo(addr *net.UnixAddr, assertDial bool) {
	var (
		err  error
		sock *net.UnixConn
	)
	oob := []byte{}
	for buf := range xeth.txch {
		if sock == nil {
			sock, err = net.DialUnix(netname, nil, addr)
			if err != nil {
				if assertDial {
					panic(err)
				}
				Pool.Put(buf)
				continue
			} else {
				go xeth.rxgo(sock)
			}
		}
		_, _, err = sock.WriteMsgUnix(buf, oob, nil)
		Pool.Put(buf)
		if err != nil {
			xeth.shutdown(sock)
			sock = nil
		}
	}
	xeth.shutdown(sock)
	close(xeth.rxch)
}

func (xeth *Xeth) rxgo(sock *net.UnixConn) {
	oob := []byte{}
	for {
		buf := Pool.Get(PageSize)
		n, noob, flags, addr, err := sock.ReadMsgUnix(buf, oob)
		_ = noob
		_ = flags
		_ = addr
		if err != nil || xeth.closed {
			Pool.Put(buf)
			break
		}
		xeth.rxch <- buf[:n]
	}
}

func (xeth *Xeth) shutdown(sock *net.UnixConn) error {
	const (
		SHUT_RD = iota
		SHUT_WR
		SHUT_RDWR
	)
	if sock == nil {
		return nil
	}
	f, err := sock.File()
	if err == nil {
		err = syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
	}
	if cerr := sock.Close(); err == nil {
		err = cerr
	}
	xeth.closed = true
	return err
}
