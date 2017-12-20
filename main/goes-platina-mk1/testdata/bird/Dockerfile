# HEADER
FROM        debian:jessie
MAINTAINER  stigt

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    arping \
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

ADD         supervisord.conf /etc/supervisord.conf

# add BIRD routing daemon and client
ADD bird /root/bird
ADD birdc /root/birdc
ADD birdcl /root/birdcl

# Configuration files
VOLUME /etc/bird

# Expose ports
EXPOSE 179 2601 2604 2605 5201

# Command
CMD ["/root/startup.sh"]
