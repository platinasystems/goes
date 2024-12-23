// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"io"
	"strings"
	"testing"
)

const scanTestInput = `
shell comment; # ...
slash comment; // ...
c-lang end-of-line comment; /* ... */
c-lang in-line /* ... */ comment;
/*
* ...
*/ c-lang block-comment prefix;
c-lang block-comment suffix; /*
* ...
*/

/*
 * first comment
 */

/*
 * second comment
 */

two c-lang block comments;

/*
 /* nested comment */
 */
nested comment;

string "hello world";
list { first; second; third; };
compact-list {first;second;third;};
bool-true true;
bool-false false;
bool-yes yes;
bool-no no;
uint 0xffe;
int 123;
-int -123;
+int +123;
-float -1.23;
+float +1.23;
efloat 1.23e10;
localhost-ip4 127.0.0.1;
localhost-ip6 ::1;

controls {
        inet 127.0.0.1 port 9953 allow { any; } keys { rndc_key; };
};
`

var scanTestExpect = []string{
	"shell", "comment", ";",
	"slash", "comment", ";",
	"c-lang", "end-of-line", "comment", ";", "/*", "...", "*/",
	"c-lang", "in-line", "/*", "...", "*/", "comment", ";",
	"/*", "*", "...", "*/", "c-lang", "block-comment", "prefix", ";",
	"c-lang", "block-comment", "suffix", ";",
	"/*", "*", "...", "*/",
	"/*", "*", "first", "comment", "*/",
	"/*", "*", "second", "comment", "*/",
	"two", "c-lang", "block", "comments", ";",
	"/*", "/*", "nested", "comment", "*/", "*/",
	"nested", "comment", ";",
	"string", `hello world`, ";",
	"list", "{", "first", ";", "second", ";", "third", ";", "}", ";",
	"compact-list", "{", "first", ";", "second", ";", "third", ";", "}", ";",
	"bool-true", "true", ";",
	"bool-false", "false", ";",
	"bool-yes", "yes", ";",
	"bool-no", "no", ";",
	"uint", "0xffe", ";",
	"int", "123", ";",
	"-int", "-123", ";",
	"+int", "+123", ";",
	"-float", "-1.23", ";",
	"+float", "+1.23", ";",
	"efloat", "1.23e10", ";",
	"localhost-ip4", "127.0.0.1", ";",
	"localhost-ip6", "::1", ";",
	"controls", "{",
	"inet", "127.0.0.1", "port", "9953",
	"allow", "{", "any", ";", "}",
	"keys", "{", "rndc_key", ";", "}", ";",
	"}", ";",
}

func TestScanner(t *testing.T) {
	tlog = t.Log
	tlogf = t.Logf
	var i int
	tlog = t.Log
	tlogf = t.Logf
	r := strings.NewReader(scanTestInput)
	sc := NewScanner(r)
	for i = 0; sc.Scan(); i++ {
		s := sc.Text()
		if i > len(scanTestExpect) {
			t.Errorf("unexpected: %q", s)
			return
		}
		if s != scanTestExpect[i] {
			t.Errorf("%d: %q != %q", i, s, scanTestExpect[i])
			return
		}
	}
	if i < len(scanTestExpect) {
		t.Error("missing", scanTestExpect[i:])
	}
}

const parseTestInput = `
hello \"world\";
key - word hyphenated;
hyphenated val - ue;
ip6 0 : 1 : 2 : 3 : 4 : 5 : 6 : 7 : 8 : 9 : a : b : c : d : e : f;
controls {
  inet 127.0.0.1 port 9953 allow { any; } keys { rndc_key; };
};
`

var parseTestExpect = Conf{
	Statement{"hello", `"world"`},
	Statement{"key-word", "hyphenated"},
	Statement{"hyphenated", "val-ue"},
	Statement{"ip6", "0:1:2:3:4:5:6:7:8:9:a:b:c:d:e:f"},
	Statement{"controls", Block{Statement{
		"inet", "127.0.0.1",
		"port", "9953",
		"allow", Block{Statement{"any"}},
		"keys", Block{Statement{"rndc_key"}},
	}}},
}

func TestParser(t *testing.T) {
	tlog = t.Log
	tlogf = t.Logf
	r := strings.NewReader(parseTestInput)
	conf, err := NewScanner(r).conf()
	if err != nil {
		t.Error(err)
	} else if err := conf.verify(parseTestExpect); err != nil {
		t.Error(err)
	} else {
		t.Logf("OK:\n%v", conf)
	}
}

func testdata(t *testing.T, fn string, expect Conf) {
	t.Helper()
	tlog = t.Log
	tlogf = t.Logf
	Open = func(filename string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("")), nil
	}
	if conf, err := NewConf(fn); err != nil {
		t.Error(err)
	} else if err := conf.verify(expect); err != nil {
		t.Error(err)
	} else {
		t.Logf("OK:\n%v", conf)
	}
}
