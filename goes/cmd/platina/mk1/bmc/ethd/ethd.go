// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ethd

import (
	"fmt"
	"net/rpc"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name    = "ethd"
	Apropos = "ethd redis ethernet control"
	Usage   = "ethd"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

var (
	Init = func() {}
	once sync.Once

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)

	first int
)

type Command struct {
	Info
}

type Info struct {
	mutex sync.Mutex
	rpc   *sockfile.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Name }

func (c *Command) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t

	err := redis.IsReady()
	if err != nil {
		return err
	}

	first = 1
	c.stop = make(chan struct{})
	c.last = make(map[string]float64)
	c.lasts = make(map[string]string)
	c.lastu = make(map[string]uint16)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if c.rpc, err = sockfile.NewRpcServer(Name); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", Name, "Info")
		if err != nil {
			return err
		}
	}

	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err = c.update(); err != nil {
			}
		}
	}
	return nil
}

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

var ipaddr = "10.11.12.13"

func (c *Command) update() error {
	if err := writeRegs(); err != nil {
		return err
	}

	if first == 1 {
		//TODO use NL to read current IPADDR
		first = 0
	}

	for k, _ := range VpageByKey {
		if strings.Contains(k, "ipaddr") {
			v := ipaddr
			if (v != "") && (v != c.lasts[k]) {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
	}
	return nil
}

func writeRegs() error {
	for k, v := range WrRegVal {
		switch WrRegFn[k] {
		case "speed":
			if false {
				log.Print("test", k, v)
			}
		}
		delete(WrRegVal, k)
	}
	return nil
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	fmt.Println("ARGS FIELD = ", args.Field)
	fmt.Println("ARGS VALUE = ", args.Value)
	_, p := WrRegFn[args.Field]
	if !p {
		return fmt.Errorf("cannot hset: %s", args.Field)
	}
	_, q := WrRegRng[args.Field]
	if !q {
		err := i.set(args.Field, string(args.Value), false)
		if err == nil {
			*reply = 1
			WrRegVal[args.Field] = string(args.Value)
		}
		return err
	}
	var a [2]int
	var b [2]string
	var e [2]error
	if len(WrRegRng[args.Field]) == 1 {
		for i, v := range WrRegRng[args.Field] {
			b[i] = string(v)
		}
		if b[0] == "0.0.0.0" {
			n := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
			r := n + "\\." + n + "\\." + n + "\\." + n
			x := regexp.MustCompile(r)
			if x.MatchString(string(args.Value)) {
				err := i.set(args.Field,
					string(args.Value), false)
				if err == nil {
					*reply = 1
					WrRegVal[args.Field] =
						string(args.Value)
					//TODO use NL to write eth0
				}
				return err
			}
			return fmt.Errorf("Cannot hset.  Valid range is: %s",
				WrRegRng[args.Field])
		}
	}
	if len(WrRegRng[args.Field]) == 2 {
		for i, v := range WrRegRng[args.Field] {
			a[i], e[i] = strconv.Atoi(v)
		}
		if e[0] == nil && e[1] == nil {
			val, err := strconv.Atoi(string(args.Value))
			if err != nil {
				return err
			}
			if val >= a[0] && val <= a[1] {
				err := i.set(args.Field,
					string(args.Value), false)
				if err == nil {
					*reply = 1
					WrRegVal[args.Field] =
						string(args.Value)
				}
				return err
			}
			return fmt.Errorf("Cannot hset.  Valid range is: %s",
				WrRegRng[args.Field])
		}
	}
	for _, v := range WrRegRng[args.Field] {
		if v == string(args.Value) {
			err := i.set(args.Field, string(args.Value), false)
			if err == nil {
				*reply = 1
				WrRegVal[args.Field] = string(args.Value)
			}
			return err
		}
	}
	return fmt.Errorf("Cannot hset.  Valid values are: %s",
		WrRegRng[args.Field])
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}
