// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tempd

import (
	"encoding/hex"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const (
	Name    = "tempd"
	Apropos = "temperature monitoring daemon, publishes to redis"
	Usage   = "tempd"
)

var bmcIpv6LinkLocalRedis string

var Init = func() {}

var once sync.Once

func New() *Command { return new(Command) }

type Command struct {
	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint8
}

var VpageByKey map[string]uint8

func (*Command) Apropos() lang.Alt { return apropos }

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var err error

	if err = redis.IsReady(); err != nil {
		log.Print("redis not ready")
		return err
	}

	once.Do(Init)

	var si syscall.Sysinfo_t

	c.stop = make(chan struct{})
	c.last = make(map[string]float64)
	c.lasts = make(map[string]string)
	c.lastu = make(map[string]uint8)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}
	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err = c.update(); err != nil {
				close(c.stop)
				return err
			}
		}
	}

	return nil
}

func (*Command) String() string { return Name }
func (*Command) Usage() string  { return Usage }

func (c *Command) update() error {
	for k, _ := range VpageByKey {
		if strings.Contains(k, "coretemp") {
			v := cpuCoreTemp()
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		} else if strings.Contains(k, "bmc.redis.status") {
			i, v := bmcStatus()
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
			/* FIXME placeholder for bmc eth0 address field */
			k = "bmc.eth0.ipv4"
			if i != c.lasts[k] {
				c.pub.Print(k, ": ", i)
				c.lasts[k] = i
			}
		}
	}
	return nil
}

func bmcStatus() (string, string) {
	var v, i string

	if bmcIpv6LinkLocalRedis == "" {
		m, err := redis.Hget(redis.DefaultHash, "eeprom.BaseEthernetAddress")
		if err == nil {
			o := strings.Split(m, ":")
			b, _ := hex.DecodeString(o[0])
			b[0] = b[0] ^ byte(2)
			o[0] = hex.EncodeToString(b)
			bmcIpv6LinkLocalRedis = "[fe80::" + o[0] + o[1] + ":" + o[2] + "ff:fe" + o[3] + ":" + o[4] + o[5] + "%eth0]:6379"
		}
	}
	if bmcIpv6LinkLocalRedis != "" {
		d, err := redigo.Dial("tcp", bmcIpv6LinkLocalRedis)
		if err != nil {
			v = "down"
		} else {
			v = "up"
			bmcIpv4, _ := d.Do("HGET", redis.DefaultHash, "eth0.ipv4")
			r, _ := regexp.Compile("([0-9]+).([0-9]+).([0-9]+).([0-9]+)")
			i = r.FindString(string(bmcIpv4.([]uint8)))
			d.Close()
		}
	}
	return i, v

}

func cpuCoreTemp() string {
	var max float64
	var v string
	t, err := exec.Command("sensors").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(t), "\n")

	max = -1000
	r, _ := regexp.Compile("[\\+,-]([0-9]+).[0-9]")
	for _, line := range lines {
		temps := r.FindAllString(line, -1)
		if temps != nil {
			f, err := strconv.ParseFloat(temps[0], 64)
			if err == nil {
				if f > max {
					max = f
				}
			}
		}
		v = strconv.FormatFloat(max, 'f', 1, 64)
	}
	if v == "-1000" {
		return ""
	}
	if bmcIpv6LinkLocalRedis == "" {
		m, err := redis.Hget(redis.DefaultHash, "eeprom.BaseEthernetAddress")
		if err == nil {
			o := strings.Split(m, ":")
			b, _ := hex.DecodeString(o[0])
			b[0] = b[0] ^ byte(2)
			o[0] = hex.EncodeToString(b)
			bmcIpv6LinkLocalRedis = "[fe80::" + o[0] + o[1] + ":" + o[2] + "ff:fe" + o[3] + ":" + o[4] + o[5] + "%eth0]:6379"
		}
	}
	if bmcIpv6LinkLocalRedis != "" {
		d, err := redigo.Dial("tcp", bmcIpv6LinkLocalRedis)
		if err == nil {
			d.Do("HSET", redis.DefaultHash, "host.temp.units.C", v)
			d.Close()
		}
	}
	return v
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
