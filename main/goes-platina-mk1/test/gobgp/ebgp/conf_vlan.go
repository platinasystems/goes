// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ebgp

const ConfVlan = `
volume: "/testdata/gobgp/ebgp/"
mapping: "/etc/gobgp"
routers:
- hostname: R1
  image: "stigt/gobgp:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 192.168.120.5/24
    vlan: 10
  - name: {{index . 0 0}}
    address: 192.168.150.5/24
    vlan: 40
  - name: dummy0
    address: 192.168.1.5/32
- hostname: R2
  image: "stigt/gobgp:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.120.10/24
    vlan: 10
  - name: {{index . 0 0}}
    address: 192.168.222.10/24
    vlan: 20
  - name: dummy0
    address: 192.168.1.10/32
- hostname: R3
  image: "stigt/gobgp:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.150.2/24
    vlan: 40
  - name: {{index . 0 0}}
    address: 192.168.111.2/24
    vlan: 30
  - name: dummy0
    address: 192.168.2.2/32
- hostname: R4
  image: "stigt/gobgp:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.111.4/24
    vlan: 30
  - name: {{index . 0 1}}
    address: 192.168.222.4/24
    vlan: 20
  - name: dummy0
    address: 192.168.2.4/32
`
