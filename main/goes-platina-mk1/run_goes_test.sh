#!/bin/bash

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi

testcases=("Test/vnet.ready"
	   "Test/nodocker/twohost"
	   "Test/nodocker/onerouter"
	   "Test/docker/net/slice/vlan"
	   "Test/docker/net/dhcp/eth"
	   "Test/docker/net/dhcp/vlan"
	   "Test/docker/net/static/eth"
	   "Test/docker/net/static/vlan"
	   "Test/docker/frr/ospf/eth"
	   "Test/docker/frr/ospf/vlan"
	   "Test/docker/frr/isis/eth"
	   "Test/docker/frr/isis/vlan"
	   "Test/docker/frr/bgp/eth"
	   "Test/docker/frr/bgp/vlan"
	   "Test/docker/bird/bgp/eth"
	   "Test/docker/bird/bgp/vlan"
	   "Test/docker/bird/ospf/eth"
	   "Test/docker/bird/ospf/vlan"
	   "Test/docker/gobgp/ebgp/eth"
	   "Test/docker/gobgp/ebgp/vlan")

tester=./goes-platina-mk1.test

quit=0

sigint() {
    echo "quit after current testcase"
    quit=1
}

trap 'sigint'  INT

fix_it() {
  goes stop
  docker stop R1 R2 R3 R4 > /dev/null 2>&1
  docker rm -v R1 R2 R3 R4 > /dev/null 2>&1
  docker stop CA-1 RA-1 RA-2 CA-2 CB-1 RB-1 RB-2 CB-2 > /dev/null 2>&1
  docker rm -v CA-1 RA-1 RA-2 CA-2 CB-1 RB-1 RB-2 CB-2 > /dev/null 2>&1
  ip -all netns del
}

for t in ${testcases[@]}; do
    fix_it
    if [ "$quit" -eq 1 ]; then
	exit 0
    fi
    printf "Running %27s: " $t
    ${tester} -test.v -test.run=$t > /tmp/out 2>&1
    if [ $? == 0 ]; then
	echo "OK"
    else
	log=${t//\//_}.out
	mv /tmp/out ./$log
	if grep -q panic ./$log; then
	    echo "Crashed"	
	else
	    echo "Failed"	
	fi
    fi
done

exit 0
