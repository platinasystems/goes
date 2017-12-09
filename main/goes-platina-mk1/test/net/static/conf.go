// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package static

const Conf = `
volume: "/testdata/net/static/"
mapping: "/etc/frr"
routers:
- hostname: CA-1
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 0}}
    address: 10.1.0.1/24
- hostname: RA-1
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 0 1}}
    address: 10.1.0.2/24
  - name: {{index . 1 0}}
    address: 10.2.0.2/24
- hostname: RA-2
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 1 1}}
    address: 10.2.0.3/24
  - name: {{index . 2 0}}
    address: 10.3.0.3/24
- hostname: CA-2
  image: "stigt/debian-frr:latest"
  cmd: "/root/startup.sh"
  intfs:
  - name: {{index . 2 1}}
    address: 10.3.0.4/24
`
