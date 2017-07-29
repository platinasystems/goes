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

The network to be built will be like this diagram:
