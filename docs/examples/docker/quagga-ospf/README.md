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

 ![OSPF example](https://github.com/platinasystems/go/blob/master/docs/examples/docker/ospf_example.jpeg?raw=true)

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
Previously the docker_move.sh script was used to move a single interface into a container.  There is another script called updown.sh which calls docker_move.sh with the specific parameters to match the network diagram above.  For the next example docker-compose will be used to start the 4 router containers and then updown.sh will be run to move/configure all the interfaces.
```
$ docker-compose up -d
Creating network "docker_default" with the default driver
Creating R2 ... 
Creating R1 ... 
Creating R4 ... 
Creating R3 ... 
Creating R2
Creating R1
Creating R4
Creating R3 ... done

$ sudo ./updown.sh up
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
