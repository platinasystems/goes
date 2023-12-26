// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package flagset

import (
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func Search[T comparable](flags *flag.FlagSet, name string) T {
	var zero T
	f := flags.Lookup(name)
	if f == nil {
		return zero
	}
	elem := reflect.ValueOf(f.Value).Elem()
	vtype := reflect.TypeOf(zero)
	if elem.CanConvert(vtype) {
		return elem.Convert(vtype).Interface().(T)
	}
	return zero
}

func SilentParse(flags *flag.FlagSet, args []string) error {
	if flags.Parsed() {
		return nil
	}
	flags.Init("", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}
	return flags.Parse(args)
}

func Sprint(flags *flag.FlagSet) string {
	sb := new(strings.Builder)
	fmt.Fprintln(sb)
	flags.SetOutput(sb)
	flags.PrintDefaults()
	flags.SetOutput(io.Discard)
	return sb.String()
}
