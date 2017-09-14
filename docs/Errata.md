This is a list of known problems in the latest code. This is a living document that will
change over time as issues are resolved.

# Broken Functionality

## Tuntap mode:

- punt path iperf3 issues inv4->inv5 (transmitter locks up)
- IF_OPER status at linux tuntap not reflecting actual link-state (on bug list)
- adding /31 route breaks routing (can be reproduced easily - needs to be fixed)
- multipath stops routing (4 ns rig - script ns2b.sh - needs to be fixed)
- multipath adjacency handling breaks gre (Stig's 4 inv rig - 2 paths one gre the other just eth - needs to be fixed)
- gre-tunnels: Having trouble getting a script working here. What is correct
	config for this case? See nsgrevnet.inv[5,4].sh

- panic by vnetd hangs it (not exiting)
- tuntap gets in mode where goes restart always fails with 
	panic: pci 0000:03:00.0 Intel 0x15ab: open /dev/vfio/13: device or resource busy

## SRIOV-mode:

- panics on every goes start (adddelnexthop issue)
- iperf load breaks punt-path
- punt performance only 1 Gbps (punting through PCI)
- Issue 78 (Donn) 

# Known panics

1. tuntap mode: Panic in (*Fib).addDelRouteNextHop - can reproduce by putting interface admin-down in a namespace
Aug 23 12:11:41 invader1 goes.vnetd[3608]: runtime error: slice bounds out of range: goroutine 40 [running]:
Aug 23 12:11:41 invader1 goes.vnetd[3608]: runtime/debug.Stack(0xc4211f95d0, 0xb53000, 0x1370010)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /usr/local/go/src/runtime/debug/stack.go:24 +0x79
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/elib/loop.(*Loop).doEvent.func1(0xc420164000)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/elib/loop/event.go:126 +0x72
Aug 23 12:11:41 invader1 goes.vnetd[3608]: panic(0xb53000, 0x1370010)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip.(*Main).AddDelNextHop(0xc4201e6068, 0x1400000013, 0xc400000001, 0x7f174ad06498, 0xc42024c7c0, 0x201f000001, 0xc42020a601)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip/adjacency.go:584 +0x280
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelRouteNextHop(0xc420718280, 0xc4201e6000, 0xc420435b80, 0xc41f000001, 0x14cfce0, 0xc42024c7c0, 0x91f901, 0xc4201e6518, 0x4)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:704 +0x34b
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Fib).replaceNextHop(0xc420718280, 0xc4201e6000, 0xc420435b80, 0xc420718280, 0xf00000014, 0x1f000001, 0x14cfce0, 0xc42024c7c0, 0x3, 0xc420190ea0)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:730 +0x1e4
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*mapFibResult).replaceWithLessSpecific(0xc4211f9aa0, 0xc4201e6000, 0xc420718280, 0xc4211f9a90)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:385 +0x26f
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable(0xc420718280, 0xc4201e6000, 0xc4204c0dc8, 0x100000014)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:450 +0x10a
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel(0xc420718280, 0xc4201e6000, 0xc4204c0dc8, 0xbeb2da0100000014, 0xc4211f9b88)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:274 +0x413
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute(0xc4201e6000, 0xc420eefa40, 0x1400000003, 0x8f8a01, 0xc4201e60b0, 0x7f, 0x3)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/fib.go:633 +0xf4
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ip4.(*Main).(github.com/platinasystems/go/vnet/ip4.addDelRoute)-fm(0xc420eefa40, 0x1400000003, 0xc400000001, 0xc4201652e0, 0x1, 0xc)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ip4/package.go:28 +0x4d
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor(0xc4201f0068, 0xc4201e6068, 0xc420eefa20, 0x7f1f000001, 0xc400000001, 0x94effd, 0xc421a82000)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/ethernet/neighbor.go:91 +0x283
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).ip4NeighborMsg(0xc420715530, 0xc42020c1c0, 0xc42020c1c0, 0x1)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/unix/netlink.go:549 +0x27c
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).EventAction(0xc420715530)
Aug 23 12:11:41 invader1 goes.vnetd[3608]:         /home/stig/go/src/github.com/platinasystems/go/vnet/unix/netlink.go:431 +0x682
Aug 23 12:11:41 invader1 goes.vnetd[3608]: github.com/platinasystems/go/elib/loop.(*loopEvent).do(0xc420715560)

