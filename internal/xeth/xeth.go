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
	"io"
	"net"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const DefaultSizeofCh = 4
const netname = "unixpacket"

// This provides an interface to the XETH driver's side-band socket.
type Xeth struct {
	name string
	addr *net.UnixAddr
	sock *net.UnixConn

	RxCh <-chan []byte
	rxch chan<- []byte
	txch chan []byte
}

var expMsgSz = map[Kind]int{
	XETH_MSG_KIND_ETHTOOL_FLAGS:    SizeofMsgEthtoolFlags,
	XETH_MSG_KIND_ETHTOOL_SETTINGS: SizeofMsgEthtoolSettings,
	XETH_MSG_KIND_IFA:              SizeofMsgIfa,
	XETH_MSG_KIND_IFDEL:            SizeofMsgIfdel,
	XETH_MSG_KIND_IFINFO:           SizeofMsgIfinfo,
	XETH_MSG_KIND_IFVID:            SizeofMsgIfvid,
	XETH_MSG_KIND_NEIGH_UPDATE:     SizeofMsgNeighUpdate,
}

func IsEAGAIN(err error) bool {
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			if oserr, ok := operr.Err.(*os.SyscallError); ok {
				if oserr.Err == syscall.EAGAIN {
					return true
				}
			}
		}
	}
	return false
}

func IsTimeout(err error) bool {
	if err != nil {
		if op, ok := err.(*net.OpError); ok {
			return op.Timeout()
		}
	}
	return false
}

// New(driver)
// driver :: XETH driver name (e.g. "platina-mk1")
func New(driver string) (*Xeth, error) {
	var sock *net.UnixConn
	addr, err := net.ResolveUnixAddr(netname, "@xeth")
	if err != nil {
		return nil, err
	}
	for {
		sock, err = net.DialUnix(netname, nil, addr)
		if err == nil {
			break
		}
		if !IsEAGAIN(err) {
			return nil, err
		}
	}
	rxch := make(chan []byte, 4)
	xeth := &Xeth{
		name: driver,
		addr: addr,
		sock: sock,
		RxCh: rxch,
		rxch: rxch,
		txch: make(chan []byte, 4),
	}
	go xeth.gorx()
	go xeth.gotx()
	return xeth, err
}

func (xeth *Xeth) gorx() {
	const minrxto = 10 * time.Millisecond
	const maxrxto = 320 * time.Millisecond
	rxto := minrxto
	rxbuf := Pool.Get(PageSize)
	defer Pool.Put(rxbuf)
	rxoob := Pool.Get(PageSize)
	defer Pool.Put(rxoob)
	defer close(xeth.rxch)
	for xeth.sock != nil {
		err := xeth.sock.SetReadDeadline(time.Now().Add(rxto))
		if err != nil {
			fmt.Fprintln(os.Stderr, "xeth set rx deadline", err)
			break
		}
		n, noob, flags, addr, err :=
			xeth.sock.ReadMsgUnix(rxbuf, rxoob)
		_ = noob
		_ = flags
		_ = addr
		if n == 0 || IsTimeout(err) {
			if rxto < maxrxto {
				rxto *= 2
			}
		} else if err == nil {
			rxto = minrxto
			msg := Pool.Get(n)
			copy(msg, rxbuf[:n])
			xeth.rxch <- msg
		} else {
			e, ok := err.(*os.SyscallError)
			if !ok || e.Err.Error() != "EOF" {
				fmt.Fprintln(os.Stderr, "xeth rx", err)
			}
			break
		}
	}
}

func (xeth *Xeth) gotx() {
	for msg := range xeth.txch {
		xeth.tx(msg, 10*time.Millisecond)
		Pool.Put(msg)
	}
}

func (xeth *Xeth) tx(buf []byte, timeout time.Duration) error {
	var oob []byte
	var dl time.Time
	if xeth.sock == nil {
		return io.EOF
	}
	if timeout != time.Duration(0) {
		dl = time.Now().Add(timeout)
	}
	err := xeth.sock.SetWriteDeadline(dl)
	if err != nil {
		return err
	}
	_, _, err = xeth.sock.WriteMsgUnix(buf, oob, nil)
	return err
}

func (xeth *Xeth) String() string { return xeth.name }

func (xeth *Xeth) Close() error {
	const (
		SHUT_RD = iota
		SHUT_WR
		SHUT_RDWR
	)
	if xeth.sock == nil {
		return nil
	}
	if f, err := xeth.sock.File(); err == nil {
		syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
	}
	xeth.sock.Close()
	close(xeth.txch)
	xeth.sock = nil
	return nil
}

func (xeth *Xeth) Carrier(ifname string, flag CarrierFlag) error {
	buf := Pool.Get(SizeofMsgStat)
	defer Pool.Put(buf)
	msg := (*MsgCarrier)(unsafe.Pointer(&buf[0]))
	msg.Kind = uint8(XETH_MSG_KIND_CARRIER)
	copy(msg.Ifname[:], ifname)
	msg.Flag = uint8(flag)
	return xeth.tx(buf, 0)
}

func (xeth *Xeth) DumpFib() error {
	buf := Pool.Get(SizeofMsgBreak)
	defer Pool.Put(buf)
	msg := (*MsgBreak)(unsafe.Pointer(&buf[0]))
	msg.Kind = uint8(XETH_MSG_KIND_DUMP_FIBINFO)
	return xeth.tx(buf, 0)
}

func (xeth *Xeth) DumpIfinfo() error {
	buf := Pool.Get(SizeofMsgBreak)
	defer Pool.Put(buf)
	msg := (*MsgBreak)(unsafe.Pointer(&buf[0]))
	msg.Kind = uint8(XETH_MSG_KIND_DUMP_IFINFO)
	return xeth.tx(buf, 0)
}

func (xeth *Xeth) SetStat(ifname, stat string, count uint64) error {
	var statindex uint64
	var kind uint8
	if linkstat, found := LinkStatOf(stat); found {
		kind = uint8(XETH_MSG_KIND_LINK_STAT)
		statindex = uint64(linkstat)
	} else if ethtoolstat, found := EthtoolStatOf(stat); found {
		kind = uint8(XETH_MSG_KIND_ETHTOOL_STAT)
		statindex = uint64(ethtoolstat)
	} else {
		return fmt.Errorf("%q unknown", stat)
	}
	buf := Pool.Get(SizeofMsgStat)
	defer Pool.Put(buf)
	msg := (*MsgStat)(unsafe.Pointer(&buf[0]))
	msg.Kind = kind
	copy(msg.Ifname[:], ifname)
	msg.Index = statindex
	msg.Count = count
	return xeth.tx(buf, 10*time.Millisecond)
}

func (xeth *Xeth) Speed(ifname string, count uint64) error {
	buf := Pool.Get(SizeofMsgSpeed)
	defer Pool.Put(buf)
	msg := (*MsgSpeed)(unsafe.Pointer(&buf[0]))
	msg.Kind = uint8(XETH_MSG_KIND_SPEED)
	copy(msg.Ifname[:], ifname)
	msg.Mbps = uint32(count)
	return xeth.tx(buf, 0)
}

// leaky bucket
func (xeth *Xeth) Tx(buf []byte) {
	msg := Pool.Get(len(buf))
	copy(msg, buf)
	select {
	case xeth.txch <- msg:
	default:
		Pool.Put(msg)
	}
}

func (xeth *Xeth) UntilBreak(f func([]byte) error) error {
	var err error
	for msg := range xeth.RxCh {
		kind := KindOf(msg)
		if kind == XETH_MSG_KIND_BREAK {
			err = io.EOF
		} else if kind != XETH_MSG_KIND_NOT_MSG {
			exp, found := expMsgSz[kind]
			if found && exp != len(msg) {
				err = fmt.Errorf("mismatched %s", kind)
			} else {
				err = f(msg)
			}
		}
		Pool.Put(msg)
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return err
}
