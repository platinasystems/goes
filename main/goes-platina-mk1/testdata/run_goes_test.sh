#!/bin/bash

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi

# document all known issues on same line as test
testcases=("Test/ready"
	   "Test/nodocker/ping-gateways"
	   "Test/nodocker/ping-remotes"
	   "Test/nodocker/hping01"
	   "Test/nodocker/hping10"
	   "Test/nodocker/hping30"
	   "Test/nodocker/hping60"	   	   	   
	   "Test/docker/bird/bgp/eth"
	   "Test/docker/bird/bgp/vlan"
	   "Test/docker/bird/ospf/eth"
	   "Test/docker/bird/ospf/vlan" # --- FAIL: Test/docker/bird/ospf/vlan/connectivity (RTA_MULTIPATH w/out gw)
	   "Test/docker/frr/bgp/eth"
	   "Test/docker/frr/bgp/vlan"
	   "Test/docker/frr/isis/eth"
	   "Test/docker/frr/isis/vlan"
	   "Test/docker/frr/ospf/eth"
	   "Test/docker/frr/ospf/vlan"
	   "Test/docker/gobgp/ebgp/eth"
	   "Test/docker/gobgp/ebgp/vlan"
	   "Test/docker/net/slice/vlan" # --- FAIL: Test/docker/net/slice/vlan/stress-pci (reboot to recover)
	   "Test/docker/net/dhcp/eth"
	   "Test/docker/net/dhcp/vlan"
	   "Test/docker/net/static/eth"
	   "Test/docker/net/static/vlan")

tester=../goes-platina-mk1.test

quit=0
fails=0

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
  ./xeth_util.sh test_init
}

if [ "$1" == "list" ]; then
    id=0
    for t in ${testcases[@]}; do
        id=$(($id+1))	
        echo $id ":" $t
    done
    echo
    grep -A 30 testcases\= $0 | grep \#
elif [ "$1" == "run" ]; then
    test_range=${testcases[@]}
elif [ "$1" == "run_range" ]; then
    shift
    start=$1
    start=$(($start-1))
    shift
    stop=$1
    stop=$(($stop-1))
    shift
    test_range=""
    for i in $(seq $start $stop); do
        test_range="${test_range} ${testcases[$i]}"
    done
else
    echo "list | run | run_range <start end>"
fi

test_count=$(echo $test_range | wc -w)

if [ $test_count != 0 ]; then
    echo "Running $test_count tests"
else
    exit 0
fi

if [ -z "$GOPATH" ]; then
    echo "GOPATH not set, try 'sudo -E ./$0 $*'"
    exit 1
fi

count=0
for t in ${test_range}; do
    log=${t//\//_}.out
    count=$(($count+1))
    printf "Running %46s " $t" ($count of $test_count) : "
    fix_it
    GOPATH=$GOPATH ${tester} -test.v -test.run=$t > /tmp/out 2>&1

    if [ $? == 0 ]; then
        echo "OK"
        mv /tmp/out ./$log.OK
    else
        mv /tmp/out ./$log
        if grep -q panic ./$log; then
            echo "Crashed"
        else
            echo "Failed"
        fi
	fails=$(($fails+1))
    fi

    if [ "$quit" -eq 1 ]; then
        echo "Aborted"
        break
    fi
done
echo
echo "$fails testcase(s) failed."

fix_it

exit 0
