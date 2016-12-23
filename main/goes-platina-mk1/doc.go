// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

/*

Notes on the configuration options (via redis) for machine platina-mk1-amd64:

General port nomencalture:

eth-PORT-SUBPORT

PORT - represents the physical plughole on the machine's physical front-panel
SUBPORT - represents the subordinate port and is based upon speeds and provisioning
          of the port e.g. for a 100 Gbps port only 1 subport is required so the
          port's name would be eth-PORT-0; while for a port with an expander cable
          one could configure 4 25 Gbps ports represented as eth-PORT-0 .. eth-PORT-3

Port provisioning:

"hset vnet.eth-PORT.provision <valid descriptor>

          where valid descriptor is a comma-separated string representing the subports and their speeds.
          Valid desciptor examples:

          "100" (eth-PORT-0 is 100 Gbps)
          "50"  (eth-PORT-0 is 50 Gbps)
          "40"  (eth-PORT-0 is 40 Gbps)
          "25"  (eth-PORT-0 is 25 Gbps)
          "20"  (eth-PORT-0 is 20 Gbps)
          "10"  (eth-PORT-0 is 10 Gbps)
          "1"  (eth-PORT-0 is 1 Gbps)

          "25,25,25,25" (eth-PORT-0 is 25 Gbps, eth-PORT-1 is 25 Gbps, eth-PORT-2 is 25Gbps, eth-PORT-3 is 25 Gbps)
          "20,20,20,20" (eth-PORT-0 is 20 Gbps, eth-PORT-1 is 20 Gbps, eth-PORT-2 is 20Gbps, eth-PORT-3 is 20 Gbps)
          "10,10,10,10" (eth-PORT-0 is 10 Gbps, eth-PORT-1 is 10 Gbps, eth-PORT-2 is 10Gbps, eth-PORT-3 is 10 Gbps)
          "1,1,1,1" (eth-PORT-0 is 1 Gbps, eth-PORT-1 is 1 Gbps, eth-PORT-2 is 1Gbps, eth-PORT-3 is 1 Gbps)

          "50,25,25" (eth-PORT-0 is 50Gbps, eth-PORT-2 is 25 Gbps, eth-PORT-3 is 25 Gbps)
          "25,25,50" (eth-PORT-0 is 25 Gbps, eth-PORT-1 is 25 Gbps, eth-PORT-2 is 50Gbps)
          "50,20,20" (eth-PORT-0 is 50Gbps, eth-PORT-2 is 20 Gbps, eth-PORT-3 is 20 Gbps)
          "20,20,50" (eth-PORT-0 is 20 Gbps, eth-PORT-1 is 20 Gbps, eth-PORT-2 is 50Gbps)
          "40,20,20" (eth-PORT-0 is 40Gbps, eth-PORT-2 is 20 Gbps, eth-PORT-3 is 20 Gbps)
          "20,20,40" (eth-PORT-0 is 20 Gbps, eth-PORT-1 is 20 Gbps, eth-PORT-2 is 40Gbps)
          "20,20,10,10" (eth-PORT-0 is 20 Gbps, eth-PORT-1 is 20 Gbps, eth-PORT-2 is 10Gbps, eth-PORT-3 is 10Gbps)

           The port's provisioning will be sent to the NPU driver and if valid will be applied and redis
           database updated. If invalid, an error will be returned to the setter and redis database will not be updated.
*/

package main