2. tuntap mode: Panic in MapFib.getLessSpecific()

Aug 23 10:24:51 invader5 goes.vnetd[1916]: panic(0xb7cac0, 0x13a2000)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/ip4.(*MapFib).getLessSpecific(0xc421946008, 0xc4201c7790, 0x0, 0xc4200a62e8, 0x200000a, 0xc4201c7790)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:485 +0x12b
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable(0xc421946000, 0xc420211000, 0xc4201c7790, 0x2)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:447 +0xb1
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel(0xc421946000, 0xc420211000, 0xc4201c7790, 0x43ad630000000002, 0x10000c4224dfcb8)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:314 +0x258
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute(0xc420211000, 0xc42053cba0, 0x200000001, 0x0, 0xbbae80, 0xc421c1e101, 0xc42053cba0)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:633 +0xf4
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/ip4.(*Main).(github.com/platinasystems/go/vnet/ip4.addDelRoute)-fm(0xc42053cba0, 0x200000001, 0x0, 0xc420149d78, 0x1, 0xc420149d78)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/package.go:28 +0x4d
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).ip4IfaddrMsg(0xc4201130e0, 0xc421c1e420, 0xc421c1e420, 0x1)
Aug 23 10:24:51 invader5 goes.vnetd[1916]:         /home/dlobete/src/github.com/platinasystems/go/vnet/unix/netlink.go:507 +0x208
Aug 23 10:24:51 invader5 goes.vnetd[1916]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).EventAction(0xc4201130e0)

3. tuntap mode: Panic in (*DmaRequest).l3_generic_interface_admin_up_down

Aug 24 16:14:49 invader5 goes.vnetd[741]: runtime error: index out of range: goroutine 37 [running]:
Aug 24 16:14:49 invader5 goes.vnetd[741]: runtime/debug.Stack(0xc4205e79f0, 0xb5b4e0, 0x13796a0)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /usr/local/go/src/runtime/debug/stack.go:24 +0x79
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/go/elib/loop.(*Loop).doEvent.func1(0xc4201b2000)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/event.go:126 +0x72
Aug 24 16:14:49 invader5 goes.vnetd[741]: panic(0xb5b4e0, 0x13796a0)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/fe1/internal/fe1a.(*DmaRequest).l3_generic_interface_admin_up_down(0xc420915930, 0x100000088, 0xc4200b58c8, 0xc421229920)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3_interface.go:672 +0x92d
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).swIfAdminUpDown(0xc420900000, 0xc4201b2000, 0x100000088, 0x0, 0x3)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3_interface.go:596 +0xf0
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).(github.com/platinasystems/fe1/internal/fe1a.swIfAdminUpDown)-fm(0xc4201b2000, 0x100000088, 0xc42031ed80, 0x901047)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3.go:30 +0x45
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/go/vnet.(*SwIf).SetAdminUp(0xc4203bc080, 0xc4201b2000, 0xc42020c001, 0x88, 0xc4205e8000)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/go/vnet/interface.go:335 +0xa5
Aug 24 16:14:49 invader5 goes.vnetd[741]: github.com/platinasystems/go/vnet.Si.SetAdminUp(0xc400000088, 0xc4201b2000, 0xc4205e8001, 0xc4205e8000, 0x1)
Aug 24 16:14:49 invader5 goes.vnetd[741]:         /home/dlobete/src/github.com/platinasystems/go/vnet/interface.go:345 +0x63

A related panic:

