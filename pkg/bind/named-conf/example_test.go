// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataExample(t *testing.T) {
	testdata(t, fnExample, expectExample)
}

const fnExample = "testdata/bind9/contrib/dlz/example/named.conf"

var expectExample = Conf{
	Statement{"options", Block{
		Statement{"allow-transfer", Block{Statement{"any"}}},
		Statement{"allow-query", Block{Statement{"any"}}},
		Statement{"notify", "yes"},
		Statement{"recursion", "no"},
	}},
	Statement{"dlz", "example", Block{Statement{"database",
		"dlopen ./dlz_example.so example.nil",
	}}},
}
