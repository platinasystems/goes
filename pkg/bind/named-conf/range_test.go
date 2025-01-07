// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
)

func TestRange(t *testing.T) {
	Open = func(filename string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("")), nil
	}
	conf, err := NewConf(fnGetToken)
	if err != nil {
		t.Fatal(err)
		return
	}
	kw := func(acc []string) string {
		return strings.Join(acc, ".") + ":"
	}
	sb := new(strings.Builder)
	fmt.Fprintln(sb)
	conf.Range(func(acc []string, v any) bool {
		fmt.Fprintln(sb, kw(acc), v)
		return true
	}, "options", "directory")
	conf.Range(func(acc []string, v any) bool {
		fmt.Fprintln(sb, kw(acc), v)
		return true
	}, "options", "forwarders")
	conf.Range(func(acc []string, v any) bool {
		fmt.Fprintln(sb, kw(acc), v)
		return true
	}, "options", regexp.MustCompile("m??-cache-ttl"))
	conf.Range(func(acc []string, v any) bool {
		fmt.Fprintln(sb, kw(acc), v)
		return true
	}, "view", "test-view", "in", "zone", "view-zone.com",
		"allow-update-forwarding")
	got := sb.String()
	const want = `
options.directory: /tmp;
options.forwarders: {
  1.2.3.4;
  5.6.7.8;
};
options.max-cache-ttl: 999;
options.min-cache-ttl: 66;
view.test-view.in.zone.view-zone.com.allow-update-forwarding: {
  10.0.0.34;
};
`
	if got != want {
		t.Errorf("MISMATCH\n%s\n---\n%s", got, want)
	} else {
		t.Log(got)
	}
}
