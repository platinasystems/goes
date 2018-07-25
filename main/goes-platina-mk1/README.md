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
Test/xeth/bad-names
Test/xeth/good-names
Test/vnet/ready
Test/vnet/nodocker/twohost
Test/vnet/nodocker/onerouter
Test/vnet/docker/bird/bgp/eth/connectivity
Test/vnet/docker/bird/bgp/eth/bird
Test/vnet/docker/bird/bgp/eth/neighbors
Test/vnet/docker/bird/bgp/eth/routes
Test/vnet/docker/bird/bgp/eth/inter-connectivity
Test/vnet/docker/bird/bgp/eth/flap
Test/vnet/docker/bird/bgp/vlan/connectivity
Test/vnet/docker/bird/bgp/vlan/bird
Test/vnet/docker/bird/bgp/vlan/neighbors
Test/vnet/docker/bird/bgp/vlan/routes
Test/vnet/docker/bird/bgp/vlan/inter-connectivity
Test/vnet/docker/bird/bgp/vlan/flap
Test/vnet/docker/bird/ospf/eth/connectivity
Test/vnet/docker/bird/ospf/eth/bird
Test/vnet/docker/bird/ospf/eth/neighbors
Test/vnet/docker/bird/ospf/eth/routes
Test/vnet/docker/bird/ospf/eth/inter-connectivity
Test/vnet/docker/bird/ospf/eth/flap
Test/vnet/docker/bird/ospf/vlan/connectivity
Test/vnet/docker/bird/ospf/vlan/bird
Test/vnet/docker/bird/ospf/vlan/neighbors
Test/vnet/docker/bird/ospf/vlan/routes
Test/vnet/docker/bird/ospf/vlan/inter-connectivity
Test/vnet/docker/bird/ospf/vlan/flap
Test/vnet/docker/frr/bgp/eth/connectivity
Test/vnet/docker/frr/bgp/eth/frr
Test/vnet/docker/frr/bgp/eth/neighbors
Test/vnet/docker/frr/bgp/eth/routes
Test/vnet/docker/frr/bgp/eth/inter-connectivity
Test/vnet/docker/frr/bgp/eth/flap
Test/vnet/docker/frr/bgp/vlan/connectivity
Test/vnet/docker/frr/bgp/vlan/frr
Test/vnet/docker/frr/bgp/vlan/neighbors
Test/vnet/docker/frr/bgp/vlan/routes
Test/vnet/docker/frr/bgp/vlan/inter-connectivity
Test/vnet/docker/frr/bgp/vlan/flap
Test/vnet/docker/frr/isis/eth/connectivity
Test/vnet/docker/frr/isis/eth/frr
Test/vnet/docker/frr/isis/eth/config
Test/vnet/docker/frr/isis/eth/neighbors
Test/vnet/docker/frr/isis/eth/routes
Test/vnet/docker/frr/isis/eth/inter-connectivity
Test/vnet/docker/frr/isis/eth/flap
Test/vnet/docker/frr/isis/vlan/connectivity
Test/vnet/docker/frr/isis/vlan/frr
Test/vnet/docker/frr/isis/vlan/config
Test/vnet/docker/frr/isis/vlan/neighbors
Test/vnet/docker/frr/isis/vlan/routes
Test/vnet/docker/frr/isis/vlan/inter-connectivity
Test/vnet/docker/frr/isis/vlan/flap
Test/vnet/docker/frr/ospf/eth/connectivity
Test/vnet/docker/frr/ospf/eth/frr
Test/vnet/docker/frr/ospf/eth/neighbors
Test/vnet/docker/frr/ospf/eth/routes
Test/vnet/docker/frr/ospf/eth/inter-connectivity
Test/vnet/docker/frr/ospf/eth/flap
Test/vnet/docker/frr/ospf/vlan/connectivity
Test/vnet/docker/frr/ospf/vlan/frr
Test/vnet/docker/frr/ospf/vlan/neighbors
Test/vnet/docker/frr/ospf/vlan/routes
Test/vnet/docker/frr/ospf/vlan/inter-connectivity
Test/vnet/docker/frr/ospf/vlan/flap
Test/vnet/docker/gobgp/ebgp/eth/connectivity
Test/vnet/docker/gobgp/ebgp/eth/gobgp
Test/vnet/docker/gobgp/ebgp/eth/neighbors
Test/vnet/docker/gobgp/ebgp/eth/routes
Test/vnet/docker/gobgp/ebgp/eth/inter-connectivity
Test/vnet/docker/gobgp/ebgp/eth/flap
Test/vnet/docker/gobgp/ebgp/vlan/connectivity
Test/vnet/docker/gobgp/ebgp/vlan/gobgp
Test/vnet/docker/gobgp/ebgp/vlan/neighbors
Test/vnet/docker/gobgp/ebgp/vlan/routes
Test/vnet/docker/gobgp/ebgp/vlan/inter-connectivity
Test/vnet/docker/gobgp/ebgp/vlan/flap
Test/vnet/docker/net/slice/vlan/connectivity
Test/vnet/docker/net/slice/vlan/frr
Test/vnet/docker/net/slice/vlan/routes
Test/vnet/docker/net/slice/vlan/inter-connectivity
Test/vnet/docker/net/slice/vlan/isolation
Test/vnet/docker/net/slice/vlan/stress
Test/vnet/docker/net/slice/vlan/stress-pci
Test/vnet/docker/net/dhcp/eth/connectivity
Test/vnet/docker/net/dhcp/eth/server
Test/vnet/docker/net/dhcp/eth/client
Test/vnet/docker/net/dhcp/eth/connectivity2
Test/vnet/docker/net/dhcp/eth/vlan-tag
Test/vnet/docker/net/dhcp/vlan/connectivity
Test/vnet/docker/net/dhcp/vlan/server
Test/vnet/docker/net/dhcp/vlan/client
Test/vnet/docker/net/dhcp/vlan/connectivity2
Test/vnet/docker/net/dhcp/vlan/vlan-tag
Test/vnet/docker/net/static/eth/connectivity
Test/vnet/docker/net/static/eth/frr
Test/vnet/docker/net/static/eth/routes
Test/vnet/docker/net/static/eth/inter-connectivity
Test/vnet/docker/net/static/eth/flap
Test/vnet/docker/net/static/eth/inter-connectivity2
Test/vnet/docker/net/static/eth/punt-stress
Test/vnet/docker/net/static/vlan/connectivity
Test/vnet/docker/net/static/vlan/frr
Test/vnet/docker/net/static/vlan/routes
Test/vnet/docker/net/static/vlan/inter-connectivity
Test/vnet/docker/net/static/vlan/flap
Test/vnet/docker/net/static/vlan/inter-connectivity2
Test/vnet/docker/net/static/vlan/punt-stress
PASS
```
---

*&copy; 2015-2017 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE

