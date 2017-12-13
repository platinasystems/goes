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

# add Free Range Routing debian
ADD frr_3.0.2-1-debian8.1_amd64.deb /root/frr.deb

# install ffr
RUN DEBIAN_FRONTEND=noninteractive dpkg -i frr.deb

RUN rm -f /root/frr.deb

# Configuration files
VOLUME /etc/frr

# Expose ports
EXPOSE 179 2601 2604 2605 5201

# Command
CMD ["/root/startup.sh"]
