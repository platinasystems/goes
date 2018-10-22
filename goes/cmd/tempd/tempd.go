// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tempd

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
)

const (
	hwmon = "/sys/class/hwmon/"
)

var bmcIpv6LinkLocalRedis string

type Command struct {
	VpageByKey map[string]uint8

	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint8
}

func (*Command) String() string { return "tempd" }

func (*Command) Usage() string { return "tempd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "temperature monitoring daemon, publishes to redis",
	}
}

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var err error
	var si syscall.Sysinfo_t

	if err = redis.IsReady(); err != nil {
		log.Print("redis not ready")
		return err
	}

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

func (c *Command) update() error {
	for k, _ := range c.VpageByKey {
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
			if i != "" && i != c.lasts[k] {
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
			var bmcIpv4 interface{}
			bmcIpv4, err = d.Do("HGET", redis.DefaultHash+"-bmc", "eth0.ipv4")
			if err != nil {
				bmcIpv4, _ = d.Do("HGET", "platina", "eth0.ipv4") // to support old bmc builds
			}
			s := fmt.Sprint(bmcIpv4)
			if !strings.Contains(s, "ERROR") {
				r, _ := regexp.Compile("([0-9]+).([0-9]+).([0-9]+).([0-9]+)")
				i = r.FindString(string(bmcIpv4.([]uint8)))
			}
			d.Close()
		}
	}
	return i, v

}

func cpuCoreTemp() string {
	hi := float64(0)
	t, err := readTemp("hwmon0", "core") // assumes lm-sensors
	if err == nil && t > hi {
		hi = t
	}
	t, err = readTemp("hwmon1", "lm75") // assumes device is discovered
	if err == nil && t > hi {
		hi = t
	}
	v := fmt.Sprintf("%.2f\n", hi/1000)
	if v == "0" {
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
			_, err = d.Do("HSET", redis.DefaultHash+"-bmc", "host.temp.units.C", v)
			if err != nil {
				d.Do("HSET", "platina", "host.temp.units.C", v) // to support old bmc builds
			}
			d.Close()
		}
	}
	return v
}

func readTemp(dir string, dev string) (h float64, err error) {
	h = float64(0)
	n, err := ioutil.ReadFile(hwmon + dir + "/name")
	if err != nil {
		return h, err
	}
	if strings.Contains(string(n), dev) {
		l, err := ioutil.ReadDir(hwmon + dir)
		if err != nil {
			return h, err
		}
		for _, f := range l {
			if strings.Contains(f.Name(), "_input") {
				t, err := ioutil.ReadFile(hwmon + dir + "/" + f.Name())
				if err == nil {
					tt := strings.Split(string(t), "\n")
					ttf, err := strconv.ParseFloat(tt[0], 64)
					if err == nil {
						if ttf > h {
							h = ttf
						}
					}
				}
			}
		}
	}
	return h, nil
}
