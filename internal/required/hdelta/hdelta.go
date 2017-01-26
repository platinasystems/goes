// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hdelta

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "hdelta"
const sep = ": "

type cmd bool

type entry struct {
	s string
	t time.Time
}

func New() *cmd { return new(cmd) }

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "hdelta [CHANNEL]" }

func (c *cmd) Close() error {
	*c = true
	return nil
}

func (c *cmd) Main(args ...string) error {
	switch len(args) {
	case 0:
		args = []string{redis.DefaultHash}
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	psc, err := redis.Subscribe(args[0])
	if err != nil {
		return err
	}
	defer psc.Close()

	m := make(map[string]*entry)

	for {
		v := psc.Receive()
		switch t := v.(type) {
		case redigo.Message:
			if t.Channel != args[0] {
				continue
			}
			x := bytes.Split(t.Data, []byte(sep))
			if len(x) != 2 {
				continue
			}
			k := string(x[0])
			s := string(x[1])
			if k == "delete" {
				delete(m, s)
				continue
			}
			now := time.Now()
			old, found := m[k]
			if !found {
				m[k] = &entry{s, now}
				continue
			}
			if old.s == s {
				old.t = now
				continue
			}
			if old.t.After(now) || old.t.Equal(now) {
				continue
			}
			fmt.Print(redis.Quotes(k), sep)
			oldf, oldferr := strconv.ParseFloat(old.s, 64)
			newf, newferr := strconv.ParseFloat(s, 64)
			if oldferr == nil && newferr == nil {
				delta := newf - oldf
				sec := now.Sub(old.t).Seconds()
				fmt.Printf("%g (%.0f/s)\n", delta, delta/sec)
			} else {
				fmt.Println(redis.Quotes(s))
			}
			old.s = s
			old.t = now
		case error:
			if !*c {
				err = t
			}
			break
		}
	}
	return err
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
	hdeta [CHANNEL]

DESCRIPTION

	Print the redis hash fields that change between invocations. If the
	field value is an int or float, this prints the difference between
	value followed by the delta divided by the seconds since last
	invocation. The CHANNEL parameter is the respective redis hash.`,
	}
}
