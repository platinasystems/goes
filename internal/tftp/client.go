// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"
)

const (
	OpcodeRead  = 1
	OpcodeWrite = 2
	OpcodeData  = 3
	OpcodeAck   = 4
	OpcodeError = 5
	Timeout     = 5 * time.Second
	Retries     = 5
	MaxPktSize  = 516
)

type Client struct {
	addr *net.UDPAddr
}

func GetFile(host string, source string, file string) (err error, size int) {
	client, err := dialTFTP(host)
	if err != nil {
		return err, 0
	}
	r, _, err := client.recv(source)
	if err != nil {
		return err, 0
	}
	f, err := os.Create(file)
	if err != nil {
		return err, 0
	}
	defer f.Close()
	s, err := r.WriteTo(f)
	if err != nil {
		return err, 0
	}
	return err, int(s)
}

func GetFileRC(host string) (io.ReadCloser, error) {
	ip := "192.168.101.142" + ":69" //FIXME parse host arg
	source := "downloads/LATEST/LIST"

	client, err := dialTFTP(ip)
	if err != nil {
		return nil, err
	}
	r, l, err := client.recv(source)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if _, err := r.WriteTo(&b); err != nil {
		return nil, err
	}
	rc := ioutil.NopCloser(&b)
	return rc, nil
}

func dialTFTP(addr string) (*Client, error) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("Connect failed to %s, %v", addr, err)
	}
	return &Client{addr: a}, nil
}

func (c Client) recv(filename string) (io.WriterTo, int, error) {
	con, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, 0, err
	}
	r := &receiverInfo{
		conn:    con,
		addr:    c.addr,
		block:   1,
		sendBuf: make([]byte, MaxPktSize),
		recvBuf: make([]byte, MaxPktSize),
	}
	p := setPktRRQ(r.sendBuf, filename)
	l, a, err := r.receivePacket(p)
	if err != nil {
		return nil, 0, err
	}

	r.length = l
	r.addr = a
	return r, l, nil
}

//RRQ,WRQ |  2 bytes   |  string  | 1 byte  | string | 1 byte |
//RRQ,WRQ | opcode 1,2 | filename |  0x0    |  mode  |  0x0   |
//
//Data    |  2 bytes   |  2 bytes | N byte  |
//Data    |  opcode 3  |  block#  |  Data   |
//
//Ack     |  2 bytes   |  2 bytes |
//Ack     |  opcode 4  |  block#  |
//
//Error   |  2 bytes   |  2 bytes | string  | 1 byte |
//Error   |  opcode 5  | err code | err msg |  0x0   |

func setPktRRQ(p []byte, filename string) int {
	binary.BigEndian.PutUint16(p, OpcodeRead)
	copy(p[2:], filename)
	p[(2 + len(filename))] = 0
	copy(p[(3+len(filename)):], "octet")
	p[(8 + len(filename))] = 0
	return (9 + len(filename))
}

func setPktError(p []byte, errorCode uint16, message string) int {
	binary.BigEndian.PutUint16(p, OpcodeError)
	binary.BigEndian.PutUint16(p[2:], errorCode)
	copy(p[4:], message)
	p[(4 + len(message))] = 0
	return (5 + len(message))
}
