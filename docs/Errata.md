This is a list of known problems in the latest code. This is a living document that will
change over time as issues are resolved.

# Broken Functionality

## Tuntap mode:

- incorporate qsfpd into vnetd so all i2c accesses are centralized and can be mutexed
- multipath adjacency handling breaks gre (Stig's 4 inv rig - 2 paths one gre the other just eth - needs to be fixed)
- inv5:/root/nsgreeth.sh doesn't ping - fib looks ok
	root@invader5:/home/dlobete/src/github.com/platinasystems/go# goes vnet show err
             Node                           Error                              Count     
	     fe1-rx                         management duplicate               58
	     fe1-single-tagged-punt         not single vlan tagged             110

- gre-tunnels: Having trouble getting a script working here. What is correct
	config for this case? See nsgrevnet.inv[5,4].sh
- vnet interface not created in every namespace

## SRIOV-mode:

- panics on every goes start (unknown interface)
- iperf load breaks punt-path
- punt performance only 1 Gbps (punting through PCI)
- Issue 78

# Known panics
1. sriov mode - on initial start

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

