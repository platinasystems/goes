// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type receiverInfo struct {
	conn    *net.UDPConn
	addr    *net.UDPAddr
	tid     int
	block   uint16
	sendBuf []byte
	recvBuf []byte
	length  int
}

func (r *receiverInfo) WriteTo(w io.Writer) (n int64, err error) {
	binary.BigEndian.PutUint16(r.sendBuf, OpcodeAck)
	for {
		if r.length > 0 {
			l, err := w.Write(r.recvBuf[4:r.length])
			n += int64(l)
			if err != nil {
				r.quit(err)
				return n, err
			}
			if r.length < len(r.recvBuf) {
				r.done()
				return n, nil
			}
		}
		binary.BigEndian.PutUint16(r.sendBuf[2:], r.block)
		r.block++
		l, _, err := r.receivePacket(4)
		if err != nil {
			r.quit(err)
			return n, err
		}
		r.length = l
	}
}

func (r *receiverInfo) receivePacket(l int) (int, *net.UDPAddr, error) {
	i := 0
	for {
		i++
		n, a, err := r.receivePkt(l)
		if _, ok := err.(net.Error); ok && i < Retries {
			time.Sleep(time.Second * time.Duration(1))
			continue
		}
		return n, a, err
	}
}

func (r *receiverInfo) receivePkt(l int) (int, *net.UDPAddr, error) {
	err := r.conn.SetReadDeadline(time.Now().Add(Timeout))
	if err != nil {
		return 0, nil, err
	}
	_, err = r.conn.WriteToUDP(r.sendBuf[:l], r.addr)
	if err != nil {
		return 0, nil, err
	}
	for {
		l, addr, err := r.conn.ReadFromUDP(r.recvBuf)
		if err != nil {
			return 0, nil, err
		}
		if !addr.IP.Equal(r.addr.IP) || (r.tid != 0 && addr.Port != r.tid) {
			continue
		}
		p := r.recvBuf[:l]
		r.tid = addr.Port
		if binary.BigEndian.Uint16(p) == OpcodeData {
			if binary.BigEndian.Uint16(p[2:]) == r.block {
				return l, addr, nil
			}
		}
		if binary.BigEndian.Uint16(p) == OpcodeError {
			return 0, addr, fmt.Errorf("Error receive: %s", p[4:])
		}
	}
}

func (r *receiverInfo) done() error {
	if r.conn == nil {
		return nil
	}
	defer r.conn.Close()
	binary.BigEndian.PutUint16(r.sendBuf[2:4], r.block)
	_, err := r.conn.WriteToUDP(r.sendBuf[:4], r.addr)
	if err != nil {
		return err
	}
	return nil
}

func (r *receiverInfo) quit(err error) error {
	if r.conn == nil {
		return nil
	}
	defer r.conn.Close()
	n := setPktError(r.sendBuf, 1, err.Error())
	_, err = r.conn.WriteToUDP(r.sendBuf[:n], r.addr)
	if err != nil {
		return err
	}
	r.conn = nil
	return nil
}
