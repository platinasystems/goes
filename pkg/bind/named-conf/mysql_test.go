// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataMySQL(t *testing.T) {
	testdata(t, fnMySQL, expectMySQL)
}

const fnMySQL = "testdata/bind9/contrib/dlz/modules/mysql/testing/named.conf"

var expectMySQL = Conf{
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
		`dlopen ../dlz_mysql_dynamic.so
           {
             host=127.0.0.1 port=3306 socket=/tmp/mysql.sock
             dbname=BindDB user=USER pass=PASSWORD threads=2
           }
           {SELECT zone FROM records WHERE zone = '$zone$'}
           {SELECT ttl, type, mx_priority, IF(type = 'TXT', CONCAT('"',data,'"'), data) AS data FROM records WHERE zone = '$zone$' AND host = '$record$' AND type <> 'SOA' AND type <> 'NS'}
           {SELECT ttl, type, data, primary_ns, resp_contact, serial, refresh, retry, expire, minimum FROM records WHERE zone = '$zone$' AND (type = 'SOA' OR type='NS')}
           {SELECT ttl, type, host, mx_priority, IF(type = 'TXT', CONCAT('"',data,'"'), data) AS data, resp_contact, serial, refresh, retry, expire, minimum FROM records WHERE zone = '$zone$' AND type <> 'SOA' AND type <> 'NS'}
           {SELECT zone FROM xfr where zone='$zone$' AND client = '$client$'}`,
	}}},
}