Aug 25 12:11:07 invader5 goes.vnetd[2079]: delete zero disposition: goroutine 37 [running]:
Aug 25 12:11:07 invader5 goes.vnetd[2079]: runtime/debug.Stack(0xc420e8d9a8, 0xafc360, 0xc420c04a30)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /usr/local/go/src/runtime/debug/stack.go:24 +0x79
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/go/elib/loop.(*Loop).doEvent.func1(0xc420160000)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/event.go:126 +0x72
Aug 25 12:11:07 invader5 goes.vnetd[2079]: panic(0xafc360, 0xc420c04a30)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/go/vnet/ethernet.(*vlan_tagged_punt_node).del_disposition(0xc4201c60b0, 0xc400000000, 0xa6b060)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ethernet/punt_1tag.go:55 +0xa1
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/go/vnet/ethernet.(*SingleTaggedPuntNode).DelDisposition(0xc4201c60b0, 0xc400000000, 0xc420e8db50)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ethernet/punt_1tag.go:66 +0x33
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/fe1/internal/fe1a.(*DmaRequest).l3_generic_interface_admin_up_down(0xc420721930, 0xf, 0xc4213d6458, 0xc4204e03f0)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3_interface.go:708 +0x1dd
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).swIfAdminUpDown(0xc42070c000, 0xc420160000, 0xf, 0x0, 0x3)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3_interface.go:596 +0xf0
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).(github.com/platinasystems/fe1/internal/fe1a.swIfAdminUpDown)-fm(0xc420160000, 0xf, 0xc420132f70, 0x901047)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/l3.go:30 +0x45
Aug 25 12:11:07 invader5 goes.vnetd[2079]: github.com/platinasystems/go/vnet.(*SwIf).SetAdminUp(0xc4201654f0, 0xc420160000, 0xc4201bd000, 0xf, 0xc42129c380)
Aug 25 12:11:07 invader5 goes.vnetd[2079]:         /home/dlobete/src/github.com/platinasystems/go/vnet/interface.go:335 +0xa5

4. tuntap mode: Panic in rx punt path running iperf3
Aug 31 14:22:26 invader11 goes.vnetd[1280]: panic: packet too large
Aug 31 14:22:26 invader11 goes.vnetd[1280]: 
Aug 31 14:22:26 invader11 goes.vnetd[1280]: goroutine 53 [running]:
Aug 31 14:22:26 invader11 goes.vnetd[1280]: github.com/platinasystems/go/vnet.(*interfaceNode).slowPath(0xc420263688, 0xc42
1474020, 0xc4219fbb00, 0xc420e66c00, 0x8, 0x120, 0xc421474030, 0x100, 0x100, 0x1, ...)
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/vnet/interface_node.go:2
96 +0x47e
Aug 31 14:22:26 invader11 goes.vnetd[1280]: github.com/platinasystems/go/vnet.(*interfaceNode).ifOutput(0xc420263688, 0xc42
1474020)
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/vnet/interface_node.go:2
22 +0x599
Aug 31 14:22:26 invader11 goes.vnetd[1280]: github.com/platinasystems/go/vnet.(*interfaceNode).LoopOutput(0xc420263688, 0xc
420226000, 0x158a580, 0xc421474020)
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/vnet/interface_node.go:4
9 +0x4f
Aug 31 14:22:26 invader11 goes.vnetd[1280]: github.com/platinasystems/go/elib/loop.(*Out).call(0xc420507200, 0xc420226000, 
0xc42027a360, 0xc420507200)
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/elib/loop/call.go:485 +0
x29a
Aug 31 14:22:26 invader11 goes.vnetd[1280]: github.com/platinasystems/go/elib/loop.(*Loop).dataPoll(0xc420226000, 0x7fc9007
f01e0, 0xc42028c1c8)
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/elib/loop/loop.go:350 +0
x15e
Aug 31 14:22:26 invader11 goes.vnetd[1280]: created by github.com/platinasystems/go/elib/loop.(*Loop).startDataPoller
Aug 31 14:22:26 invader11 goes.vnetd[1280]:         /home/stig/go/src/github.com/platinasystems/go/elib/loop/loop.go:361 +0
xe5
Aug 31 14:22:28 invader11 goes.vnetd[1280]: exit status 2

