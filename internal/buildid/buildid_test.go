// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package buildid

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBuildId(t *testing.T) {
	self, err := os.Readlink("/proc/self/exe")
	if err != nil {
		t.Fatal(err)
	}
	got, err := New(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(got)
	b, err := exec.Command("go", "tool", "buildid", self).Output()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(string(b))
	if got != string(want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
