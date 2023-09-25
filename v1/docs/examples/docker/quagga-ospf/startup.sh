#!/bin/bash

chown -R quagga:quagga /etc/quagga
chmod 644 /etc/quagga/*
exec /usr/bin/supervisord -c /etc/supervisord.conf
