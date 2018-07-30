This is a GO Embedded System for Platina Systems' *mark 1* packet switches.

To run unit tests, loopback 6 pairs for ports and edit the configuration
as follows:
```console
$ editor main/goes-platina-mk1/testdata/netport.yaml
$ git update-index --assume-unchanged \
	main/goes-platina-mk1/testdata/netport.yaml
$ editor main/goes-platina-mk1/testdata/ethtool.yaml
$ git update-index --assume-unchanged \
	main/goes-platina-mk1/testdata/ethtool.yaml
```

Then build the unit test and run.
```console
$ go install github.com:platinasystems/go/main/goes-build
$ goes-build goes-platina-mk1.test
$ sudo ./goes-platina-mk1.test
```

Options:
```console
-test.alpha	this is a zero based alpha system
-test.cd	change to named directory before running tests
-test.dryrun	don't run, just print test names
-test.main	internal flag to run given goes command
-test.pause	enable progromatic pause to start debugger
-test.run=Test/PATTERN
		run the matching tests
-test.v		verbose
-test.vv	log test.Program output
-test.vvv	log test.Program execution
```

For example:
```console
sudo ./goes-platina-mk1.test -test.v -test.run Test/docker/frr/ospf/eth
sudo ./goes-platina-mk1.test -test.v -test.run Test/docker/frr/.*/eth
sudo ./goes-platina-mk1.test -test.v -test.run Test/.*/.*/.*/vlan
```

To list all tests:
```console
$ ./goes-platina-mk1.test -test.dryrun
Test
Test/xeth
Test/xeth/bad-names
Test/xeth/good-names
Test/vnet
Test/vnet/ready
Test/vnet/nodocker
Test/vnet/nodocker/twohost
Test/vnet/nodocker/onerouter
Test/vnet/docker
Test/vnet/docker/bird
Test/vnet/docker/bird/bgp
Test/vnet/docker/bird/bgp/eth
Test/vnet/docker/bird/bgp/vlan
Test/vnet/docker/bird/ospf
Test/vnet/docker/bird/ospf/eth
Test/vnet/docker/bird/ospf/vlan
Test/vnet/docker/frr
Test/vnet/docker/frr/bgp
Test/vnet/docker/frr/bgp/eth
Test/vnet/docker/frr/bgp/vlan
Test/vnet/docker/frr/isis
Test/vnet/docker/frr/isis/eth
Test/vnet/docker/frr/isis/vlan
Test/vnet/docker/frr/ospf
Test/vnet/docker/frr/ospf/eth
Test/vnet/docker/frr/ospf/vlan
Test/vnet/docker/gobgp
Test/vnet/docker/gobgp/ebgp
Test/vnet/docker/gobgp/ebgp/eth
Test/vnet/docker/gobgp/ebgp/vlan
Test/vnet/docker/net
Test/vnet/docker/net/slice
Test/vnet/docker/net/slice/vlan
Test/vnet/docker/net/dhcp
Test/vnet/docker/net/dhcp/eth
Test/vnet/docker/net/dhcp/vlan
Test/vnet/docker/net/static
PASS
```
---

*&copy; 2015-2017 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE

