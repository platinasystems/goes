// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package subscribe

import (
	"fmt"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/redis"
)

type Command struct{}

func (Command) String() string { return "subscribe" }

func (Command) Usage() string { return "subscribe CHANNEL" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print messages published to given redis channel",
	}
}

func (Command) Main(args ...string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("CHANNEL: missing")
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	psc, err := redis.Subscribe(args[0])
	if err != nil {
		return err
	}
	defer psc.Close()
	for {
		v := psc.Receive()
		switch t := v.(type) {
		case redigo.Message:
			if t.Channel == redis.DefaultHash {
				fmt.Println(string(t.Data))
			} else {
				fmt.Printf("%s <- %q\n", t.Channel, t.Data)
			}
		case error:
			err = t
			break
		}
	}
	return err
}
