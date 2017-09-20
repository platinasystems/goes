#!/bin/bash

STOP=0
if [ $# -eq 1 ]
then
  STOP=$1
fi

INTF=eth-1-1   # interface to check counter
LOOPS=1

while true
do
  logger -f /var/log/syslog "START iteration $LOOPS"
  date
  echo "interation $LOOPS"
  echo "stopping"
  goes stop
  if [ $? -ne 0 ]
  then
    echo "Failed to stop"
    exit 1
  fi
  echo "starting"
  goes start
  if [ $? -ne 0 ]
  then
    echo "Failed to start"
    exit 2
  fi
  NOTDONE=1
  while [ "$NOTDONE" -eq 1 ]
  do
    VAL=$(goes hget platina vnet.ready)
    if [ "$VAL" == "true" ]
    then
       NOTDONE=0
    else
       echo "vnet not ready"
       sleep 1
    fi
  done
  echo "vnet.ready = true"
  VAL=$(ps aux | grep vnet | grep -v grep | wc -l)
  if [ "$VAL" -ne 1 ]
  then
     echo "vnet not running?"
     exit 3
  fi
  echo "reading vnet counters"
  IFS=$'\n'
  array=$(timeout 5 goes vnet show interface $INTF)
  if [ $? -ne 0 ]
  then
    echo "Failed to get counter - vnet hung?"
    unset IFS
    exit 4
  fi
  NUM=foo
  for t in $array
  do
    if [[ "$t" =~ "tx bytes" ]]
    then
      if [[ "$t" =~ ([0-9]+)$ ]]
      then
        NUM=${BASH_REMATCH[1]}
      fi
    fi
  done
  unset IFS
  if [ "$NUM" != "foo" ]
  then
    VAL=$NUM
  fi
  echo "val = $VAL"
  if [ "$VAL" -eq 0 ]
  then
     echo "Traffic not flowing?"
     exit 6
  fi
  echo -n "Add route - "
  ip link set up $INTF
  ip add add 10.9.1.16/24 dev $INTF
  ip neigh add 10.9.1.17 lladdr 6c:ec:5a:07:c8:ba dev $INTF
  ip ro add 2.2.0.0/16 via 10.9.1.17
  VAL=$(goes vnet show ip fib | grep 2.2.0.0/16 | wc -l)
  if [ "$VAL" -eq "1" ]
  then
     echo "OK"
  else
     echo "Not OK, $VAL"
  fi
  echo -n "Del route - "
  ip ro del 2.2.0.0/16 via 10.9.1.17
  VAL=$(goes vnet show ip fib | grep 2.2.0.0/16 | wc -l)
  if [ "$VAL" -eq "0" ]
  then
     echo "OK"
  else
     echo "Not OK, $VAL"
  fi
  logger -f /var/log/syslog "END iteration $LOOPS"

  if [ "$LOOPS" -eq "$STOP" ]
  then
     echo Done
     exit 0
  fi
  ((LOOPS++))
  sleep 1
  echo
done
