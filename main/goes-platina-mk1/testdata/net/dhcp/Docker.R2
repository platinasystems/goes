# HEADER
FROM        debian:jessie
MAINTAINER  stigt

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    arping \
    isc-dhcp-server \
    hping3 \
    init-system-helpers \    
    iperf3 \
    iproute \
    libc-ares2 \
    libjson-c2 \
    logrotate \
    net-tools \
    supervisor \    
    telnet \
    tcpdump \    
    traceroute\
    vim      

WORKDIR /root

ADD         startup.sh /root/startup.sh

ADD         supervisord.R2.conf /etc/supervisord.conf

ADD	    dhcpd.conf /root/dhcpd.conf

RUN	    touch /root/dhcpd.leases

RUN	    echo 'INTERFACES="eth-25-0 eth-5-0.10"' > /etc/default/isc-dhcp-server

# Command
CMD ["/root/startup.sh"]