5. tuntap mode: Running iperf3 on rx punt path
Aug 31 16:24:36 invader11 goes.vnetd[3516]: fatal error: runtime: out of memory
Aug 31 16:24:36 invader11 goes.vnetd[3516]:
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime stack:
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.throw(0xcfec9b, 0x16)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/panic.go:596 +0x95
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.sysMap(0xc801c00000, 0x180000000, 0x0, 0x15f6bd8)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mem_linux.go:216 +0x1d0
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).sysAlloc(0x15d4840, 0x180000000, 0x7f296116fce0)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/malloc.go:428 +0x374
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).grow(0x15d4840, 0xc0000, 0x0)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mheap.go:774 +0x62
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).allocSpanLocked(0x15d4840, 0xc0000, 0xc4200e9040)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mheap.go:678 +0x44f
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).alloc_m(0x15d4840, 0xc0000, 0x100000000, 0x0)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mheap.go:562 +0xe2
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).alloc.func1()
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mheap.go:627 +0x4b
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.systemstack(0x7f296116fdd8)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/asm_amd64.s:343 +0xab
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.(*mheap).alloc(0x15d4840, 0xc0000, 0x10100000000, 0x0)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/mheap.go:628 +0xa0
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.largeAlloc(0x180000000, 0x451401, 0x10000c4202aa270)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/malloc.go:795 +0x93
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.mallocgc.func1()
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/malloc.go:690 +0x3e
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.systemstack(0xc420014600)
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/asm_amd64.s:327 +0x79
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.mstart()
Aug 31 16:24:36 invader11 goes.vnetd[3516]:         /usr/local/go/src/runtime/proc.go:1132
Aug 31 16:24:36 invader11 goes.vnetd[3516]:
Aug 31 16:24:36 invader11 goes.vnetd[3516]: goroutine 10 [running]:
Aug 31 16:24:36 invader11 goes.vnetd[3516]: runtime.systemstack_switch()

6. tuntap mode
Sep  6 15:57:57 invader5 goes.vnetd[1217]: panic: runtime error: invalid memory address or nil pointer dereference [recovered]
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         panic: runtime error: invalid memory address or nil pointer dereference
Sep  6 15:57:57 invader5 goes.vnetd[1217]: [signal SIGSEGV: segmentation violation code=0x1 addr=0x48 pc=0x90eadd]
Sep  6 15:57:57 invader5 goes.vnetd[1217]:
Sep  6 15:57:57 invader5 goes.vnetd[1217]: goroutine 52 [running]:
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/elib/loop.(*Loop).dataPoll.func1()
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/poller.go:336 +0x52
Sep  6 15:57:57 invader5 goes.vnetd[1217]: panic(0xbc5620, 0x143e190)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/elib/loop.(*In).SetLen(0xc421610000, 0xc420212000, 0x1)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/call.go:369 +0x5d
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/vnet.(*RefIn).SetLen(0xc421610000, 0xc420212000, 0x1)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/vnet/buf.go:200 +0x43
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/vnet.(*RefIn).SetPoolAndLen(0xc421610000, 0xc420212000, 0xc420574610, 0x1)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/vnet/buf.go:212 +0x58
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/fe1/internal/packet.(*rxNode).NodeInput(0xc420574410, 0xc420fdfce0)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/packet/rx.go:533 +0x2af
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/vnet.(*InputNode).LoopInput(0xc420574410, 0xc420212000, 0x1597800, 0xc420fdfce0)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/vnet/node.go:55 +0x5e
Sep  6 15:57:57 invader5 goes.vnetd[1217]: github.com/platinasystems/go/elib/loop.(*Loop).dataPoll(0xc420212000, 0x7f423b2cb4f8, 0xc420574410)
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/poller.go:350 +0x14e
Sep  6 15:57:57 invader5 goes.vnetd[1217]: created by github.com/platinasystems/go/elib/loop.(*Loop).startDataPoller
Sep  6 15:57:57 invader5 goes.vnetd[1217]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/loop.go:103 +0xe5

