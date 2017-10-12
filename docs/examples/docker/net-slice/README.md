## Multi-tenant network slice example

### Network Diagram

 ![OSPF example](https://github.com/platinasystems/go/blob/master/docs/examples/docker/net-slice/net-slice.jpeg?raw=true)

In this example there are 2 customers slices that both use the same network configuration and addresses, but are isolated using Docker containers.  This example also uses vlans, so only 1 loopback cable is needed to run it.  Currently the updown_vlan.sh script assumes eth-30-0 is connected by lookback to eth-31-0.  To change which interfaces are connected via loopback cable edit the following 2 variables in the updown_vlan.sh script:
```
# A loopback cable is connected between side A and B.
# All vlans go over this cable
SIDE_A=eth-30-0
SIDE_B=eth-31-0
```
To run the script:
```
sudo ./updown_vlan.sh up
```
Replace "up" with "down" to tear it all down.

Once the script it will create 8 containers and link them together as in the network diagram.  Enter container CA-1 with:
```
stig@invader1:~$ docker exec -it CA-1 bash
root@CA-1:~# 
root@CA-1:~# vtysh

Hello, this is FRRouting (version 3.1-dev).
Copyright 1996-2005 Kunihiro Ishiguro, et al.

CA-1# show ip route 
Codes: K - kernel route, C - connected, S - static, R - RIP,
       O - OSPF, I - IS-IS, B - BGP, P - PIM, E - EIGRP, N - NHRP,
       T - Table, v - VNC, V - VNC-Direct, A - Babel,
       > - selected route, * - FIB route

O   10.1.0.0/24 [110/10] is directly connected, eth-30-0.10, 00:00:39
C>* 10.1.0.0/24 is directly connected, eth-30-0.10, 00:00:39
```
The containers are running Free Range Routing, but initially we only see the 1 connected route.  After a bit we can see ospf has learned the path to the far 10.3.0.0/24 network:
```
CA-1# show ip route 
Codes: K - kernel route, C - connected, S - static, R - RIP,
       O - OSPF, I - IS-IS, B - BGP, P - PIM, E - EIGRP, N - NHRP,
       T - Table, v - VNC, V - VNC-Direct, A - Babel,
       > - selected route, * - FIB route

O   10.1.0.0/24 [110/10] is directly connected, eth-30-0.10, 00:00:50
C>* 10.1.0.0/24 is directly connected, eth-30-0.10, 00:01:30
O>* 10.2.0.0/24 [110/20] via 10.1.0.2, eth-30-0.10, 00:00:40
O>* 10.3.0.0/24 [110/30] via 10.1.0.2, eth-30-0.10, 00:00:38
```
```
CA-1# traceroute 10.3.0.4
traceroute to 10.3.0.4 (10.3.0.4), 30 hops max, 60 byte packets
 1  10.1.0.2 (10.1.0.2)  0.493 ms  0.488 ms  0.487 ms
 2  10.2.0.3 (10.2.0.3)  0.484 ms  0.479 ms  0.474 ms
 3  10.3.0.4 (10.3.0.4)  0.470 ms  0.468 ms  0.465 ms
```
To prove to that customer A is isolated from customer B, open a shell into CA-1, CA-2, CB-1 and CB-2.  Start a tcpdump on CA-2 and CB-2 and then ping 10.3.0.4 from CA-1 and you should only see the tcpdump on CA-2 to see the pings.
