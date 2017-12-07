// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dhcp

const ConfVlan = `
routers:
- hostname: R1
  image: "stigt/debian-dhcpc:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 192.168.120.5/24
- hostname: R2
  image: "stigt/debian-dhcps:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 192.168.120.10/24
`
