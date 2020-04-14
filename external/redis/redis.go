// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package redis provides an interface to query and modify a local server.
package redis

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/garyburd/redigo/redis"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/external/redis/rpc/args"
)

const rdtimeout = 10 * time.Second
const wrtimeout = 500 * time.Millisecond

var DefaultHash string
var keyRe *regexp.Regexp
var empty = struct{}{}

func Fprintln(w io.Writer, s string) {
	s = Quotes(s)
	if len(s) == 0 {
		return
	}
	w.Write([]byte(s))
	if s[len(s)-1] != '\n' {
		w.Write([]byte{'\n'})
	}
}

func Quotes(s string) string {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}

func IsReady() error {
	return Hwait(DefaultHash, "redis.ready", "true", 10*time.Second)
}

// Complete redis Key and Subkey. This skips over leading '-' prefaced flags.
func Complete(args ...string) (c []string) {
	if len(args) != 0 && strings.HasPrefix(args[0], "-") {
		args = args[1:]
	}
	switch len(args) {
	case 0:
		c, _ = Keys(".*")
	case 1:
		c, _ = Keys(args[0] + ".*")
	case 2:
		subkeys, _ := Hkeys(args[0])
		for _, subkey := range subkeys {
			if strings.HasPrefix(subkey, args[1]) {
				c = append(c, subkey)
			}
		}
	}
	return
}

// Split a structured redis key.
//
//	"eth0" -> ["eth0"]
//	"eth0.mtu" -> ["eth0", "mtu"]
//	"eth0.addr.[10.0.2.15]" -> ["eth0", "addr", "10.0.2.15"]
//	"eth0.addr.[10.0.2.15].broadcast" ->
//		["eth0", "addr", "10.0.2.15", "broadcast"]
func Split(key string) []string {
	if keyRe == nil {
		keyRe = regexp.MustCompile("[\\[][^\\]]+[\\]]|[^.]+")
	}
	fields := keyRe.FindAllString(key, -1)
	for i, field := range fields {
		if field[0] == '{' {
			if field[len(field)-1] == '}' {
				fields[i] = field[1 : len(field)-1]
			}
		} else if field[0] == '[' {
			if field[len(field)-1] == ']' {
				fields[i] = field[1 : len(field)-1]
			}
		}
	}
	return fields
}

func NewRedisRegAtSock() (*rpc.Client, error) {
	return atsock.NewRpcClient("redis.reg")
}

func NewRedisdAtSock() (net.Conn, error) {
	return atsock.Dial("redisd")
}

// Assign an RPC handler for the given key.
func Assign(key, sockname, name string) error {
	cl, err := NewRedisRegAtSock()
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Reg.Assign", args.Assign{key, sockname, name}, &empty)
}

// Unassign an RPC handler for the given key.
func Unassign(key string) error {
	cl, err := NewRedisRegAtSock()
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Reg.Unassign", args.Unassign{key}, &empty)
}

// Connect to the redis file socket.
func Connect() (redis.Conn, error) {
	conn, err := NewRedisdAtSock()
	if err != nil {
		return nil, err
	}
	return redis.NewConn(conn, rdtimeout, wrtimeout), nil
}

func Get(key string) (s string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("GET", key)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Hdel(key, field string, fields ...string) (i int, err error) {
	if len(key) == 0 {
		key = DefaultHash
	}
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HDEL", key, field, fields)
	if err == nil {
		i = int(ret.(int64))
	}
	return
}

func Hexists(key, field string) (i int, err error) {
	if len(key) == 0 {
		key = DefaultHash
	}
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HEXISTs", key, field)
	if err == nil {
		i = int(ret.(int64))
	}
	return
}

func Hget(key, field string) (s string, err error) {
	if len(key) == 0 {
		key = DefaultHash
	}
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("HGET", key, field)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Hkeys(key string) (keys []string, err error) {
	if len(key) == 0 {
		key = DefaultHash
	}
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HKEYS", key)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

func Hset(key, field string, v interface{}) (i int, err error) {
	if len(key) == 0 {
		key = DefaultHash
	}
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("HSET", key, field, v)
	if ret != nil && err == nil {
		i = int(ret.(int64))
	}
	return
}

func Keys(pattern string) (keys []string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("KEYS", pattern)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

// Wait for the given (key, field) to have value or anything if value is "".
func Hwait(key, field, value string, dur time.Duration) error {
	const t = 250 * time.Millisecond
	for end := time.Now().Add(dur); time.Now().Before(end); time.Sleep(t) {
		s, err := Hget(key, field)
		if err == nil && len(s) > 0 {
			if len(value) > 0 && s != value {
				err = fmt.Errorf("(%s,%s) is %q instead of %q",
					key, field, s, value)
			}
			return err
		}
	}
	return fmt.Errorf("(%s,%s) timeout", key, field)
}

func Lrange(key string, start, stop int) (keys []string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	ret, err := conn.Do("LRANGE", key, start, stop)
	if ret != nil && err == nil {
		vs := ret.([]interface{})
		keys = make([]string, 0, len(vs))
		for _, v := range vs {
			keys = append(keys, vstring(v))
		}
	}
	return
}

// Publish messages to the named redis channel.  Messages sent through the
// returned channel are forwarded to the redis server until the channel is
// closed.
//
//	pub, err := redis.Publish(NAME[, DEPTH])
//	if err {
//		panic()
//	}
//	defer close(pub)
//	...
//	pub.Print("hello world")
//
// The default message channel depth is 16.
func Publish(name string, depth ...int) (chan<- string, error) {
	if len(depth) == 0 {
		depth = []int{16}
	}
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	ch := make(chan string, depth[0])
	go func(name string, out <-chan string, conn redis.Conn) {
		defer conn.Close()

		for {
			// block until next message
			msg, opened := <-out
			if !opened {
				return
			}
			conn.Send("PUBLISH", name, msg)

		drain: // drain buffer of up to 64 total messages
			for n := 1; n < 64; n++ {
				select {
				case msg, opened = <-out:
					if !opened {
						conn.Do("")
						return
					}
					conn.Send("PUBLISH", name, msg)
				default:
					break drain
				}
			}
			conn.Do("")
		}

	}(name, ch, conn)
	return ch, nil
}

func Set(key string, value interface{}) (s string, err error) {
	conn, err := Connect()
	if err != nil {
		return
	}
	defer conn.Close()
	v, err := conn.Do("SET", key, value)
	if v != nil && err == nil {
		s = vstring(v)
	}
	return
}

func Subscribe(channel string) (psc redis.PubSubConn, err error) {
	conn, err := NewRedisdAtSock()
	if err != nil {
		return
	}
	psc = redis.PubSubConn{redis.NewConn(conn, 0, wrtimeout)}
	err = psc.Subscribe(channel)
	if err != nil {
		psc.Close()
	}
	return
}

func vstring(v interface{}) (s string) {
	type stringer interface {
		String() string
	}
	switch t := v.(type) {
	case []byte:
		s = string(t)
	case string:
		s = t
	default:
		if method, found := t.(stringer); found {
			s = method.String()
		} else {
			s = fmt.Sprint(t)
		}
	}
	return
}
