// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package static

const ConfVlan = `
volume: "/testdata/net/static/"
mapping: "/etc/frr"
routers:
- hostname: CA-1
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 10.1.0.1/24
    vlan: 10
  - name: dummy0
    address: 192.168.0.1/32
- hostname: RA-1
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 10.1.0.2/24
    vlan: 10
  - name: {{index . 0 0}}
    address: 10.2.0.2/24
    vlan: 20
- hostname: RA-2
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 10.2.0.3/24
    vlan: 20
  - name: {{index . 0 0}}
    address: 10.3.0.3/24
    vlan: 30
- hostname: CA-2
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 10.3.0.4/24
    vlan: 30
  - name: dummy0
    address: 192.168.0.2/32
`