7. tuntap mode
Sep  6 15:50:23 invader5 goes.vnetd[561]: panic: runtime error: invalid memory address or nil pointer dereference [recovered]
Sep  6 15:50:23 invader5 goes.vnetd[561]:         panic: runtime error: invalid memory address or nil pointer dereference
Sep  6 15:50:23 invader5 goes.vnetd[561]: [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x935773]
Sep  6 15:50:23 invader5 goes.vnetd[561]:
Sep  6 15:50:23 invader5 goes.vnetd[561]: goroutine 11 [running]:
Sep  6 15:50:23 invader5 goes.vnetd[561]: github.com/platinasystems/go/elib/loop.(*Loop).dataPoll.func1()
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/poller.go:336 +0x52
Sep  6 15:50:23 invader5 goes.vnetd[561]: panic(0xbc5620, 0x143e190)
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Sep  6 15:50:23 invader5 goes.vnetd[561]: github.com/platinasystems/go/vnet.(*enqueue).sync(0xc421607440)
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/vnet/node.go:103 +0x83
Sep  6 15:50:23 invader5 goes.vnetd[561]: github.com/platinasystems/go/vnet.(*InOutNode).LoopInputOutput(0xc420104e08, 0xc420224000, 0x15977c0, 0xc4214c8020, 0x1597800, 0xc4205338c0)
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/vnet/node.go:176 +0x10d
Sep  6 15:50:23 invader5 goes.vnetd[561]: github.com/platinasystems/go/elib/loop.(*Out).call(0xc420533740, 0xc420224000, 0xc4213ba000, 0xc420533740)
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/call.go:481 +0x220
Sep  6 15:50:23 invader5 goes.vnetd[561]: github.com/platinasystems/go/elib/loop.(*Loop).dataPoll(0xc420224000, 0x7f5c0ffa9a30, 0xc42022b3c8)
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/poller.go:351 +0x17a
Sep  6 15:50:23 invader5 goes.vnetd[561]: created by github.com/platinasystems/go/elib/loop.(*Loop).startDataPoller
Sep  6 15:50:23 invader5 goes.vnetd[561]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/loop.go:103 +0xe5

8. tuntap mode: Multipath fib problem
Sep  7 15:20:20 invader5 goes.vnetd[14181]: panic: runtime error: invalid memory address or nil pointer dereference [recovered]
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         panic: runtime error: invalid memory address or nil pointer dereference
Sep  7 15:20:20 invader5 goes.vnetd[14181]: [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x95e18e]
Sep  7 15:20:20 invader5 goes.vnetd[14181]:
Sep  7 15:20:20 invader5 goes.vnetd[14181]: goroutine 37 [running]:
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/elib/loop.(*Loop).eventHandler.func1(0xc420226000, 0xc420226670)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/elib/loop/event.go:173 +0x159
Sep  7 15:20:20 invader5 goes.vnetd[14181]: panic(0xbc6f80, 0x14447b0)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /usr/local/go/src/runtime/panic.go:489 +0x2cf
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/ip.(*Main).createMpAdj(0xc420279068, 0xc4202f8600, 0x2, 0x8, 0x7fe51065e328, 0xc42063c960, 0x0)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip/adjacency.go:503 +0x39e
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/ip.(*Main).addDelHelper(0xc420279068, 0xc4202f8600, 0x2, 0x8, 0x0, 0x7fe51065e328, 0xc42063c960, 0xc4205654d0)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip/adjacency.go:625 +0xc9
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/ip.(*Main).AddDelNextHop(0xc420279068, 0xc0000000d, 0xc400000001, 0x7fe51065e328, 0xc42063c960, 0x20026fa801, 0xc42115e301)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip/adjacency.go:613 +0x24c
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelRouteNextHop(0xc420169400, 0xc420279000, 0xc42063c950, 0xc4026fa8c0, 0x15a5860, 0xc42063c960, 0xc2a001, 0xc420f9a001, 0xc42063c960)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:704 +0x34b
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/ip4.(*Main).AddDelRouteNextHop(0xc420279000, 0xc42063c950, 0xc42063c960, 0x3b026fa801, 0x1, 0xc420f9a000)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/ip4/fib.go:670 +0x82
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).ip4RouteMsg(0xc4204ac440, 0xc4204464e0, 0xc420446400, 0x1, 0xc420102000)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/unix/netlink.go:648 +0x44e
Sep  7 15:20:20 invader5 goes.vnetd[14181]: github.com/platinasystems/go/vnet/unix.(*netlinkEvent).EventAction(0xc4204ac440)
Sep  7 15:20:20 invader5 goes.vnetd[14181]:         /home/dlobete/src/github.com/platinasystems/go/vnet/unix/netlink.go:469 +0x5dd

