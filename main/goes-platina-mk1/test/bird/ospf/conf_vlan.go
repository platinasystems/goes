// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ospf

const ConfVlan = `
volume: "/testdata/bird/ospf/"
mapping: "/etc/bird"
routers:
- hostname: R1
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 192.168.120.5/24
    vlan: 10
  - name: {{index . 0 1}}
    address: 192.168.150.5/24
    vlan: 40
  - name: {{index . 0 0}}
    address: 192.168.150.5/24
    vlan: 50
- hostname: R2
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.120.10/24
    vlan: 10
  - name: {{index . 0 0}}
    address: 192.168.222.10/24
    vlan: 20
  - name: {{index . 0 0}}
    address: 192.168.60.10/24
    vlan: 60
- hostname: R3
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 192.168.111.2/24
    vlan: 30
  - name: {{index . 0 1}}
    address: 192.168.222.2/24
    vlan: 20
  - name: {{index . 0 1}}
    address: 192.168.50.2/24
    vlan: 50
- hostname: R4
  image: "stigt/bird:v2.0.0"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.111.4/24
    vlan: 30
  - name: {{index . 0 0}}
    address: 192.168.150.4/24
    vlan: 40
  - name: {{index . 0 1}}
    address: 192.168.60.4/24
    vlan: 60
`
