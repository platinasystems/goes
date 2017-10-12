#!/bin/bash

chown -R frr:frr /etc/frr
chmod 644 /etc/frr/*
exec /usr/bin/supervisord -c /etc/supervisord.conf
