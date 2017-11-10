ifconfig lo 192.168.32.1 netmask 255.255.255.0
route add -net 192.168.29.0 netmask 255.255.255.0 gw 10.0.19.29
route add -net 192.168.30.0 netmask 255.255.255.0 gw 10.0.3.30
