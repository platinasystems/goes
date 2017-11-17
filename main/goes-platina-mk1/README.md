This is a GO Embedded System for Platina Systems' *mark 1* packet switches.

To run unit tests, loopback 6 pairs for ports and edit the configuration
as follows:

```console
$ editor main/goes-platina-mk1/test/port2port/conf.go
$ git update-index --assume-unchanged \
	main/goes-platina-mk1/test/port2port/conf.go
```

Then build the unit test and run.

```console
$ make -B goes-platina-mk1.test
$ sudo ./goes-platina-mk1.test -test.v		# -test.run=./SUB/TEST
```

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE

