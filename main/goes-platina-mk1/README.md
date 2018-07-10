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
$ go install github.com:platinasystems/go/main/goes-build
$ goes-build goes-platina-mk1.test
$ sudo ./goes-platina-mk1.test
```

Options:
```console
-test.v		verbose
-test.vv	log test.Program output
-test.cd	change to named directory before running tests
-test.run=Test/TEST
		run named test instead of all
-test.Pause	pause before and after suite
```

Current tests:
```
Test/vnet.ready

Test/nodocker/twohost
Test/nodocker/onerouter

Test/docker/net/slice/vlan

Test/docker/net/dhcp/eth
Test/docker/net/dhcp/vlan

Test/docker/net/static/eth
Test/docker/net/static/vlan

Test/docker/frr/ospf/eth
Test/docker/frr/ospf/vlan
Test/docker/frr/isis/eth
Test/docker/frr/isis/vlan
Test/docker/frr/bgp/eth
Test/docker/frr/bgp/vlan

Test/docker/bird/bgp/eth
Test/docker/bird/bgp/vlan
Test/docker/bird/ospf/eth
Test/docker/bird/ospf/vlan

Test/docker/gobgp/ebgp/eth
Test/docker/gobgp/ebgp/vlan
```

For example:
```console
sudo ./goes-platina-mk1.test -test.vv -test.run Test/docker/frr/ospf/eth
sudo ./goes-platina-mk1.test -test.vv -test.run ./.*/.*/.*/vlan
```

---

*&copy; 2015-2017 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE

