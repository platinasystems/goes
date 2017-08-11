#!/bin/bash

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi

if [ -z "$1" ]; then
    echo "Usages: $0 up|down"
    exit 1
fi

case $1 in
    "up")
	ip link add dummy0 type dummy 2> /dev/null
	ip link add dummy1 type dummy 2> /dev/null
	ip link add dummy2 type dummy 2> /dev/null
	ip link add dummy3 type dummy 2> /dev/null

	./docker_move.sh up R1 eth-25-0 192.168.120.5/24
	./docker_move.sh up R1 eth-4-0 192.168.150.5/24
	./docker_move.sh up R1 dummy0 192.168.1.5/32

	./docker_move.sh up R2 eth-24-0 192.168.120.10/24
	./docker_move.sh up R2 eth-14-0 192.168.222.10/24
	./docker_move.sh up R2 dummy1 192.168.1.10/32

	./docker_move.sh up R3 eth-30-0 192.168.111.2/24
	./docker_move.sh up R3 eth-15-0 192.168.222.2/24
	./docker_move.sh up R3 dummy2 192.168.2.2/32

	./docker_move.sh up R4 eth-31-0 192.168.111.4/24
	./docker_move.sh up R4 eth-5-0 192.168.150.4/24
	./docker_move.sh up R4 dummy3 192.168.2.4/32
	;;
    "down")
	./docker_move.sh down R1 eth-25-0
	./docker_move.sh down R1 eth-4-0
	./docker_move.sh down R1 dummy0

	./docker_move.sh down R2 eth-24-0
	./docker_move.sh down R2 eth-14-0
	./docker_move.sh down R2 dummy1

	./docker_move.sh down R3 eth-30-0
	./docker_move.sh down R3 eth-15-0
	./docker_move.sh down R3 dummy2

	./docker_move.sh down R4 eth-31-0
	./docker_move.sh down R4 eth-5-0
	./docker_move.sh down R4 dummy3
	;;
    *)
	echo "Unknown action"
	;;
esac
