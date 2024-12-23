// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataWildcard(t *testing.T) {
	testdata(t, fnWildcard, expectWildcard)
}

const fnWildcard = "testdata/bind9/contrib/dlz/modules/wildcard/testing/named.conf"

var expectWildcard = Conf{
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
	Statement{"dlz", "test", Block{Statement{"database", `
dlopen ../dlz_wildcard_dynamic.so
        *example*.com 10.53.* 1800
        @      3600    SOA   {ns3.example.nil. support.example.nil. 42 14400 7200 2592000 600}
        @      3600    NS     ns3.example.nil.
        @      3600    NS     ns4.example.nil.
        @      3600    NS     ns8.example.nil.
        @      3600    MX     {5 mail.example.nil.}
        ftp    86400   A      192.0.0.1
        sql    86400   A      192.0.0.2
        tmp    {}      A      192.0.0.3
        www    86400   A      192.0.0.3
        www    86400   AAAA   ::1
        txt    300     TXT    {"you requested $record$ in $zone$"}
        *      86400   A      192.0.0.100`[1:],
	}}},
}
