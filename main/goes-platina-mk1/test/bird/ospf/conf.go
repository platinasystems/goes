// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ospf

const Conf = `
volume: "/testdata/bird/ospf/"
mapping: "/etc/bird"
routers:
- hostname: R1
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 2 1}}
    address: 192.168.120.5/24
  - name: {{index . 0 0}}
    address: 192.168.150.5/24
- hostname: R2
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 2 0}}
    address: 192.168.120.10/24
  - name: {{index . 1 0}}
    address: 192.168.222.10/24
- hostname: R3
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 0}}
    address: 192.168.111.2/24
  - name: {{index . 1 1}}
    address: 192.168.222.2/24
- hostname: R4
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 192.168.111.4/24
  - name: {{index . 0 1}}
    address: 192.168.150.4/24
`
