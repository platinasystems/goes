// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package subscribe

import (
	"fmt"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "subscribe"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }

func (cmd) Usage() string { return Name + " CHANNEL" }

func (cmd) Main(args ...string) error {
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print messages published to the given redis channel",
	}
}
