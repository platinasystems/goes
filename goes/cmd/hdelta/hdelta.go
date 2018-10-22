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
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct {
	closed bool
}

type entry struct {
	s string
	t time.Time
}

func (*Command) String() string { return "hdelta" }

func (*Command) Usage() string { return "hdelta [CHANNEL]" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print the changed fields of a redis hash",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print the redis hash fields that change between invocations. If the
	field value is an int or float, this prints the difference between
	value followed by the delta divided by the seconds since last
	invocation. The CHANNEL parameter is the respective redis hash.`,
	}
}

func (c *Command) Main(args ...string) error {
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
			const sep = ": "
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
			if !c.closed {
				err = t
			}
			break
		}
	}
	return err
}

func (c *Command) Close() error {
	c.closed = true
	return nil
}
