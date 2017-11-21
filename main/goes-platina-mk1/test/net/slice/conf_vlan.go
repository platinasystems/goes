// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package slice

const ConfVlan = `
image: "stigt/debian-frr:latest"
volume: "/main/goes-platina-mk1/test/net/slice/"
mapping: "/etc/frr"
routers:
- hostname: CA-1
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 0}}
    address: 10.1.0.1/24
    vlan: 10
- hostname: RA-1
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.1.0.2/24
    vlan: 10
  - name: {{index . 3 0}}
    address: 10.2.0.2/24
    vlan: 20
- hostname: RA-2
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.2.0.3/24
    vlan: 20
  - name: {{index . 3 0}}
    address: 10.3.0.3/24
    vlan: 30
- hostname: CA-2
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.3.0.4/24
    vlan: 30
- hostname: CB-1
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 0}}
    address: 10.1.0.1/24
    vlan: 40
- hostname: RB-1
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.1.0.2/24
    vlan: 40
  - name: {{index . 3 0}}
    address: 10.2.0.2/24
    vlan: 50
- hostname: RB-2
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.2.0.3/24
    vlan: 50
  - name: {{index . 3 0}}
    address: 10.3.0.3/24
    vlan: 60
- hostname: CB-2
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 3 1}}
    address: 10.3.0.4/24
    vlan: 60
`
