#!/bin/bash
set -x -e

# This build will put binaries in /root/workspace/go/bin/
# because playbook 'becomes' root.
source /home/platina/.profile
/usr/local/go/bin/go get -v -x -u github.com/osrg/gobgp/gobgp
/usr/local/go/bin/go get -v -x -u github.com/osrg/gobgp/gobgpd

cp /root/workspace/go/bin/gobgp /usr/local/bin/
cp /root/workspace/go/bin/gobgpd /usr/local/bin/
