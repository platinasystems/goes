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

	$D_MOVE up R1 eth-24-0 
	$D_MOVE up R2 eth-25-0 192.168.120.10/24
	;;
    "down")
	$D_MOVE down R1 eth-24-0
	$D_MOVE down R2 eth-25-0		

	docker-compose down
	
	for ns in R1 R2 ; do
	   ip netn del $ns
	done
	;;
    *)
	echo "Unknown action"
	;;
esac
