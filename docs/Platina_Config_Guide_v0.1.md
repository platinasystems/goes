**Platina PSW-3001-32C Configuration Guide\
Last Update 10/10/2018**

**Quick Start Guide**

Please refer to the quick start guide included with the unit for
hardware package/unpackage instructions, MAC addresses, port labeling,
field replaceable unit labeling, and other general instructions.

**General Architecture**

The PSW-3001-32C is a 32x100GE switch with hardware forwarding up to
3.2Tbps via the Broadcom Tomahawk ASIC. Built into the switch is a
4-core Intel Broadwell DE CPU with 16GB RAM and 128GB SSD. There is also
a separate BMC ARM processor with 2GB of DRAM and 2GB of uSD. The high
level block diagram is as follows:

![](./ConfigGuide_image1.png)

**BMC Processor**

BMC processor controls the PSU, board voltage rails, fan speed, and
other GPIO on the board. For this release, there is nothing to configure
on the BMC. Everything, including temperature and fan speed, is
automated.

**Intel Processor**

After power up, the RJ45 console serial should default to Intel CPU
console. The prompt is the Linux prompt. Everything needed for using,
configuring, and monitoring the unit can be done from here.

**Linux**

Linux is the operating system for PSW-3001-32C. The unit is pre-loaded
with stock Debian Jessie and kernel version 4.13.0. The admin account is
“root”; password “platina”. All standard Linux command and usage applies
as is, including configuration of Ethernet interfaces, installation of
new apps, etc.

By default eth0 is the RJ45 Management Ethernet port on the front panel.

**Platina GOES Service**

Platina software, GOES, is a user space application that runs in Linux.
A copy of the GOES binary is at\
/root/goes-platina-mk1

GOES is pre-installed on each switch. To manually install GOES (i.e. SW
upgrade/downgrade, re-install) enter the following command:\
sudo /root/goes-platina-mk1 install

The ‘install’ command will install GOES at /usr/bin/goes, by default.

Once installed, GOES will startup automatically on reboot going forward.

You can verify GOES is running properly by entering

    goes status

and look for the following:

*GOES status*

*======================*

*Mode - XETH*

*PCI - OK*

*Check daemons - OK*

*Check Redis - OK*

*Check vnet - OK*

To uninstall goes, enter:

    sudo goes uninstall

To stop goes without doing a full uninstall enter:

    sudo goes stop

To start up goes again, enter:

    sudo goes start

Each stop/start will reset all ASIC configuration/memory and reinitialize ASIC.

GOES is an open source project developed by Platina. To see the source
code, visit

