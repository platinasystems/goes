#!/bin/bash

chown -R bird:bird /etc/bird
chmod 644 /etc/bird/*
exec /usr/bin/supervisord -c /etc/supervisord.conf
