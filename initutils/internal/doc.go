// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

/*
Package sbininit provides both /sbin/init and /usr/sbin/goesd go-es commands.
The former is registered if the PID is 1; otherwise, the later is used to
start/stop the packaged daemons. Thus a single binary may be used as the sole
binary in an embedded system or packaged binary within a full distro. Both
start a redis sever before running all daemons.

You may assign a new rpc handler for a given key with,

	redis.Assign(KEY, SOCKFILE, NAME)

Where NAME is prefaced to each rpc receiver as NAME.METHOD.

The redis server searches for the longest match KEY. If KEY is of the form
HASH:FIELD, the server looks for the longest match FIELD in the named HASH.

You may later disassociate this handler with,

	redis.Unassign(KEY)

Note that while all machines provide the following system interfaces, some may
provide more; so, please see the respective machine documentation.

cmdline - a hash interface to /proc/cmdline.

	redis hexists cmdline NAME
		returns 1 if the named paramter is in cmdline and 0 otherwise

	redis hgetall cmdline
		lists all cmdline parameters

	redis hget cmdline NAME
		returns the named cmdline value

daemons - a hash interface of the machine's running daemons

	redis hexists daemons PID
		returns 1 if given PID is active and 0 otherwise

	redis hgetall daemons
		lists active daemon process Ids with program name

	redis hget daemons PID
		returns the program name of the given PID

	redis hset daemons PID NAME
		should only used by internal goes methods
*/
package internal