[*https://github.com/platinasystems/go*](https://github.com/platinasystems/go)

There is an addition private repository needed to compile the source. To
inquire if you meet the license requirement to access the private
repository, email
[*support@platinasystems.com*](mailto:support@platinasystems.com).

**Admin Privilege**

Running goes start/stop/install/uninstall require superuser privilege in
Linux. The Quick Start guide includes default password for root access
preconfigured on the unit.

For other ‘goes…’ commands, it is sufficient to be just part of the
‘adm’ group to be able to execute the command. A new user account should
be added to the ‘adm’ group if goes command access is required,
otherwise goes commands will need to be invoked with sudo.

**Redis Database/Interface**

GOES includes a Redis server daemon. Any configuration and stats not
normally associated with Linux can be found in this Redis database. The
GOES Redis server is a standard Redis server listening on port 6379 of
the loopback interface and eth0. The database can be accessed remotely
anywhere via standard Redis-client and Redis-cli using the IP address of
eth0.

To invoke Redis commands directly from the Linux CLI, precede the Redis
command with “goes”, otherwise standard Redis command format should be
used from a Redis client. In this guide, the command examples will
assume the user is invoking from the switch’s Linux CLI.

To see a list of all parameters in this Redis database, enter:

From Linux CLI:

    goes hgetall platina-mk1

From a Redis client:

    hgetall platina-mk1

“platina-mk1” is the hash for which all of the unit’s configurations and
stats are stored under. The field/value associated with platina-mk1
appear in the output as field:value.

You can view just a particular set of fields using hget. For example to
get all configuration/stats associated with front panel port xeth1,
enter:

    goes hget platina-mk1 xeth1

Redis database is updated every 5 seconds by default for stats counters.
The update interval for the switch ASIC stats can be changed. For
example, to change the interval to 1 second enter:

    goes hset platina-mk1 vnet.pollInterval 1

Included in the GOES Redis interface is a convenience command hdelta.
hdelta returns only the non-zero differences from the last Redis update
and includes rate calculation if applicable. hdelta runs continuously
until stopped by ctrl-c. hdelta is not a standard Redis command and must
be invoked with the following command via CLI.

    goes hdelta

Since the commands above are invoked at the standard Linux command line
prompt, ‘grep’ or other Linux tools can be used to further filter the
output.

**QSFP28 Ports**

Once GOES is installed and running, all front panel ports will show up
as normal eth interfaces in Linux. By default they are xeth1, xeth2, …
,xeth32. The format is
`xeth&lt;port\_number&gt;-&lt;sub-port\_number&gt;` where `port\_number`
corresponds to the 32 front panel ports and `sub-port\_number` corresponds
to optionally configured breakout ports within each port.

All 32 QSFP28 eth interfaces can be configured via Linux using standard
Linux methods, e.g. ip link, ip addr, /etc/network/interfaces, etc.

Interface configurations not available in Linux, such as interface speed
and media type, can be done via the ethtool.

**Set Media Type and Speed**

To set the media type and speed on xeth1 to copper (i.e. DAC cable) and
100g fixed speed for example, enter

    goes hset platina-mk1 vnet.xeth1.media copper
    sudo goes stop && sudo ip link set xeth1 up && sudo ethtool -s xeth1 speed 100000 autoneg off && sudo ifconfig xeth1 10.0.1.47/24 && sleep 3 && sudo goes start

To set the speed to autoneg enter

    sudo goes stop && sudo ip link set xeth1 up && sudo ethtool -s xeth1 autoneg on && sudo ifconfig xeth1 10.0.1.47/24 && sleep 3 && sudo goes start

***Note: In this version, media copper must be set before speed in order
for link training and receive equalization to work optimally.***

***Note: Speed options supported are auto, 100g, 50g, 40g, 25g, 10g.***

The settings do not persist across goes stop/start or reboot. To
configure permanently, add the configuration in /etc/network/interfaces
as described below.

**Persistent Configuration**

Each time GOES starts up, it will read the network configuration file
/etc/network/interfaces.

Following is the example network config file:

    # This file describes the network interfaces available on your system*
    # and how to activate them. For more information, see interfaces(5).

    source /etc/network/interfaces.d/*

    # The loopback network interface

    auto lo
    iface lo inet loopback

    # The primary network interface*

    allow-hotplug eth0
    #iface eth0 inet dhcp

    auto eth0
    iface eth0 inet static
    address 172.17.2.47
    netmask 255.255.254.0
    gateway 172.17.2.1
    dns-nameservers 8.8.8.8 8.8.4.4

    auto xeth1-1
    iface xeth1-1 inet static
    address 10.1.1.47
    netmask 255.255.255.0
    pre-up ip link set $IFACE up
    pre-up ethtool -s $IFACE speed 10000 autoneg off
    pre-up ethtool --set-priv-flags $IFACE copper on
    pre-up ethtool --set-priv-flags $IFACE fec74 on
    pre-up ethtool --set-priv-flags \$IFACE fec91 off
    post-down ip link set \$IFACE down
    allow-vnet xeth1-1

    auto xeth1-2
    iface xeth1-2 inet static
    address 10.1.2.47
    netmask 255.255.255.0
    pre-up ip link set $IFACE up
    pre-up ethtool -s $IFACE speed 10000 autoneg off
    pre-up ethtool --set-priv-flags \$IFACE copper on
    pre-up ethtool --set-priv-flags \$IFACE fec74 on
    pre-up ethtool --set-priv-flags \$IFACE fec91 off
    post-down ip link set \$IFACE down
    allow-vnet xeth1-2

In the example above, ethtool cmds are executed to set speed,media,fec
for each interface.

**BGP**

The trial unit does not come pre-installed with BGP. Any BGP protocol
stack can be installed directly onto Linux as if installing on a server.
The GOES daemon will handle any translation between the Linux kernel and
ASIC. All FIB/RIB can be obtained directly from Linux or the BGP stack.

To get the FIB from the ASIC for verification, enter:

    goes vnet show ip fib

**List of Known Issues**

For list of known issues, please visit

[*https://github.com/platinasystems/go/issues*](https://github.com/platinasystems/go/issues)

**Support**

Any question or to report new issues, please email

[*support@platinasystems.com*](mailto:support@platinasystems.com)

**Appendix 1: Redis Fields Guide**

The examples in the appendix show standard Redis commands. When directly
on the switch’s Linux CLI, add “goes” in front of the Redis command.

From Redis-client:

    hgetall platina-mk1

From switch Linux prompt:

    goes hgetall platina-mk1

To get multiple field names that contain a specific string, use hget
platina-mk1 &lt;string&gt;, for example:

    goes hget platina-mk1 xeth1

When using redis-cli, it is recommended to connect using the --raw
parameter so that newline characters between the multiple fields are
parsed properly. For example:

    redis-cli --raw -h <ipv4_address> hget platina-mk1 xeth1

**Fields that can be set in Platina’s Redis:**

Most of the fields in the Platina Redis are read-only, but some can be
set. If a set is successful, Redis will return an integer 1. Otherwise
Redis will return an error message.

**Media Type**\
Each port can be configured as copper or fiber mode. Copper mode is for
QSFP28 media such as direct attach cable (DAC) where the switch ASIC
serdes is driving the line directly. Fiber mode is for all other QSFP28
optical pluggable modules where ASIC serdes interacts typically with a
CDR at the local QSFP28 module. 100GE autoneg function and link training
is only available in copper mode.

Example:

    hset platina-mk1 vnet.xeth1.media copper

    hget platina-mk1 vnet.xeth1.media
    copper

**Speed**\
Each QSFP port supports 1GE, 10GE, 20GE, 25GE, 40GE, 50GE, and 100GE
speeds, either in single port mode or breakout port mode. In this
version, only 10GE, 25GE, 40GE, 50GE and 100GE modes are supported. If a
speed is specified, the port will be fixed at that speed. If set to
“auto” (applicable only if port is in copper media mode), the port will
negotiate with neighbor switch to establish the speed.

Example:


    sudo goes stop && sudo ip link set xeth1 up && sudo ethtool -s xeth1 autoneg on && sudo ifconfig xeth1 10.0.1.47/24 && sleep 3 && sudo goes start

    hget platina-mk1 xeth1.speed
    auto

    sudo goes stop && sudo ip link set xeth1 up && sudo ethtool -s xeth1 speed 100000 autoneg off && sudo ifconfig xeth1 10.0.1.47/24 && sleep 3 && sudo goes start

    hget platina-mk1 xeth1.speed
    100g

**Stats Counter Update Interval**

Stats counters such as transmit/receive packet counters, packet drops,
etc. are maintained real time in hardware, but updated to Redis data at
a fixed interval. This interval defaults to 5 seconds. This interval can
be configured, the minimum being 1 second. Other asynchronous event,
such as line up/down, are updated to Redis real time as the event occur.

To change the Redis counter update interval.

Example:

    hset platina-mk1 vnet.pollInterval 1

    hget platina-mk1 vnet.pollInterval 1
    vnet.pollInterval: 1

**Read-only Redis Fields**

These fields are read only. Attempts to set them will return error
messages.

**Packages**\
Packages shows the software version numbers of GOES. These version
numbers match the github commit/versions in
[*https://github.com/platinasystems*](https://github.com/platinasystems)
from which the GOES binary is built.

Example

    hget platina-mk1 packages|grep "version: "
    version: 10a3277079f775e342e48b739dd64921b22e1f6f
    version: 6dbc2b8ffebed236ce5cc8e821815cbb25cc3525
    version: 60f39141fbbf78ddb2260dba74c68f2789374f18

**Physical link state (as reported by switch ASIC)**\
vnet.\[interface\].link indicates the physical link state as reported by
the switch ASIC. The value is “true” if link is up, “false” if link is
down.

Example:

    goes hget platina-mk1 vnet.xeth3.link
    false

**Linux Packet/Byte counters**\
Deprecated from redis

The Linux interface counter (e.g. via ip link or ifconfig) counts
packets that have been sent/receive from/to the ASIC to/from the CPU
only. Packets that are locally received and forwarded in ASIC are not
included in the Linux interface counter. To get the ASIC level interface
counters, use vnet.\[interface\].port redis commands instead.

    ip -s link show xeth12

    16: xeth12@eth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
    link/ether 50:18:4c:00:11:87 brd ff:ff:ff:ff:ff:ff
    RX: bytes packets errors dropped overrun mcast
    1120240923480 1054840793 0 0 0 10
    TX: bytes packets errors dropped carrier collsns
    1966 15 0 0 0 0


**Port Counters**\
vnet.\[interface\].port… are counters that reflect the interface
counters as reported by the switch ASIC. These counters count the number
of packets/bytes that are coming in or going out of the switch ASIC’s,
as well as counters for various types of recognized packet types, for
the corresponding front panel ports.

Example: (assume launching from switches local prompt; otherwise hget
each field 1 at a time)

    goes hget platina-mk1 vnet.xeth9.port

*vnet.xeth9.port-rx-1024-to-1518-byte-packets: 1445258343*

*vnet.xeth9.port-rx-128-to-255-byte-packets: 0*

*vnet.xeth9.port-rx-1519-to-1522-byte-vlan-packets: 0*

*vnet.xeth9.port-rx-1519-to-2047-byte-packets: 0*

*vnet.xeth9.port-rx-1tag-vlan-packets: 0*

*vnet.xeth9.port-rx-2048-to-4096-byte-packets: 0*

*vnet.xeth9.port-rx-256-to-511-byte-packets: 0*

*vnet.xeth9.port-rx-2tag-vlan-packets: 0*

*vnet.xeth9.port-rx-4096-to-9216-byte-packets: 0*

*vnet.xeth9.port-rx-512-to-1023-byte-packets: 0*

*vnet.xeth9.port-rx-64-byte-packets: 0*

*vnet.xeth9.port-rx-65-to-127-byte-packets: 0*

*vnet.xeth9.port-rx-802-3-length-error-packets: 0*

*vnet.xeth9.port-rx-9217-to-16383-byte-packets: 0*

*vnet.xeth9.port-rx-alignment-error-packets: 0*

*vnet.xeth9.port-rx-broadcast-packets: 0*

*vnet.xeth9.port-rx-code-error-packets: 0*

*vnet.xeth9.port-rx-control-packets: 0*

*vnet.xeth9.port-rx-crc-error-packets: 0*

*vnet.xeth9.port-rx-eee-lpi-duration: 0*

*vnet.xeth9.port-rx-eee-lpi-events: 0*

*vnet.xeth9.port-rx-false-carrier-events: 0*

*vnet.xeth9.port-rx-flow-control-packets: 0*

*vnet.xeth9.port-rx-fragment-packets: 0*

*vnet.xeth9.port-rx-good-packets: 1445258427*

*vnet.xeth9.port-rx-jabber-packets: 0*

*vnet.xeth9.port-rx-mac-sec-crc-matched-packets: 0*

*vnet.xeth9.port-rx-mtu-check-error-packets: 0*

*vnet.xeth9.port-rx-pfc-packets: 0*

*vnet.xeth9.port-rx-pfc-priority-0: 0*

*vnet.xeth9.port-rx-pfc-priority-1: 0*

*vnet.xeth9.port-rx-pfc-priority-2: 0*

*vnet.xeth9.port-rx-pfc-priority-3: 0*

*vnet.xeth9.port-rx-pfc-priority-4: 0*

*vnet.xeth9.port-rx-pfc-priority-5: 0*

*vnet.xeth9.port-rx-pfc-priority-6: 0*

*vnet.xeth9.port-rx-pfc-priority-7: 0*

*vnet.xeth9.port-rx-promiscuous-packets: 1445258414*

*vnet.xeth9.port-rx-runt-bytes: 0*

*vnet.xeth9.port-rx-src-address-not-unicast-packets: 0*

*vnet.xeth9.port-rx-truncated-packets: 0*

*vnet.xeth9.port-rx-unicast-packets: 1445258362*

*vnet.xeth9.port-rx-unsupported-dst-address-control-packets: 0*

*vnet.xeth9.port-rx-unsupported-opcode-control-packets: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-0: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-1: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-2: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-3: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-4: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-5: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-6: 0*

*vnet.xeth9.port-rx-xon-to-xoff-priority-7: 0*

*vnet.xeth9.port-tx-1024-to-1518-byte-packets: 0*

*vnet.xeth9.port-tx-128-to-255-byte-packets: 2*

*vnet.xeth9.port-tx-1519-to-1522-byte-vlan-packets: 0*

*vnet.xeth9.port-tx-1519-to-2047-byte-packets: 0*

*vnet.xeth9.port-tx-1tag-vlan-packets: 0*

*vnet.xeth9.port-tx-2048-to-4096-byte-packets: 0*

*vnet.xeth9.port-tx-256-to-511-byte-packets: 0*

*vnet.xeth9.port-tx-2tag-vlan-packets: 0*

*vnet.xeth9.port-tx-4096-to-9216-byte-packets: 0*

*vnet.xeth9.port-tx-512-to-1023-byte-packets: 0*

*vnet.xeth9.port-tx-64-byte-packets: 0*

*vnet.xeth9.port-tx-65-to-127-byte-packets: 0*

*vnet.xeth9.port-tx-9217-to-16383-byte-packets: 0*

*vnet.xeth9.port-tx-broadcast-packets: 0*

*vnet.xeth9.port-tx-control-packets: 0*

*vnet.xeth9.port-tx-eee-lpi-duration: 0*

*vnet.xeth9.port-tx-eee-lpi-events: 0*

*vnet.xeth9.port-tx-excessive-collision-packets: 0*

*vnet.xeth9.port-tx-fcs-errors: 0*

*vnet.xeth9.port-tx-flow-control-packets: 0*

*vnet.xeth9.port-tx-fragments: 0*

*vnet.xeth9.port-tx-good-packets: 2*

*vnet.xeth9.port-tx-jabber-packets: 0*

*vnet.xeth9.port-tx-late-collision-packets: 0*

*vnet.xeth9.port-tx-multicast-packets: 2*

*vnet.xeth9.port-tx-multiple-collision-packets: 0*

*vnet.xeth9.port-tx-multiple-deferral-packets: 0*

*vnet.xeth9.port-tx-oversize: 0*

*vnet.xeth9.port-tx-pfc-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-0-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-1-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-2-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-3-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-4-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-5-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-6-packets: 0*

*vnet.xeth9.port-tx-pfc-priority-7-packets: 0*

*vnet.xeth9.port-tx-single-collision-packets: 0*

*vnet.xeth9.port-tx-single-deferral-packets: 0*

*vnet.xeth9.port-tx-system-error-packets: 0*

*vnet.xeth9.port-tx-unicast-packets: 0*

**Interface MMU Counters**\
vnet.\[interface\].mmu… are counters that reflect drops at the switch
ASICs memory management unit (MMU). This is applicable to the Broadcom
Tomahawk switch ASIC that employs the MMU to manage/switch traffic
between 4 packet processing pipelines. The \[interface\] is the egress
port, and the queues are queues associated with that port. For example,
if “vnet.xeth9.mmu-multicast-tx-cos1-drop-packets” is non-zero, that
means packets destined for xeth9 is dropping at the MMU’s multicast,
priority 1 queue.

Example:

    goes hget platina-mk1 vnet.xeth9.mmu

*vnet.xeth9.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.xeth9.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.xeth9.mmu-multicast-tx-cos1-drop-bytes: 0*

*vnet.xeth9.mmu-multicast-tx-cos1-drop-packets: 0*

*.*

*.*

*.*

*vnet.xeth9.mmu-multicast-tx-cos7-drop-bytes: 0*

*vnet.xeth9.mmu-multicast-tx-cos7-drop-packets: 0*

*vnet.xeth9.mmu-multicast-tx-qm-drop-bytes: 0*

*vnet.xeth9.mmu-multicast-tx-qm-drop-packets: 0*

*vnet.xeth9.mmu-multicast-tx-sc-drop-bytes: 0*

*vnet.xeth9.mmu-multicast-tx-sc-drop-packets: 0*

*vnet.xeth9.mmu-rx-threshold-drop-bytes: 0*

*vnet.xeth9.mmu-rx-threshold-drop-packets: 0*

*vnet.xeth9.mmu-tx-cpu-cos-0-drop-bytes: 0*

*vnet.xeth9.mmu-tx-cpu-cos-0-drop-packets: 0*

*vnet.xeth9.mmu-tx-cpu-cos-1-drop-bytes: 0*

*vnet.xeth9.mmu-tx-cpu-cos-1-drop-packets: 0*

*.*

*.*

*.*

*vnet.xeth9.mmu-tx-cpu-cos-47-drop-bytes: 0*

*vnet.xeth9.mmu-tx-cpu-cos-47-drop-packets: 0*

*vnet.xeth9.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.xeth9.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.xeth9.mmu-unicast-tx-cos1-drop-bytes: 0*

*.*

*.*

*.*

*vnet.xeth9.mmu-unicast-tx-cos7-drop-bytes: 0*

*vnet.xeth9.mmu-unicast-tx-cos7-drop-packets: 0*

*vnet.xeth9.mmu-unicast-tx-qm-drop-bytes: 0*

*vnet.xeth9.mmu-unicast-tx-qm-drop-packets: 0*

*vnet.xeth9.mmu-unicast-tx-sc-drop-bytes: 0*

*vnet.xeth9.mmu-unicast-tx-sc-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos1-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos2-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos3-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos4-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos5-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos6-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-cos7-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-qm-drop-packets: 0*

*vnet.xeth9.mmu-wred-queue-sc-drop-packets: 0*

**Interface Pipe Counters**\
vnet.\[interface\].rx-pipe… and vnet.\[interface\].tx-pipe… are counters
that reflect counters at the switch ASICs packet processing pipelines.
The rx-pipe… reflect counters for the ingress pipeline, and tx-pipe…
reflect counters for the egress pipeline. The \[interface\] is the
egress port, and the queues are queues associated with that port. For
example “vnet.xeth9.tx-pipe-multicast-queue-cos0-packets” indicates
multicast packet counters on queue 0

Example:

    goes hget platina-mk1 vnet.xeth9.rx-pipe

*vnet.xeth9.rx-pipe-debug-6: 0*

*vnet.xeth9.rx-pipe-debug-7: 0*

*vnet.xeth9.rx-pipe-debug-8: 0*

*vnet.xeth9.rx-pipe-dst-discard-drops: 0*

*vnet.xeth9.rx-pipe-ecn-counter: 0*

*vnet.xeth9.rx-pipe-hi-gig-broadcast-packets: 0*

*vnet.xeth9.rx-pipe-hi-gig-control-packets: 0*

*vnet.xeth9.rx-pipe-hi-gig-l2-multicast-packets: 0*

*vnet.xeth9.rx-pipe-hi-gig-l3-multicast-packets: 0*

*vnet.xeth9.rx-pipe-hi-gig-unknown-opcode-packets: 0*

*vnet.xeth9.rx-pipe-ibp-discard-cbp-full-drops: 0*

*vnet.xeth9.rx-pipe-ip4-header-errors: 0*

*vnet.xeth9.rx-pipe-ip4-l3-drops: 0*

*vnet.xeth9.rx-pipe-ip4-l3-packets: 0*

*vnet.xeth9.rx-pipe-ip4-routed-multicast-packets: 0*

*vnet.xeth9.rx-pipe-ip6-header-errors: 0*

*vnet.xeth9.rx-pipe-ip6-l3-drops: 0*

*vnet.xeth9.rx-pipe-ip6-l3-packets: 0*

*vnet.xeth9.rx-pipe-ip6-routed-multicast-packets: 0*

*vnet.xeth9.rx-pipe-l3-interface-bytes: 0*

*vnet.xeth9.rx-pipe-l3-interface-packets: 0*

*vnet.xeth9.rx-pipe-multicast-drops: 0*

*vnet.xeth9.rx-pipe-niv-forwarding-error-drops: 0*

*vnet.xeth9.rx-pipe-niv-frame-error-drops: 0*

*vnet.xeth9.rx-pipe-port-table-bytes: 0*

*vnet.xeth9.rx-pipe-port-table-packets: 0*

*vnet.xeth9.rx-pipe-rxf-drops: 0*

*vnet.xeth9.rx-pipe-spanning-tree-state-not-forwarding-drops: 0*

*vnet.xeth9.rx-pipe-trill-non-trill-drops: 0*

*vnet.xeth9.rx-pipe-trill-packets: 0*

*vnet.xeth9.rx-pipe-trill-trill-drops: 0*

*vnet.xeth9.rx-pipe-tunnel-error-packets: 0*

*vnet.xeth9.rx-pipe-tunnel-packets: 0*

*vnet.xeth9.rx-pipe-unicast-packets: 0*

*vnet.xeth9.rx-pipe-unknown-vlan-drops: 0*

*vnet.xeth9.rx-pipe-vlan-tagged-packets: 0*

*vnet.xeth9.rx-pipe-zero-port-bitmap-drops: 0*

*goes hget platina-mk1 vnet.xeth9.tx-pipe*

*vnet.xeth9.tx-pipe-cpu-0x10-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-0x10-packets: 0*

*.*

*.*

*.*

*vnet.xeth9.tx-pipe-cpu-0x2f-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-0x2f-packets: 0*

*vnet.xeth9.tx-pipe-cpu-error-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-error-packets: 0*

*vnet.xeth9.tx-pipe-cpu-punt-1tag-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-punt-1tag-packets: 0*

*vnet.xeth9.tx-pipe-cpu-punt-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-punt-packets: 0*

*vnet.xeth9.tx-pipe-cpu-vlan-redirect-bytes: 0*

*vnet.xeth9.tx-pipe-cpu-vlan-redirect-packets: 0*

*vnet.xeth9.tx-pipe-debug-a: 0*

*vnet.xeth9.tx-pipe-debug-b: 0*

*vnet.xeth9.tx-pipe-ecn-errors: 0*

*vnet.xeth9.tx-pipe-invalid-vlan-drops: 0*

*vnet.xeth9.tx-pipe-ip-length-check-drops: 0*

*vnet.xeth9.tx-pipe-ip4-unicast-aged-and-dropped-packets: 0*

*vnet.xeth9.tx-pipe-ip4-unicast-packets: 0*

*vnet.xeth9.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.xeth9.tx-pipe-multicast-queue-cos0-packets: 0*

*.*

*.*

*.*

*vnet.xeth9.tx-pipe-multicast-queue-cos7-bytes: 0*

*vnet.xeth9.tx-pipe-multicast-queue-cos7-packets: 0*

*vnet.xeth9.tx-pipe-multicast-queue-qm-bytes: 0*

*vnet.xeth9.tx-pipe-multicast-queue-qm-packets: 0*

*vnet.xeth9.tx-pipe-multicast-queue-sc-bytes: 0*

*vnet.xeth9.tx-pipe-multicast-queue-sc-packets: 0*

*vnet.xeth9.tx-pipe-packet-aged-drops: 0*

*vnet.xeth9.tx-pipe-packets-dropped: 0*

*vnet.xeth9.tx-pipe-port-table-bytes: 300*

*vnet.xeth9.tx-pipe-port-table-packets: 2*

*vnet.xeth9.tx-pipe-purge-cell-error-drops: 0*

*vnet.xeth9.tx-pipe-spanning-tree-state-not-forwarding-drops: 0*

*vnet.xeth9.tx-pipe-trill-access-port-drops: 0*

*vnet.xeth9.tx-pipe-trill-non-trill-drops: 0*

*vnet.xeth9.tx-pipe-trill-packets: 0*

*vnet.xeth9.tx-pipe-tunnel-error-packets: 0*

*vnet.xeth9.tx-pipe-tunnel-packets: 0*

*vnet.xeth9.tx-pipe-unicast-queue-cos0-bytes: 300*

*vnet.xeth9.tx-pipe-unicast-queue-cos0-packets: 2*

*.*

*.*

*.*

*vnet.xeth9.tx-pipe-unicast-queue-cos7-bytes: 0*

*vnet.xeth9.tx-pipe-unicast-queue-cos7-packets: 0*

*vnet.xeth9.tx-pipe-unicast-queue-qm-bytes: 0*

*vnet.xeth9.tx-pipe-unicast-queue-qm-packets: 0*

*vnet.xeth9.tx-pipe-unicast-queue-sc-bytes: 0*

*vnet.xeth9.tx-pipe-unicast-queue-sc-packets: 0*

*vnet.xeth9.tx-pipe-vlan-tagged-packets: 0*

**fe1 counters**

fe1 counters captures counters not associated with any front panel
interface. Examples are loopback port and cpu ports as seen back the
switch ASIC.

Example:

    goes hget platina-mk1 fe1|grep cos0

*vnet.fe1-cpu.bst-tx-queue-cos0-count: 0*

*vnet.fe1-cpu.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-cpu.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.fe1-cpu.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-cpu.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.fe1-cpu.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.fe1-cpu.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.fe1-cpu.tx-pipe-multicast-queue-cos0-packets: 0*

*vnet.fe1-cpu.tx-pipe-unicast-queue-cos0-bytes: 0*

*vnet.fe1-cpu.tx-pipe-unicast-queue-cos0-packets: 0*

*vnet.fe1-pipe0-loopback.bst-tx-queue-cos0-count: 0*

*vnet.fe1-pipe0-loopback.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe0-loopback.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe0-loopback.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe0-loopback.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe0-loopback.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.fe1-pipe0-loopback.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe0-loopback.tx-pipe-multicast-queue-cos0-packets: 0*

*vnet.fe1-pipe0-loopback.tx-pipe-unicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe0-loopback.tx-pipe-unicast-queue-cos0-packets: 0*

*vnet.fe1-pipe1-loopback.bst-tx-queue-cos0-count: 0*

*vnet.fe1-pipe1-loopback.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe1-loopback.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe1-loopback.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe1-loopback.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe1-loopback.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.fe1-pipe1-loopback.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe1-loopback.tx-pipe-multicast-queue-cos0-packets: 0*

*vnet.fe1-pipe1-loopback.tx-pipe-unicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe1-loopback.tx-pipe-unicast-queue-cos0-packets: 0*

*vnet.fe1-pipe2-loopback.bst-tx-queue-cos0-count: 0*

*vnet.fe1-pipe2-loopback.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe2-loopback.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe2-loopback.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe2-loopback.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe2-loopback.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.fe1-pipe2-loopback.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe2-loopback.tx-pipe-multicast-queue-cos0-packets: 0*

*vnet.fe1-pipe2-loopback.tx-pipe-unicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe2-loopback.tx-pipe-unicast-queue-cos0-packets: 0*

*vnet.fe1-pipe3-loopback.bst-tx-queue-cos0-count: 0*

*vnet.fe1-pipe3-loopback.mmu-multicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe3-loopback.mmu-multicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe3-loopback.mmu-unicast-tx-cos0-drop-bytes: 0*

*vnet.fe1-pipe3-loopback.mmu-unicast-tx-cos0-drop-packets: 0*

*vnet.fe1-pipe3-loopback.mmu-wred-queue-cos0-drop-packets: 0*

*vnet.fe1-pipe3-loopback.tx-pipe-multicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe3-loopback.tx-pipe-multicast-queue-cos0-packets: 0*

*vnet.fe1-pipe3-loopback.tx-pipe-unicast-queue-cos0-bytes: 0*

*vnet.fe1-pipe3-loopback.tx-pipe-unicast-queue-cos0-packets: 0*

**Eth0 stats**

Eth0 is the RJ45 management ethernet port that goes straight from the
CPU to/from the front panel RJ45. All the eth0 stats are available in
Linux.

**EEPROM contents**

These Redis fields show the content of the EEPROM that includes
manufacturing information on the specific hardware unit such as the part
number and serial number.

Example:

    goes hget platina eeprom

*eeprom.BaseEthernetAddress: 6c:ec:5a:07:cb:54*

*eeprom.CountryCode: 86*

*eeprom.Crc: 0x4bc3c14e*

*eeprom.DeviceVersion: 0x02*

*eeprom.DiagVersion: DIAG\_TOR1\_1.1.1*

*eeprom.LabelRevision: 0x00*

*eeprom.ManufactureDate: 2016/12/02 15:45:45*

*eeprom.Manufacturer: Platina*

*eeprom.NEthernetAddress: 134*

*eeprom.Onie.Data: "TlvInfo\\x00" 0x546c76496e666f00*

*eeprom.Onie.Version: 0x01*

*eeprom.OnieVersion: 2015.11*

*eeprom.PartNumber: BT77O759.00*

*eeprom.PlatformName: Intel-DE\_Platina*

*eeprom.ProductName: TOR1*

*eeprom.SerialNumber: HAY16B7B0000B*

*eeprom.ServiceTag: Empty*

*eeprom.Vendor: Platina*

*eeprom.VendorExtension:
"\\xbceP\\x01\\x00Q\\x01\\x02R\\x01\\x01S\\x0e900-000000-002T\\vHB1N645000NT\\vHB2N645000WT\\vHB4N645000M*"

**Appendix 2: BMC Redis**

The BMC runs a separate Redis server that provides configuration and
status of the hardware platform including power supplies, fan trays, and
environmental monitoring.

Accessing the BMC redis server should be done via Redis client. For
convenience, the redis-tools package is pre-installed in the switch’s
Linux (“sudo apt-get install redis-tools”) and includes the redis-cli
command that can be used to access the BMC redis from the switch’s Linux
CLI.

The BMC Redis server listens on port 6379 of the BMC’s eth0 IPv4 address
and IPv6 link local address.

**Connecting via IPv4**

By default the BMC eth0 IPv4 address is 192.168.101.100. To connect from
the switch’s Linux CLI:

    root@platina:~# redis-cli -h 192.168.101.100
    192.168.101.100:6379> hget platina machine*
    "platina-mk1-bmc"

**Connecting via IPv6 Link Local**

The BMC eth0 IPv6 link local address can be directly translated from the
BMC eth0 MAC address. The BMC eth0 MAC address is simply the base MAC
address of the system which can be retrieved from the switch’s Linux
CLI:

    root@platina:~# goes hget platina eeprom.BaseEthernetAddress
    50:18:4c:00:13:04

Convert *50:18:4c:00:13:04* MAC address to IPv6 link local:
fe80::5218:4cff:fe00:1304 and connect to BMC redis:

    root@platina:~# redis-cli -h fe80::5218:4cff:fe00:1304%eth0
    [fe80::5218:4cff:fe00:1304%eth0]:6379> hget platina machine*
    "platina-mk1-bmc"

**Notable BMC Redis Fields**

Temperature status

    root@platina:~# redis-cli --raw -h fe80::5218:4cff:fe00:1304%eth0 hget platina temp

*bmc.temperature.units.C: 37.01*

*hwmon.front.temp.units.C: 49.000*

*hwmon.rear.temp.units.C: 54.000*

*psu1.temp1.units.C: 33.375*

*psu1.temp2.units.C: 37.812*

Fan and PSU status

    root@platina:~# redis-cli --raw -h fe80::5218:4cff:fe00:1304%eth0 hget platina status

*fan\_tray.1.status: ok.front-&gt;back*

*fan\_tray.2.status: ok.front-&gt;back*

*fan\_tray.3.status: ok.front-&gt;back*

*fan\_tray.4.status: ok.front-&gt;back*

*psu1.status: powered\_on*

*psu2.status: not\_installed*

Fan Speed

    root@platina:~# redis-cli --raw -h fe80::5218:4cff:fe00:1304%eth0 hget platina fan_tray

*fan\_tray.1.1.speed.units.rpm: 7031*

*fan\_tray.1.2.speed.units.rpm: 7031*

*fan\_tray.1.status: ok.front-&gt;back*

*fan\_tray.2.1.speed.units.rpm: 7031*

*fan\_tray.2.2.speed.units.rpm: 6490*

*fan\_tray.2.status: ok.front-&gt;back*

*fan\_tray.3.1.speed.units.rpm: 7031*

*fan\_tray.3.2.speed.units.rpm: 7031*

*fan\_tray.3.status: ok.front-&gt;back*

*fan\_tray.4.1.speed.units.rpm: 7031*

*fan\_tray.4.2.speed.units.rpm: 6490*

*fan\_tray.4.status: ok.front-&gt;back*

*fan\_tray.duty: 0x4d*

*fan\_tray.speed: auto*

**PSU Information**

    root@platina:~# redis-cli --raw -h fe80::5218:4cff:fe00:1304%eth0 hget platina psu

*psu1.admin.state: enabled*

*psu1.eeprom:*
*01000000010900f5010819c54757202020cb47572d4352505335353020ca58585858585858585858c3585846ce5053555134303030303037475720c0c0c10000000000000000002f00021822c42602390319052823b036504620672f3f0c1f94c20000001f01020d09e701b0047404ec047800e803c8af01820d274982b0047404ec0478000000b80b2020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020*

*psu1.fan\_speed.units.rpm: 4020*

*psu1.i\_out.units.A: 10.688*

*psu1.mfg\_id: Great Wall*

*psu1.mfg\_model: CRPS550*

*psu1.p\_in.units.W: 141.000*

*psu1.p\_out.units.W: 129.000*

*psu1.status: powered\_on*

*psu1.temp1.units.C: 33.375*

*psu1.temp2.units.C: 37.812*

*psu1.v\_in.units.V: 117.000*

*psu1.v\_out.units.V: 12.047*

*psu2.admin.state: enabled*

*psu2.status: not\_installed*

**Power Monitor Information**

    root@platina:~# redis-cli --raw -h fe80::5218:4cff:fe00:1304%eth0 hget platina vmon

*vmon.1v0.tha.units.V: 1.041*

*vmon.1v0.thc.units.V: 1.014*

*vmon.1v2.ethx.units.V: 1.187*

*vmon.1v25.sys.units.V: 1.24*

*vmon.1v8.sys.units.V: 1.805*

*vmon.3v3.bmc.units.V: 3.28*

*vmon.3v3.sb.units.V: 3.344*

*vmon.3v3.sys.units.V: 3.298*

*vmon.3v8.bmc.units.V: 3.826*

*vmon.5v.sb.units.V: 4.926*

*vmon.poweroff.events:*
*1970-01-01T23:30:34Z.1970-01-01T23:32:28Z.1970-01-01T23:40:17Z.1970-01-01T23:50:54Z.1970-01-01T23:52:04Z*
