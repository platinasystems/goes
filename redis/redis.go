// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package redis commands to query and modify the local redis server.
package redis

import (
	"fmt"
	"regexp"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/redis/rpc/args"
	"github.com/platinasystems/go/sockfile"
)

const Timeout = 500 * time.Millisecond

var Machine = "platina"

var keyRe *regexp.Regexp
var empty = struct{}{}

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

// Assign an RPC handler for the given key.
func Assign(key, file, name string) error {
	cl, err := sockfile.NewRpcClient("redis-reg")
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Reg.Assign", args.Assign{key, file, name}, &empty)
}

// Unassign an RPC handler for the given key.
func Unassign(key string) error {
	cl, err := sockfile.NewRpcClient("redis-reg")
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Reg.Unassign", args.Unassign{key}, &empty)
}

// Connect to the redis file socket.
func Connect() (redis.Conn, error) {
	conn, err := sockfile.Dial("redisd")
	if err != nil {
		return nil, err
	}
	return redis.NewConn(conn, Timeout, Timeout), nil
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
//	pub, err := redis.Publish(NAME)
//	if err {
//		panic()
//	}
//	defer close(pub)
//	...
//	pub.Print("hello world")
func Publish(name string) (chan<- string, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 16)
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

func Subscribe(channel string) (out <-chan string) {
	var err error
	ch := make(chan string, 4)
	out = ch
	defer func() {
		if err != nil {
			ch <- err.Error()
			close(ch)
		}
	}()
	conn, err := sockfile.Dial("redisd")
	if err != nil {
		return
	}
	psc := redis.PubSubConn{redis.NewConn(conn, 0, Timeout)}
	if err := psc.Subscribe(channel); err != nil {
		return
	}
	go func(psc redis.PubSubConn, in chan<- string) {
		for {
			v := psc.Receive()
			switch t := v.(type) {
			case redis.Message:
				in <- string(t.Data)
			case error:
				in <- t.Error()
				close(in)
				return
			}
		}
	}(psc, ch)
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
