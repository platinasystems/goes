// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import "testing"

func TestDataLDAP(t *testing.T) {
	testdata(t, fnLDAP, expectLDAP)
}

const fnLDAP = "testdata/bind9/contrib/dlz/modules/ldap/testing/named.conf"

var expectLDAP = Conf{
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
	Statement{"controls", Block{
		Statement{
			"inet", "127.0.0.1",
			"port", "9953",
			"allow", Block{Statement{"any"}},
			"keys", Block{Statement{"rndc_key"}},
		},
	}},
	Statement{"dlz", "test", Block{Statement{"database",
		`dlopen ../dlz_ldap_dynamic.so 2
        v3 simple {cn=Manager,o=bind-dlz} {secret} {127.0.0.1}
        ldap:///dlzZoneName=$zone$,ou=dns,o=bind-dlz???objectclass=dlzZone
        ldap:///dlzHostName=$record$,dlzZoneName=$zone$,ou=dns,o=bind-dlz?dlzTTL,dlzType,dlzPreference,dlzData,dlzIPAddr?sub?(&(objectclass=dlzAbstractRecord)(!(dlzType=soa)))
        ldap:///dlzHostName=@,dlzZoneName=$zone$,ou=dns,o=bind-dlz?dlzTTL,dlzType,dlzData,dlzPrimaryNS,dlzAdminEmail,dlzSerial,dlzRefresh,dlzRetry,dlzExpire,dlzMinimum?sub?(&(objectclass=dlzAbstractRecord)(dlzType=soa))
        ldap:///dlzZoneName=$zone$,ou=dns,o=bind-dlz?dlzTTL,dlzType,dlzHostName,dlzPreference,dlzData,dlzIPAddr,dlzPrimaryNS,dlzAdminEmail,dlzSerial,dlzRefresh,dlzRetry,dlzExpire,dlzMinimum?sub?(&(objectclass=dlzAbstractRecord)(!(dlzType=soa)))
        ldap:///dlzZoneName=$zone$,ou=dns,o=bind-dlz??sub?(&(objectclass=dlzXFR)(dlzIPAddr=$client$))`,
	}}},
}
