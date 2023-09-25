## Using Docker Containers with Platina interfaces

### Manual Example 
In this directory is a **Dockerfile** that has the instructions for how to build a Docker container that uses debian Jessie has a base file system and then adds quagga and a few other networking tools.  To build the Docker image from the directory with the Dockerfile, run the command:
```
docker  build -t debian-quagga:latest .
```
After that the new image should up with the *docker images* command:
```
$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
debian-quagga       latest              506cf706ad55        43 seconds ago      257 MB
debian              jessie              86baf4e8cde9        4 days ago          123 MB
```
Note that it first pulled the official debian jessie image and then build debian-quagga on top of that.  To run the image:
```
$ docker run -d debian-quagga:latest
14de1a940b537807bec96ce33420eb190158ecffa30422382052381f7d56e568
```
To see if the container is now running:
```
$ docker ps -a
CONTAINER ID        IMAGE                  COMMAND                  CREATED              STATUS              PORTS                                        NAMES
14de1a940b53        debian-quagga:latest   "/usr/bin/supervis..."   About a minute ago   Up About a minute   179/tcp, 2601/tcp, 2604-2605/tcp, 5201/tcp   nostalgic_pare
```
To enter the container use the container ID that was shown with *docker ps -a* and run bash in the container:
```
$ docker exec -it 14de1a940b53 bash
root@14de1a940b53:/# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.0  46912 15096 ?        Ss   00:02   0:00 /usr/bin/python /usr/bin/supervisord -c /etc/
root        25  0.0  0.0  20248  3056 ?        Ss   00:05   0:00 bash
root        31  0.0  0.0  17504  2036 ?        R+   00:05   0:00 ps aux
root@14de1a940b53:/# exit
exit
```
To stop the container use the same container ID:
```
$ docker stop 14de1a940b53
14de1a940b53

$ docker ps -a
CONTAINER ID        IMAGE                  COMMAND                  CREATED             STATUS                     PORTS               NAMES
14de1a940b53        debian-quagga:latest   "/usr/bin/supervis..."   4 minutes ago       Exited (0) 6 seconds ago                       nostalgic_pare
```
Notice that even after stopping the container that it still shows up under docker ps.  That's because we didn't use the --rm option on the run command.  To clean up:
```
$ docker rm -v 14de1a940b53
14de1a940b53

$ docker ps -a
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```
Notice that quagga was not running when the container was entered.  That's because there was no quagga configuration.  This repo has some quagga configuration file in the volumes directory and we can mount that directory into the container with the following:
```
docker run --rm --privileged=true -P --hostname=R1 --name=R1 -d -v $(pwd)/volumes/quagga/R1:/etc/quagga debian-quagga:latest
```
The -v option says to map the local ./volumes/quagga/R1 in the container as /etc/quagga.  The --name option gave the container a name that can be used instead of the container id.  Now in the container zebra and ospf are running:
```
$ docker exec -it R1 bash
root@R1:/# 
root@R1:/# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.3  0.0  46924 15268 ?        Ss   00:20   0:00 /usr/bin/python /usr/bin/supervisord -c /etc/
quagga      10  0.0  0.0  28896  3576 ?        S    00:20   0:00 /usr/lib/quagga/ospfd -f /etc/quagga/ospfd.co
quagga      12  0.0  0.0  26256  3060 ?        S    00:20   0:00 /usr/lib/quagga/zebra -f /etc/quagga/zebra.co
root        19  0.0  0.0  20248  3072 ?        Ss   00:20   0:00 bash
root        25  0.0  0.0  17504  2068 ?        R+   00:20   0:00 ps aux
root@R1:/# 
root@R1:/# exit
exit
```
To move a Platina interface into the container the *docker_move.sh* script can be run:
```
$ sudo ./docker_move.sh up R1 eth-25-0 192.168.120.5/24
```
In the container the interface shows up:
```
$ docker exec -it R1 bash
root@R1:/# 
root@R1:/# ip addr show eth-25-0
1790: eth-25-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:46:8a:00:02:bc brd ff:ff:ff:ff:ff:ff
    inet 192.168.120.5/24 scope global eth-25-0
       valid_lft forever preferred_lft forever
    inet6 fe80::46:8aff:fe00:2bc/64 scope link 
       valid_lft forever preferred_lft forever
```
To move the interface from container R1 back to the host OS use the *down* option on the script:
```
$ sudo ./docker_move.sh down R1 eth-25-0 

$ ip add show eth-25-0
1790: eth-25-0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 02:46:8a:00:02:bc brd ff:ff:ff:ff:ff:ff
```
To really show OSPF working in the container, another container is needed and a loopback cable will be connected to those interfaces.  

