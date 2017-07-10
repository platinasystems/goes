# HEADER
FROM        debian:jessie
MAINTAINER  stigt

RUN apt-get update && apt-get install -y iperf3 tcpdump quagga supervisor hping3 vim arping net-tools telnet traceroute

# Supervisord
ADD         supervisord.conf /etc/supervisord.conf

# Configuration files
VOLUME /etc/quagga

# Fix ownership/permissions
RUN chown -R quagga:quagga /etc/quagga

# Expose ports
EXPOSE 179 2601 2604 2605 5201

# Command
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
