#!/bin/bash

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi
USER=$SUDO_USER

if [ -z "$1" ]; then
    echo "Usages: $0 up|down"
    exit 1
fi

D_MOVE=../docker_move.sh

case $1 in
    "up")
	docker-compose up -d

	ip link add dummy1 type dummy 2> /dev/null
	ip link add dummy2 type dummy 2> /dev/null
	
	$D_MOVE up CA-1 eth-4-0 10.1.0.1/24
	$D_MOVE up CA-1 dummy1  192.168.0.1/32 
	$D_MOVE up RA-1 eth-5-0 10.1.0.2/24
	$D_MOVE up RA-1 eth-14-0 10.2.0.2/24
	$D_MOVE up RA-2 eth-15-0 10.2.0.3/24
	$D_MOVE up RA-2 eth-24-0 10.3.0.3/24
	$D_MOVE up CA-2 eth-25-0 10.3.0.4/24
	$D_MOVE up CA-2 dummy2   192.168.0.2/32

	;;
    "down")
	$D_MOVE down CA-1 eth-4-0
	$D_MOVE down RA-1 eth-5-0
	$D_MOVE down RA-1 eth-14-0
	$D_MOVE down RA-2 eth-15-0
	$D_MOVE down RA-2 eth-24-0 
	$D_MOVE down CA-2 eth-25-0

	ip link del dummy1
	ip link del dummy2	

	docker-compose down
	
	for ns in CA-1 RA-1 RA-2 CA-2 ; do
	   ip netn del $ns
	done
	chown -R $USER:$USER volumes
	;;
    *)
	echo "Unknown action"
	;;
esac
