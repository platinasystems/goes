#!/bin/bash

set -x -e

/sbin/ifdown $1
/sbin/ifup $1

/usr/bin/goes vnet set ha sp $1 auto
/usr/bin/goes vnet set ha sp $1 100g
/usr/bin/goes vnet set ha sp $1 auto dis
/usr/bin/goes vnet set ha sp $1 auto ena

