// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataPerl(t *testing.T) {
	testdata(t, fnPerl, expectPerl)
}

const fnPerl = "testdata/bind9/contrib/dlz/modules/perl/testing/named.conf"

var expectPerl = Conf{
	Statement{"options", Block{
		Statement{"port", "5300"},
		Statement{"pid-file", "named.pid"},
		Statement{"session-keyfile", "session.key"},
		Statement{"listen-on", Block{Statement{"127.0.0.1"}}},
		Statement{"listen-on-v6", Block{Statement{"none"}}},
		Statement{"recursion", "no"},
		Statement{"notify", "no"},
	}},
	Statement{"dlz", "perl zone", Block{Statement{"database", `
dlopen ../dlz_perl_driver.so dlz_perl_example.pm dlz_perl_example`[1:],
	}}},
}
