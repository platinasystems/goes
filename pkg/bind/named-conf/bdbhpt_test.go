// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataBDBHPT(t *testing.T) {
	testdata(t, fnBDBHPT, expectBDBHPT)
}

const fnBDBHPT = "testdata/bind9/contrib/dlz/modules/bdbhpt/testing/named.conf"

var expectBDBHPT = Conf{
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
	Statement{"dlz", "bdbhpt_dynamic", Block{Statement{"database",
		"dlopen ../dlz_bdbhpt_dynamic.so T . test.db",
	}}},
}
