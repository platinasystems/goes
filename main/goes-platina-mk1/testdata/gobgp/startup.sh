#!/bin/bash

cp /etc/gobgp/zebra.conf /etc/quagga
chown -R quagga:quagga /etc/quagga

chown -R gobgpd:gobgpd /etc/gobgp
chmod 644 /etc/gobgp/*
chmod 644 /etc/quagga/*

chown gobgpd:gobgpd /root/go*

exec /usr/bin/supervisord -c /etc/supervisord.conf
