// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataMySQLDyn(t *testing.T) {
	testdata(t, fnMySQLDyn, expectMySQLDyn)
}

const fnMySQLDyn = "testdata/bind9/contrib/dlz/modules/mysqldyn/testing/named.conf"

var expectMySQLDyn = Conf{
	Statement{"controls", Block{}},
	Statement{"options", Block{
		Statement{"directory", "."},
		Statement{"port", "5300"},
		Statement{"pid-file", "named.pid"},
		Statement{"session-keyfile", "session.key"},
		Statement{"listen-on", Block{Statement{"any"}}},
		Statement{"listen-on-v6", Block{Statement{"none"}}},
		Statement{"recursion", "no"},
	}},
	Statement{"key", "rndc_key", Block{
		Statement{"secret", "1234abcd8765"},
		Statement{"algorithm", "hmac-md5"},
	}},
	Statement{"controls", Block{Statement{
		"inet", "127.0.0.1",
		"port", "9953",
		"allow", Block{Statement{"any"}},
		"keys", Block{Statement{"rndc_key"}},
	}}},
	Statement{"dlz", "test", Block{Statement{"database",
		`dlopen ../dlz_mysqldyn_mod.so BindDB localhost root password`,
	}}},
}
