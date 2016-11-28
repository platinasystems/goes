// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package info

import (
	"fmt"

	"github.com/platinasystems/go/sch"
)

var PubCh sch.In

type Interface interface {
	// Provider should return a list of longest match keys supported by
	// info provider. Some providers may modify the list with given args.
	Prefixes(...string) []string

	// Main should detect and report machine changes like this until Close.
	//
	//	Publish(KEY, VALUE)
	//
	// or, if device or attribute is removed,
	//
	//	Publish("delete", KEY)
	//
	Main(...string) error

	// Close should stop all info go-routines and release all resources.
	Close() error

	// Del should remove the attribute then publish with,
	//
	//	Publish("delete", KEY)
	Del(string) error

	// Set should assign the given machine attribute then publish the new
	// value with,
	//
	//	Publish(KEY, VALUE)
	//
	Set(string, string) error

	// String should return the provider name
	String() string
}

type Mainer interface {
	Main(...string) error
}

func CantDel(key string) error {
	return fmt.Errorf("%s: can't delete", key)
}

func CantSet(key string) error {
	return fmt.Errorf("%s: can't set", key)
}

func Publish(key, value interface{}) {
	PubCh <- fmt.Sprint(key, ": ", value)
}
