// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package args provides types for the redis RPC arguments.
package args

type Assign struct {
	Key    string
	AtSock string
	Name   string
}

type Unassign struct {
	Key string
}

type Del struct {
	Key  string
	Keys []string
}

type Get struct{ Key string }

type Set struct {
	Key   string
	Value []byte
}

type Hdel struct {
	Key    string
	Field  string
	Fields []string
}

type Hexists struct{ Key, Field string }
type Hget struct{ Key, Field string }
type Hgetall struct{ Key string }
type Hkeys struct{ Key string }

type Hset struct {
	Key, Field string
	Value      []byte
}

type Lrange struct {
	Key   string
	Start int
	Stop  int
}

type Lindex struct {
	Key   string
	Index int
}

type Blpop struct {
	Key  string
	Keys []string
}

type Brpop struct {
	Key  string
	Keys []string
}

type Lpush struct {
	Key    string
	Value  []byte
	Values [][]byte
}

type Rpush struct {
	Key    string
	Value  []byte
	Values [][]byte
}