* tuntap mode: tx ring full running iperf
Sep  7 16:44:13 invader11 goes.vnetd[1123]: panic: ixge3-0-0: tx ring full
Sep  7 16:44:13 invader11 goes.vnetd[1123]:
Sep  7 16:44:13 invader11 goes.vnetd[1123]: goroutine 28 [running]:
Sep  7 16:44:13 invader11 goes.vnetd[1123]: github.com/platinasystems/go/vnet/devices/ethernet/ixge.(*tx_dma_queue).output(0xc4206fcc00, 0xc4202282c0)
Sep  7 16:44:13 invader11 goes.vnetd[1123]:         /home/stig/go/src/github.com/platinasystems/go/vnet/devices/ethernet/ixge/tx.go:246 +0x3c5
Sep  7 16:44:13 invader11 goes.vnetd[1123]: github.com/platinasystems/go/vnet/devices/ethernet/ixge.(*dev).InterfaceOutput(0xc420280580, 0xc4202282c0)
Sep  7 16:44:13 invader11 goes.vnetd[1123]:         /home/stig/go/src/github.com/platinasystems/go/vnet/devices/ethernet/ixge/tx.go:225 +0x48
Sep  7 16:44:13 invader11 goes.vnetd[1123]: github.com/platinasystems/go/vnet.(*interfaceNode).ifOutputThread(0xc420280608)
Sep  7 16:44:13 invader11 goes.vnetd[1123]:         /home/stig/go/src/github.com/platinasystems/go/vnet/interface_node.go:102 +0xa5
Sep  7 16:44:13 invader11 goes.vnetd[1123]: created by github.com/platinasystems/go/vnet.(*HwIf).txNodeUpDown
Sep  7 16:44:13 invader11 goes.vnetd[1123]:         /home/stig/go/src/github.com/platinasystems/go/vnet/interface_node.go:125 +0xdf
Sep  7 16:44:15 invader11 goes.vnetd[1123]: exit status 2


9. sriov mode - on initial start

Aug 31 14:10:37 invader5 goes.vnetd[735]: panic: unknown interface: eth-1-0
Aug 31 14:10:37 invader5 goes.vnetd[735]:
Aug 31 14:10:37 invader5 goes.vnetd[735]: goroutine 1 [running]:
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/go/vnet/unix.(*net_namespace_main).RegisterHwInterface(0xc420302088, 0x159a960, 0xc420da5fa8)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/go/vnet/unix/net_namespace.go:719 +0x337
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1/internal/fe1a.(*Port).registerEthernet(0xc420da5fa8, 0xc420220000, 0xc420011198, 0x7, 0x101)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/port_init.go:194 +0x1d7
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).addPort(0xc420900000, 0xc420da5fa8, 0xc420d9dba0)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/port_init.go:229 +0x279
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1/internal/fe1a.(*fe1a).PortInit(0xc420900000, 0xc420220000)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/internal/fe1a/port_init.go:443 +0x9b2
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1.(*SwitchConfig).Configure(0xc4207acad0, 0xc420220000, 0x7f4dd89eaf80, 0xc420900000)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/fe1.go:50 +0x284
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1.(*platform).boardPortInit(0xc420072300, 0x7f4dd89eaf80, 0xc420900000, 0x7f4dd89eaf80, 0xc420900000)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/platform.go:110 +0x30d
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/fe1.(*platform).Init(0xc420072300, 0x6, 0xa)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/fe1/platform.go:52 +0x2d3
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/go/vnet.(*packageMain).InitPackages(0xc420220a70, 0xc42016e620, 0x0)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/go/vnet/package.go:190 +0x308
Aug 31 14:10:37 invader5 goes.vnetd[735]: github.com/platinasystems/go/vnet.(*Vnet).configure(0xc420220000, 0xc42016e620, 0xd30358, 0xc420123b00)
Aug 31 14:10:37 invader5 goes.vnetd[735]:         /home/dlobete/src/github.com/platinasystems/go/vnet/vnet.go:43 +0x5e