The network to be built will be like this diagram with 4 routers

 ![OSPF example](https://github.com/platinasystems/go/blob/master/docs/examples/docker/quagga-ospf/ospf_example.jpeg?raw=true)

For multiple containers it would be easier to create a Docker compose file that has all the parameters for all 4 containers.  The file docker-compose.yaml defines all 4 containers and the parameters to run them.  An example for router R1 looks like:
```
version: '3'
services:
  R1:
    container_name: R1
    environment:
      - PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    expose:
      - 2601/tcp
      - 179/tcp
      - 5201/tcp
      - 2605/tcp
    hostname: R1
    image: stigt/debian-quagga:latest
    privileged: true
    volumes:
      - ./volumes/quagga/R1:/etc/quagga
    logging:
      options:
        max-size: "10m"
        max-file: "2"
```
  To start all the containers is as easy as:
```
$ docker-compose up -d
Creating network "docker_default" with the default driver
Creating R4 ... 
Creating R1 ... 
Creating R2 ... 
Creating R3 ... 
Creating R4
Creating R2
Creating R3
Creating R2 ... done

$ docker ps -a
CONTAINER ID        IMAGE                        COMMAND                  CREATED             STATUS              PORTS                                        NAMES
156ff8435b90        stigt/debian-quagga:latest   "/usr/bin/supervis..."   17 seconds ago      Up 15 seconds       179/tcp, 2601/tcp, 2604-2605/tcp, 5201/tcp   R1
1e5879fd6805        stigt/debian-quagga:latest   "/usr/bin/supervis..."   17 seconds ago      Up 15 seconds       179/tcp, 2601/tcp, 2604-2605/tcp, 5201/tcp   R3
6e4dcece2f8f        stigt/debian-quagga:latest   "/usr/bin/supervis..."   17 seconds ago      Up 15 seconds       179/tcp, 2601/tcp, 2604-2605/tcp, 5201/tcp   R2
d883e9bebefe        stigt/debian-quagga:latest   "/usr/bin/supervis..."   17 seconds ago      Up 15 seconds       179/tcp, 2601/tcp, 2604-2605/tcp, 5201/tcp   R4
```
And of course to stop those 4 containers:
```
$ docker-compose down
Stopping R1 ... done
Stopping R3 ... done
Stopping R2 ... done
Stopping R4 ... done
Removing R1 ... done
Removing R3 ... done
Removing R2 ... done
Removing R4 ... done
Removing network docker_default

$ docker ps -a
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```
Previously the docker_move.sh script was used to move a single interface into a container.  There is another script called updown.sh which calls docker_move.sh with the specific parameters to match the network diagram above and it will call docker-compose to start/stop the containers.
```
$ sudo ./updown.sh up
Creating network "docker_default" with the default driver
Creating R2 ... 
Creating R1 ... 
Creating R4 ... 
Creating R3 ... 
Creating R2
Creating R1
Creating R4
Creating R3 ... done

```
Then enter a container and see if ospf has learned the routes to all 4 routers:
```
$ docker exec -it R1 bash
root@R1:/# 
root@R1:/# vtysh

Hello, this is Quagga (version 0.99.23.1).
Copyright 1996-2005 Kunihiro Ishiguro, et al.

R1# show ip ospf neighbor 

    Neighbor ID Pri State           Dead Time Address         Interface            RXmtL RqstL DBsmL
192.168.1.1       1 Full/Backup       33.761s 192.168.120.10  eth-25-0:192.168.120.5     0     0     0
192.168.2.4       1 Full/DR           36.083s 192.168.150.4   eth-4-0:192.168.150.5     0     0     0


R1# show ip route ospf 
Codes: K - kernel route, C - connected, S - static, R - RIP,
       O - OSPF, I - IS-IS, B - BGP, A - Babel,
       > - selected route, * - FIB route

O   192.168.1.5/32 [110/10] is directly connected, dummy0, 00:02:13
O>* 192.168.1.10/32 [110/20] via 192.168.120.10, eth-25-0, 00:01:28
O>  192.168.2.2/32 [110/30] via 192.168.120.10, eth-25-0, 00:01:29
                            via 192.168.150.4, eth-4-0, 00:01:29
O>* 192.168.2.4/32 [110/20] via 192.168.150.4, eth-4-0, 00:01:31
O>* 192.168.111.0/24 [110/20] via 192.168.150.4, eth-4-0, 00:01:31
O   192.168.120.0/24 [110/10] is directly connected, eth-25-0, 00:02:14
O   192.168.150.0/24 [110/10] is directly connected, eth-4-0, 00:02:13
O>* 192.168.222.0/24 [110/20] via 192.168.120.10, eth-25-0, 00:01:29
R1#      
R1# ping 192.168.222.2
PING 192.168.222.2 (192.168.222.2): 56 data bytes
64 bytes from 192.168.222.2: icmp_seq=0 ttl=63 time=0.097 ms
64 bytes from 192.168.222.2: icmp_seq=1 ttl=63 time=0.103 ms
^C--- 192.168.222.2 ping statistics ---
2 packets transmitted, 2 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.097/0.100/0.103/0.000 ms
R1# 
```
Another script in this repo is updown_vlan.sh.  This script creates 6 vlans all on the same interface so only 1 loopback cable is required to run this ospf example.  The interfaces use for the loop can easily be changed in the script.  Look for:
```
# A loopback cable is connected between side A and B.
# All vlans go over this cable
SIDE_A=eth-4-0
SIDE_B=eth-5-0
```
To run:
```
$ sudo ./updown_vlan.sh up
Creating R4 ... 
Creating R1 ... 
Creating R2 ... 
Creating R3 ... 
Creating R4
Creating R3
Creating R2
Creating R1 ... done
```
Enter container R1 to see if ospf has is working:
```
root@invader1:~# docker exec -it R1 vtysh

Hello, this is Quagga (version 0.99.23.1).
Copyright 1996-2005 Kunihiro Ishiguro, et al.

R1# show ip route
Codes: K - kernel route, C - connected, S - static, R - RIP,
       O - OSPF, I - IS-IS, B - BGP, A - Babel,
       > - selected route, * - FIB route

C>* 127.0.0.0/8 is directly connected, lo
O   192.168.1.5/32 [110/10] is directly connected, dummy0, 00:07:48
C>* 192.168.1.5/32 is directly connected, dummy0
O>* 192.168.1.10/32 [110/20] via 192.168.120.10, eth-4-0.10, 00:06:57
O>  192.168.2.2/32 [110/30] via 192.168.120.10, eth-4-0.10, 00:06:48
                            via 192.168.150.4, eth-5-0.40, 00:06:48
O>* 192.168.2.4/32 [110/20] via 192.168.150.4, eth-5-0.40, 00:06:53
O>  192.168.50.0/24 [110/30] via 192.168.120.10, eth-4-0.10, 00:06:48
                             via 192.168.150.4, eth-5-0.40, 00:06:48
O>* 192.168.60.0/24 [110/20] via 192.168.150.4, eth-5-0.40, 00:06:53
O>* 192.168.111.0/24 [110/20] via 192.168.150.4, eth-5-0.40, 00:06:53
O   192.168.120.0/24 [110/10] is directly connected, eth-4-0.10, 00:07:48
C>* 192.168.120.0/24 is directly connected, eth-4-0.10
O   192.168.150.0/24 [110/10] is directly connected, eth-5-0.40, 00:07:48
C>* 192.168.150.0/24 is directly connected, eth-5-0.40
O>* 192.168.222.0/24 [110/20] via 192.168.120.10, eth-4-0.10, 00:06:58
```
In the container we only see the routing table for R1, but from goes we can see the default namespace and the 4 container's namespaces.
```
root@invader1:~# goes show vnet ip fib
 Table                   Destination                               Adjacency
     default                  10.15.0.0/24       3: glean eth-0-0
     default                  10.15.0.1/32       4: local eth-0-0
     default                  10.50.0.0/24       9: glean eth-21-0
     default                  10.50.0.1/32      10: local eth-21-0
docker-2239fb0b2a04               192.168.1.10/32      46: rewrite eth-5-0.20 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 20 46-46, 1 x 36
docker-2239fb0b2a04                192.168.2.2/32       2: punt
docker-2239fb0b2a04                192.168.2.4/32      44: rewrite eth-4-0.30 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 30 44-44, 1 x 31
docker-2239fb0b2a04               192.168.50.0/24      23: glean eth-5-0.50
docker-2239fb0b2a04               192.168.50.2/32      24: local eth-5-0.50
docker-2239fb0b2a04               192.168.60.0/24      44: rewrite eth-4-0.30 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 30 44-44, 1 x 31
docker-2239fb0b2a04              192.168.111.0/24      19: glean eth-4-0.30
docker-2239fb0b2a04              192.168.111.2/32      20: local eth-4-0.30
docker-2239fb0b2a04              192.168.111.4/32      31: rewrite eth-4-0.30 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 30
docker-2239fb0b2a04              192.168.120.0/24      46: rewrite eth-5-0.20 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 20 46-46, 1 x 36
docker-2239fb0b2a04              192.168.150.0/24      44: rewrite eth-4-0.30 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 30 44-44, 1 x 31
docker-2239fb0b2a04              192.168.222.0/24      21: glean eth-5-0.20
docker-2239fb0b2a04              192.168.222.2/32      22: local eth-5-0.20
docker-2239fb0b2a04             192.168.222.10/32      36: rewrite eth-5-0.20 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 20
docker-01575d264139                192.168.1.5/32      40: rewrite eth-4-0.40 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 40 40-40, 1 x 33
docker-01575d264139                192.168.2.2/32      41: rewrite eth-5-0.30 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 30 41-41, 1 x 34
docker-01575d264139                192.168.2.4/32       2: punt
docker-01575d264139               192.168.50.0/24      41: rewrite eth-5-0.30 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 30 41-41, 1 x 34
docker-01575d264139               192.168.60.0/24      29: glean eth-5-0.60
docker-01575d264139               192.168.60.4/32      30: local eth-5-0.60
docker-01575d264139              192.168.111.0/24      25: glean eth-5-0.30
docker-01575d264139              192.168.111.2/32      34: rewrite eth-5-0.30 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 30
docker-01575d264139              192.168.111.4/32      26: local eth-5-0.30
docker-01575d264139              192.168.120.0/24      40: rewrite eth-4-0.40 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 40 40-40, 1 x 33
docker-01575d264139              192.168.150.0/24      27: glean eth-4-0.40
docker-01575d264139              192.168.150.4/32      28: local eth-4-0.40
docker-01575d264139              192.168.150.5/32      33: rewrite eth-4-0.40 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 40
docker-01575d264139              192.168.222.0/24      41: rewrite eth-5-0.30 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 30 41-41, 1 x 34
docker-b6c09eb742d5                192.168.1.5/32      39: rewrite eth-5-0.10 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 10 39-39, 1 x 38
docker-b6c09eb742d5               192.168.1.10/32       2: punt
docker-b6c09eb742d5                192.168.2.2/32      45: rewrite eth-4-0.20 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 20 45-45, 1 x 35
docker-b6c09eb742d5               192.168.50.0/24      45: rewrite eth-4-0.20 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 20 45-45, 1 x 35
docker-b6c09eb742d5              192.168.111.0/24      45: rewrite eth-4-0.20 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 20 45-45, 1 x 35
docker-b6c09eb742d5              192.168.120.0/24      15: glean eth-5-0.10
docker-b6c09eb742d5              192.168.120.5/32      38: rewrite eth-5-0.10 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 10
docker-b6c09eb742d5             192.168.120.10/32      16: local eth-5-0.10
docker-b6c09eb742d5              192.168.150.0/24      39: rewrite eth-5-0.10 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 10 39-39, 1 x 38
docker-b6c09eb742d5              192.168.222.0/24      17: glean eth-4-0.20
docker-b6c09eb742d5              192.168.222.2/32      35: rewrite eth-4-0.20 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 20
docker-b6c09eb742d5             192.168.222.10/32      18: local eth-4-0.20
docker-21c38ca5a520                192.168.1.5/32       2: punt
docker-21c38ca5a520               192.168.1.10/32      42: rewrite eth-4-0.10 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 10 42-42, 1 x 37
docker-21c38ca5a520                192.168.2.4/32      43: rewrite eth-5-0.40 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 40 43-43, 1 x 32
docker-21c38ca5a520               192.168.60.0/24      43: rewrite eth-5-0.40 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 40 43-43, 1 x 32
docker-21c38ca5a520              192.168.111.0/24      43: rewrite eth-5-0.40 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 40 43-43, 1 x 32
docker-21c38ca5a520              192.168.120.0/24      11: glean eth-4-0.10
docker-21c38ca5a520              192.168.120.5/32      12: local eth-4-0.10
docker-21c38ca5a520             192.168.120.10/32      37: rewrite eth-4-0.10 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 10
docker-21c38ca5a520              192.168.150.0/24      13: glean eth-5-0.40
docker-21c38ca5a520              192.168.150.4/32      32: rewrite eth-5-0.40 IP4: 02:46:8a:00:02:ae -> 02:46:8a:00:02:b2 vlan 40
docker-21c38ca5a520              192.168.150.5/32      14: local eth-5-0.40
docker-21c38ca5a520              192.168.222.0/24      42: rewrite eth-4-0.10 IP4: 02:46:8a:00:02:b2 -> 02:46:8a:00:02:ae vlan 10 42-42, 1 x 37
```
