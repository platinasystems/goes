// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hdelta

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "hdelta"

type cmd struct {
	conn redigo.Conn
	k    string
	m    map[string]string
	t    time.Time
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.DontFork }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return "hdelta [-i SECONDS] [KEY]" }

func (c *cmd) Close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()
	}
	return
}

func (c *cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-i")
	var (
		interval int
		err      error
	)
	if len(parm["-i"]) > 0 {
		_, err := fmt.Sscan(parm["-i"], &interval)
		if err != nil {
			return err
		}
	}
	switch len(args) {
	case 0:
		args = []string{redis.DefaultHash}
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	if c.conn == nil {
		c.conn, err = redis.Connect()
		if err != nil {
			return err
		}
	}
	now := time.Now()
	if c.m == nil {
		c.k = args[0]
		c.t = now
		c.m = make(map[string]string)
	} else if c.k != args[0] {
		c.k = args[0]
		c.t = now
		for k := range c.m {
			delete(c.m, k)
		}
	}
	err = c.hdelta(args[0], now)
	if err != nil || interval == 0 {
		return err
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	t := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-sig:
			t.Stop()
			signal.Stop(sig)
			return nil
		case <-t.C:
			fmt.Println("---")
			c.hdelta(args[0], time.Now())
		}
	}
	return nil
}

func (c *cmd) hdelta(key string, now time.Time) error {
	ret, err := c.conn.Do("HGETALL", key)
	if err != nil {
		return err
	}
	list := ret.([]interface{})
	sec := now.Sub(c.t).Seconds()
	for i := 0; i < len(list); i += 2 {
		var s string
		k := string(list[i].([]byte))
		if list[i+1] != nil {
			s = string(list[i+1].([]byte))
		}
		if c.m[k] != s {
			var newf, oldf float64
			fmt.Print(redis.Quotes(k), ": ")
			if strings.ContainsAny(s, " \t\n") {
				fmt.Println(redis.Quotes(s))
			} else if n, _ := fmt.Sscan(s, &newf); n == 1 {
				fmt.Sscan(c.m[k], &oldf)
				delta := newf - oldf
				if now == c.t {
					fmt.Println(delta)
				} else {
					fmt.Printf("%g (%.3g/s)\n",
						delta, delta/sec)
				}
			} else {
				fmt.Println(redis.Quotes(s))
			}
			c.m[k] = s
		}
	}
	c.t = now
	return nil
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print the changed fields of a redis hash",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	hdelta - print the changed fields of a redis hash

SYNOPSIS
	hdeta [-i SECONDS] [KEY]

DESCRIPTION

	Print the redis hash fields that change between invocations. If the
	field value is an int or float, this prints the difference between
	value followed by the delta divided by the seconds since last
	invocation.

OPTIONS

	-i	iterates hdelta every given second.`,
	}
}
