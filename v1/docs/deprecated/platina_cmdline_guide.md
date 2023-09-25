# Platina GoES Command Line Guide

## Changelog

1. *3 November 2018 - Initial version converted to markdown*


## Overview

This document is the beginning of a command line reference for the Platina device. This is currently an internal reference and doesn't contain content for end users.

## 2.1 GoES Commands

### 2.1.1 goes help

#### Command

    > sudo goes help

#### Description

Lists all the available command for goes. Currently it is not showing
the actual commands like "goes hget, goes vnet" that can start with
     "goes" but instead it is showing "goes COMMAND"

#### Examples

    > sudo goes help
    
    
     usage: goes COMMAND [ ARGS ]...
    
     goes COMMAND -[-]HELPER [ ARGS ]...
    
     goes HELPER [ COMMAND ] [ ARGS ]...
    
     goes [ -d ] [ -x ] [[ -f ][ - | SCRIPT ]]
    
     HELPER := { apropos | complete | help | man | usage }

### 2.1.2 goes man

#### Command

    > goes man

#### Description

    > goes man
    
     NAME
    
     goes-platina-mk1 - goes machine for platina's mk1 TOR
    
     SYNOPSIS
    
     goes COMMAND [ ARGS ]...
    
     goes COMMAND -[-]HELPER [ ARGS ]...
    
     goes HELPER [ COMMAND ] [ ARGS ]...
    
     goes [ -d ] [ -x ] [[ -f ][ - | SCRIPT ]]
    
     HELPER := { apropos | complete | help | man | usage }
    
     OPTIONS
    
     -d debug block handling
    
     -x print command trace
    
     -f don't terminate script on error
    
     - execute standard input script
    
     SCRIPT execute named script file
    
     SEE ALSO
    
     goes apropos [COMMAND], goes man COMMAND
    
### 2.1.3 goes usage

#### Command

    > sudo goes usage

#### Description

    > goes usage
    
     usage: goes COMMAND [ ARGS ]...
    
     goes COMMAND -[-]HELPER [ ARGS ]...
    
     goes HELPER [ COMMAND ] [ ARGS ]...
    
     goes [ -d ] [ -x ] [[ -f ][ - | SCRIPT ]]
    
     HELPER := { apropos | complete | help | man | usage }
    

## 2.2 goes apropos

#### Command

    > goes apropos

#### Description

     ! run an external command
    
     /init bootstrap
    
     [ test conditions and set exit status
    
bootc bootc provides wipe and access to bootc.cfg.
    
     bootd http boot controller daemon
    
     cat
    
     DESCRIPTION
    
     Concatenate FILE(s), or standard input, to standard output.
    
     With no FILE, or when FILE is -, read standard input.
    
     EXAMPLES
    
     cat f - g
    
     Output f's contents, then standard input, then g's contents.

#### Examples

     ! run an external command
    
     /init bootstrap
    
     [ test conditions and set exit status
    
     bootc bootc provides wipe and access to bootc.cfg.
    
     bootd http boot controller daemon
    
     cat
    
     DESCRIPTION
    
     Concatenate FILE(s), or standard input, to standard output.
    
     With no FILE, or when FILE is -, read standard input.
    
     EXAMPLES
    
     cat f - g
    
     Output f's contents, then standard input, then g's contents.
    
     cat
    
     Copy standard input to standard output.
    
     cd change the current directory
    
     chmod change file mode
    
     cli command line interpreter
    
     cp copy files and directories
    
     dmesg print or control the kernel ring buffer
    
     echo print a line of text
    
     eeprom show, delete or modify eeprom fields
    
     else if COMMAND ; then COMMAND else COMMAND endelse
    
     env run a program in a modified environment
    
     exec execute a file
    
     exit exit the shell
    
     export set process configuration
    
     false Fail regardless of our ability
    
     femtocom tiny serial-terminal emulation
    
     fi end of if command block
    
     function function name { definition }
    
     goes-daemons start daemons and wait for their exit
    
     gpio manipulate GPIO pins
    
     hdel delete one or more redis hash fields
    
     hdelta print the changed fields of a redis hash
    
     hexists determine if the redis hash field exists
    
     hget get the value of a redis hash field
    
     hgetall get all the field values in a redis hash
    
     hkeys get all the fields in a redis hash
    
     hset set the string value of a redis hash field
    
     hwait wait until the redis hash field has given value
    
     if conditional command
    
     insmod insert a module into the Linux Kernel
    
     install install this goes machine
    
     io read/write the CPU's I/O ports
    
     ip show / manipulate routing, etc.
    
     kexec load a new kernel for later execution
    
     keys find all redis keys matching the given pattern
    
     kill signal a process
    
     ln make links between files
    
     log print text to /dev/kmsg
    
     ls list directory contents
    
     lsmod print status of Linux Kernel modules
    
     mkdir make directories
    
     mknod make block or character special files
    
     mount activated a filesystem
    
     ping send ICMP ECHO_REQUEST to network host
    
     ps print process state
    
     pwd print working directory
    
     qsfp qsfp monitoring daemon, publishes to redis
    
     reboot reboot system
    
     redisd a redis server
    
     reload SIGHUP this goes machine
    
     restart stop, then start this goes machine
    
     rm remove files or directories
    
     rmmod remove a module from the Linux Kernel
    
     show print stuff
    
     sleep suspend execution for an interval of time
    
     source import command script
    
     start start this goes machine
    
     status print status of goes daemons
    
     stop stop this goes machine
    
     stty print info for given or current TTY
    
     subscribe print messages published to given redis channel
    
     sync flush file system buffers
    
     tempd temperature monitoring daemon, publishes to redis
    
     then conditionally execute commands
    
     toggle toggle console port between x86 and BMC
    
     true Be successful not matter what
    
     umount deactivate filesystems
    
     uninstall uninstall this goes machine
    
     upgrade upgrade images
    
     uptimed record system uptime in redis
    
     vnet send commands to hidden cli
    
     vnetd FIXME
    
     wget a non-interactive network downloader

### 2.2.1 goes !

#### Command

    > goes !

#### Description

     Run an external command.
    
    > goes ! help
    
     usage: ! COMMAND [ARGS]... [&]
    
    >]

#### Examples

    > goes !

### 2.2.2 goes /init

#### Command

    > goes /init

#### Description

     Bootstrap command.
    
    > goes /init
    
     init [OPTIONS...] {COMMAND}
    
     Send control commands to the init daemon.
    
     --help Show this help
    
     --no-wall Don't send wall message before halt/power-off/reboot
    
     Commands:
    
     0 Power-off the machine
    
     6 Reboot the machine
    
     2, 3, 4, 5 Start runlevelX.target unit
    
     1, s, S Enter rescue mode
    
     q, Q Reload init daemon configuration
    
     u, U Reexecute init daemon
    
    >]

#### Examples

    > goes /init 6

### 2.2.3 goes [

#### Command

    > goes [

#### Description

     Test conditions and set exit status.
    
    > goes [ help
    
     [: missing ]
    
    >]

#### Examples

    > goes [

### 2.2.4 goes bootc

#### Command

    > goes bootc

#### Description

     Bootc provides wipe and access to bootc.cfg.
    
    > goes bootc help
    
     usage: bootc [register] [bootc] [dumpvars] [dashboard]
     [initcfg]
    
     [wipe]
    
     [getnumclients] [getclientdata] [getscript] [getbinary]
    
     [testscript]
    
     [test404] [dashboard9] [setsda6] [clrsda6] [setinstall]
    
     [clrinstall]
    
     [setsda1] [clrsda1] [readcfg] [setip] [setnetmask]
    
     [setgateway]
    
     [setkernel] [setinitrd] [getclientbootdata] [setpost]
    
#### Examples

    > goes bootc

### 2.2.5 goes bootd

#### Command

    > goes bootd

#### Description

     http-boot controller daemon.
    
    > goes bootd help
    
     usage: bootd
    
    >]

#### Examples

    > goes bootd

### 2.2.6 goes cat 

#### Command

    > goes cat

#### Description

     Concatenate FILE(s), or standard input, to standard output.
    
     With no FILE, or when FILE is -, read standard input..
    
    > goes cat help
    
     usage: cat [FILE]...
    
#### Examples

    > goes ls
    
     .bash_history .bashrc .ssh
    
     .bash_logout .profile
    
    > goes cat .profile
    
     # ~/.profile: executed by the command interpreter for login shells.
     # This file is not read by bash(1), if ~/.bash_profile or
     ~/.bash_login
     # exists.
    # see /usr/share/doc/bash/examples/startup-files for examples.
     # the files are located in the bash-doc package.
    # the default umask is set in /etc/profile; for setting the umask
    # for ssh logins, install and configure the libpam-umask package.
     #umask 022
     # if running bash
    
     if [ -n "$BASH_VERSION" ]; then
    
     # include .bashrc if it exists
    
     if [ -f "$HOME/.bashrc" ]; then
    
     . "$HOME/.bashrc"
    
     fi
    
     fi
    
    # set PATH so it includes user's private bin if it exists
    
     if [ -d "$HOME/bin" ] ; then
    
     PATH="$HOME/bin:$PATH"
    
     fi
    
    >

### 2.2.7 goes cd

#### Command

    > goes cd

#### Description

Changes the current directory. (to the directory mentioned)
    
    > goes cd help
    
     usage: cd [- | DIRECTORY]
    
    >]

#### Examples

    > goes cd
    
     (not working)

### 2.2.8 goes chmod

#### Command

    > goes chmod

#### Description

Allows to change the file mode/permission.

There are different mode-file options available
    
 - chmod 600 file -- owner can read and write
 - chmod 700 file -- owner can read, write and execute
 - chmod 666 file -- all can read and write
 - chmod 777 file -- all can read, write and execute
    
    > goes chmod help
    
     usage: chmod MODE FILE...
    
#### Examples

    > ls --la
    
     --w-rwxr-- 1 root root 585 Jul 20 04:28 subuid
    
    > goes chmod 777 subuid
    
    > ls --la
    
     -r----x--x 1 root root 585 Jul 20 04:28 subuid

### 2.2.9 goes cli

#### Command

    > goes cli

#### Description

     Changes to Command Line Interpreter.
    
    > goes cli help
    
     usage: cli [-x] [-p PROMPT] [URL]
    
    >]

#### Examples

    > goes cli
    
     invader44>

### 2.2.10 goes cp

#### Command

    > goes cp

#### Description

Copy files and directories .There are different options which can be
used with this command, which include:
    
1. Treat 'destination' as a normal file.
2. Copy all source arguments into the directory
3. Making a copy of a file in the same directory.
    
    > goes cp help
    
     usage: cp [-v] -T SOURCE DESTINATION
    
     cp [-v] -t DIRECTORY SOURCE...
    
     cp [-v] SOURCE... DIRECTORY
    
#### Examples

    > ls
    
     testfile.txt
    
    > goes cp testfile.txt testfile55.txt
    
    > ls
    
     testfile.txt testfile55.txt

### 2.2.11 goes dmesg

#### Command

    > goes dmesg

#### Description

Print or control the kernel ring buffer.
    
There are different options for this command:

1.  goes dmesg -C,--clear :- Clears the ring buffer.
2.  goes dmesg -c,--read-clear :- Clears the ring buffer after
    printing.
3.  goes dmesg --D,--console-off :- Disable printing messages to the
    console.
4.  goes dmesg --E,--console-on :- Enable Printing messages to the
    console
5.  goes dmesg --d,--show-delta :- Display the time-stamp and
    time-delta spent between the messages.
6.  goes dmesg --f,--facility list :- Restricts output to defined
    (comma-seperated) list of facilities.
7.  goes dmesg --k,--kernel :- Print Kernel messages.
8.  goes dmesg --r,--raw :- Print the law level buffer, i.e. don't skip
    the log-level-prefixes.
9.  goes dmesg --t,--notime :- Don't print kernel time-stamps.
10. goes dmesg --u,--userspace :- Print Userspace messages.
11. goes dmesg --V,--version :- Output version Information and exit.

    > goes dmesg help
    
     usage: dmesg [OPTION]...
    
#### Examples

    > goes dmesg 
    
     [0000000.000002] Calibrating delay loop (skipped), value calculated
     using timer frequency.. 4390.10 BogoMIPS (lpj=8780208)
    
     [0000000.004000] pid_max: default: 32768 minimum: 301
    
     [0000000.004000] ACPI: Core revision 20170531
    
     [0000000.004000] ACPI: 2 ACPI AML tables successfully acquired and
     loaded
    
     [0000000.004000] Mount-cache hash table entries: 32768 (order: 6,
     262144 bytes)
    
     [0000000.004000] Mountpoint-cache hash table entries: 32768 (order:
     6, 262144 by tes)
    
     [0000000.004000] CPU: Physical Processor ID: 0
    
     [0000000.004000] CPU: Processor Core ID: 0
    
     [0000000.004000] mce: CPU supports 22 MCE banks
    
     [0000000.004000] CPU0: Thermal monitoring enabled (TM1)
    
     [0000000.004000] process: using mwait in idle threads
    
     [0000000.004000] Last level iTLB entries: 4KB 64, 2MB 8, 4MB 8
    
     [0000000.004000] Last level dTLB entries: 4KB 64, 2MB 0, 4MB 0, 1GB
     4
    
     [0000000.004000] Freeing SMP alternatives memory: 24K
    
     [0000000.004000] smpboot: Max logical packages: 2
    
     [0000000.004000] DMAR: Host address width 46
    
     [0000000.004000] DMAR: DRHD base: 0x000000fbffc000 flags: 0x1
    
     [0000000.004000] DMAR: dmar0: reg_base_addr fbffc000 ver 1:0 cap
     8d2078c106f0466 ecap f020df
    
     ....
    
     ....
    
     [0237606.347447] IPv6: ADDRCONF(NETDEV_UP): eth-1-1: link is not
     ready
    
     [0240352.324668] docker0: port 1(veth772b426) entered blocking state
    
     [0240352.329847] docker0: port 1(veth772b426) entered disabled state
    
     [0240352.335051] device veth772b426 entered promiscuous mode
    
     [0240352.339547] IPv6: ADDRCONF(NETDEV_UP): veth772b426: link is
     not ready
    
     [0240352.522042] eth0: renamed from veth15c6df9
    
     [0240352.554127] IPv6: ADDRCONF(NETDEV_CHANGE): veth772b426: link
     becomes ready
    
     [0240352.560444] docker0: port 1(veth772b426) entered blocking state
    
     [0240352.565607] docker0: port 1(veth772b426) entered forwarding
     state
    
### 2.2.12 goes echo

#### Command

    > goes echo

#### Description

Print a line of text
    
    > goes echo help
    
     usage: echo [-n] [STRING]...
    
    >]

#### Examples

    > goes echo "Platina Systems"
    
     Platina Systems
    
    > goes echo -x "Platina Systems"
    
     Platina Systems Platina Systems
    
    > goes echo -n "Platina Systems"
    
### 2.2.13 goes eeprom

#### Command

    > goes eeprom

#### Description

Show, delete or modify eeprom fields.
    
    > goes eeprom help
    
     usage: eeprom [-n] [-FIELD | FIELD=VALUE]...
    
    >]

#### Examples

    > goes eeprom 

### 2.2.14 goes else

#### Command

    > goes else

#### Description

     if COMMAND ; then COMMAND else COMMAND endelse.
    
    > goes else help
    
     usage: else COMMAND
    
    >]

#### Examples

    > goes else 

### 2.2.15 goes env

#### Command

    > goes env

#### Description

Run a program in a modified environment.
    
    > goes env help
    
     usage: env [NAME[=VALUE... COMMAND [ARGS...]]]
    
    >]

#### Examples

    > goes env

### 2.2.16 goes exec

#### Command

    > goes exec

#### Description

Execute a file.
    
    > goes exec help
    
     usage: exec COMMAND...
    
    >]

#### Examples

    > goes exec

### 2.2.17 goes exit

#### Command

    > goes exit

#### Description

Used to exit the shell.
    
    > goes exit help
    
     usage: exit [N]
    
    >]

#### Examples

    > goes exit

### 2.2.18 goes export

#### Command

    > goes export

#### Description

Sets the process configuration.
    
    > goes export help
    
     usage: export [NAME[=VALUE]]...
    
#### Examples

    > goes export

### 2.2.19 goes false

#### Command

    > goes false

#### Description

Returns failure.
    
    > goes false help
    
     usage: false
    
#### Examples

    > goes false

### 2.2.20 goes femtocom

#### Command

    > goes femtocom

#### Description

Tiny serial-terminal emulation.
    
    > goes femtocom help
    
     usage: femtocom [OPTION]... DEVICE
    
#### Examples

    > goes femtocom

### 2.2.21 goes fi

#### Command

    > goes fi help

#### Description

End of if command block.
    
    > goes fi help
    
     usage: fi
    
    >]

#### Examples

    > goes fi

### 2.2.22 goes function

#### Command

    > goes function

#### Description

Function name { definition }
    
    > goes function help
    
     usage: function name { definition }
    
    >]

#### Examples

    > goes function

### 2.2.23 goes goes-daemons

#### Command

    > goes goes-daemons

#### Description

Start daemons and wait for their exit.
    
    > goes goes-daemons help
    
     usage: goes-daemons [OPTIONS]...
    
    >]

#### Examples

    > goes goes-daemons

### 2.2.24 goes gpio

#### Command

    > goes gpio

#### Description

Manipulate GPIO pins.
    
    > goes gpio help
    
     usage: gpio PIN_NAME [VALUE]
    
#### Examples

    > goes gpio

### 2.2.25 goes hdel

#### Command

    > goes hdel

#### Description

Delete one or more redis hash fields.
    
    > goes hdel help
    
     usage: hdel KEY FIELD
    
    >]

#### Examples

    > goes hdel

### 2.2.26 goes hdelta

#### Command

    > goes hdelta

#### Description

Print the changed fields of a redis hash.
    
    > goes hdelta help
    
     usage: hdelta [CHANNEL]
    
#### Examples

    > goes hdelta

### 2.2.27 goes hexists

#### Command

    > goes hexists

#### Description

     Determine if the redis hash field exists.
    
    > goes hexists help
    
     usage: hexists KEY FIELD
    
#### Examples

    > goes hexists

### 2.2.28 goes hget

#### Command

    > goes hget

#### Description

Get the value of a redis hash field.
    
    > goes hget help
    
     usage: hget KEY FIELD
    
#### Examples

    > goes hget

### 2.2.29 goes hgetall

#### Command

    > goes hgetall

#### Description

Get all the field value of a redis hash.
    
    > goes hgetall help
    
     usage: hgetall [KEY]
    
#### Examples

    > goes hgetall

### 2.2.30 goes hkeys

#### Command

    > goes hkeys

#### Description

Get all the fields in a redis hash.
    
    > goes hkeys help
    
     usage: hkeys KEY
    
    >]

#### Examples

    > goes hkeys

### 2.2.31 goes hset

#### Command

    > goes hset

#### Description

Set the string value of a redis hash field.
    
    > goes hset help
    
     usage: hset [-q] KEY FIELD VALUE
    
    >]

#### Examples

    > goes hset

### 2.2.32 goes hwait

#### Command

    > goes hwait

#### Description

Waits until the redis hash field has given value.
    
    > goes hwait help
    
     usage: hwait KEY FIELD VALUE [TIMEOUT(seconds)]
    
    >]

#### Examples

    > goes hwait

### 2.2.33 goes if

#### Command

    > goes if

#### Description

Conditional command.
    
    > goes if help
    
usage: if COMMAND ; then COMMAND else COMMAND endif
    
#### Examples

    > goes if

### 2.2.34 goes insmod

#### Command

    > goes insmod

#### Description

Insert a module into the Linux Kernel.
    
    > goes insmod help
    
     usage: insmod [OPTION]... FILE [NAME[=VAL[,VAL]]]...
    
#### Examples

    > goes insmod

### 2.2.35 goes install

#### Command

    > goes install

#### Description

Install this goes machine.
    
    > goes install help
    
     usage: install [START, STOP and REDISD options]...
    
#### Examples

    > goes install

### 2.2.36 goes io

#### Command

    > goes io

#### Description

Read/write the CPU's I/O ports.
    
    > goes io help
    
     usage: io [[-r] | -w] IO-ADDRESS [-D DATA] [-m MODE]
    
#### Examples

    > goes io

### 2.2.37 goes ip

#### Command

    > goes ip

#### Description

Show / manipulate routing, IP information etc..
    
    > goes ip help
    
     usage: ip [ NETNS ] OBJECT [ COMMAND [ FAMILY ] [ OPTIONS ]...
     [
    
     ARG ]... ]
    
     ip [ NETNS ] -batch [ -x | -f ] [ - | FILE ]
    
     NETNS := { -a[ll] | -n[etns] NAME }
    
     OBJECT := { address | fou | link | monitor | neighbor | netns |
    
     route }
    
     FAMILY := { -f[amily] { inet | inet6 | mpls | bridge | link } |
    
     { -4 | -6 | -B | -0 } }
    
     OPTION := { -s[tat[isti]cs] | -d[etails] | -r[esolve] |
    
     -human[-readable] | -iec |
    
     -l[oops] { maximum-addr-flush-attempts } | -br[ief] |
    
     -o[neline] | -t[imestamp] | -ts[hort] |
    
     -rc[vbuf] [size] | -c[olor] }
    
    >]

#### Examples

    > goes ip

### 2.2.38 goes kexec

#### Command

    > goes kexec

#### Description

     Load a new kernel for later execution.
    
    > goes kexec help
    
     usage: kexec [OPTIONS]...
    
    >]

#### Examples

    > goes kexec

### 2.2.39 goes keys

#### Command

    > goes keys

#### Description

Find all redis keys matching the given pattern.
    
    > goes keys help
    
     usage: keys [PATTERN]

#### Examples

    > goes keys

### 2.2.40 goes kill

#### Command

    > goes kill

#### Description

Signal a process.
    
    > goes kill help
    
     usage: kill [OPTION] [PID]...
    
    >]

#### Examples

    > goes kill

### 2.2.41 goes ln

#### Command

    > goes ln

#### Description

Make links between files.
    
    > goes ln help
    
     usage: ln [OPTION]... -t DIRECTORY TARGET...
    
     ln [OPTION]... -T TARGET LINK
    
     ln [OPTION]... TARGET LINK
    
     ln [OPTION]... TARGET... DIRECTORY
    
     ln [OPTION]... TARGET
    
#### Examples

    > goes ln

### 2.2.42 goes log

#### Command

    > goes log 

#### Description

Print text to /dev/kmsg.
    
    > goes log help
    
     usage: log [PRIORITY [FACILITY]] TEXT...
    
    >]

#### Examples

    > goes log

### 2.2.43 goes ls

#### Command

    > goes ls

#### Description

List directory contents.
    
    > goes ls help
    
     usage: ls [OPTION]... [FILE]...
    
    >]

#### Examples

    > goes ls
    
     container_testcase_single_vlan
    
     goes-platina-mk1-installer
    
     goesd-platina-mk1-modprobe.conf
    
     goesd-platina-mk1-modules.conf
    
     goesd-platina-mk1-sysctl.conf
    
     .ansible .profile coreboot-platina-mk1.rom
    
     .bash_history .ssh test1
    
     .bashrc .vimrc tools
    
     .history_quagga cb.rom volumes
    
    >

### 2.2.44 goes lsmod

#### Command

    > goes lsmod

#### Description

Print status of Linux Kernel modules.
    
    > goes lsmod help
    
     usage: lsmod
    
    >]

#### Examples

    > goes lsmod
    
     Module Size Used by
    
     platina_mk1 53248 0
    
     8021q 24576 0
    
     ipt_MASQUERADE 16384 1
    
     nf_nat_masquerade_ipv4 16384 1 ipt_MASQUERADE
    
     nf_conntrack_netlink 28672 0
    
     nfnetlink 16384 2 nf_conntrack_netlink
    
     iptable_nat 16384 1
    
     nf_conntrack_ipv4 16384 3
    
     nf_defrag_ipv4 16384 1 nf_conntrack_ipv4
    
     nf_nat_ipv4 16384 1 iptable_nat
    
     xt_addrtype 16384 2
    
     iptable_filter 16384 1
    
     xt_conntrack 16384 1
    
     nf_nat 24576 2 nf_nat_masquerade_ipv4,nf_nat_ipv4
    
     nf_conntrack 73728 7
ipt_MASQUERADE,nf_nat_masquerade_ipv4,nf_conntrack_netlink,nf_conntrack_ipv4,nf_nat_ipv4,xt_conntrack,nf_nat
    
     br_netfilter 24576 0
    
     bridge 110592 1 br_netfilter
    
     stp 16384 1 bridge
    
     llc 16384 2 bridge,stp
    
     overlay 57344 0
    
     kvm_intel 184320 0
    
     kvm 344064 1 kvm_intel
    
     iTCO_wdt 16384 1
    
     iTCO_vendor_support 16384 1 iTCO_wdt
    
     ixgbevf 57344 0
    
     autofs4 36864 0
    
     i2c_i801 24576 0
    
     ixgbe 249856 0
    
     mdio 16384 1 ixgbe
    
    >

### 2.2.45 goes mkdir

#### Command

    > goes mkdir

#### Description

Make directories.
    
    > goes mkdir help
    
     usage: mkdir [OPTION]... DIRECTORY...
    
    >]

#### Examples

    > ls
    
cb.rom goesd-platina-mk1-modprobe.conf goes-platina-mk1-installer
    
container_testcase_single_vlan goesd-platina-mk1-modules.conf tools
    
coreboot-platina-mk1.rom goesd-platina-mk1-sysctl.conf volumes
    
    > goes mkdir test1
    
    >
    
    > ls
    
     cb.rom goesd-platina-mk1-modules.conf tools
    
container_testcase_single_vlan goesd-platina-mk1-sysctl.conf
     volumes
    
coreboot-platina-mk1.rom goes-platina-mk1-installer
    
     goesd-platina-mk1-modprobe.conf test1
    
    >

### 2.2.46 goes mknod

#### Command

    > goes mknod

#### Description

Make block or character special files.
    
    > goes mknod help
    
     usage: mknod [OPTION]... NAME TYPE [MAJOR MINOR]
    
    >]

#### Examples

    > goes mknod

### 2.2.47 goes mount

#### Command

    > goes mount

#### Description

     Activate a file-system.
    
    > goes mount help
    
     usage: usage [OPTION]... DEVICE [DIRECTORY]
    
    >]

#### Examples

    > goes mount 

### 2.2.48 goes ping

#### Command

    > goes ping

#### Description

     Send ICMP ECHO_REQUEST to network host.
    
    > goes ping help
    
     usage: ping DESTINATION
    
    >]

#### Examples

    > goes ping 172.17.2.45
    
     PING 172.17.2.45 (172.17.2.45)
    
     64 bytes from 172.17.2.45 in 3.232638ms
    
    >

### 2.2.49 goes ps

#### Command

    > goes ps

#### Description

     Print process state.
    
    > goes ps help
    
     usage: ps [OPTION]...
    
    >]

#### Examples

    > goes ps
    
     PID TTY TIME CMD
    
     12082 pts/45 60ms goes
    
     13172 pts/45 0s sh
    
     13177 pts/45 0s bash
    
    >

### 2.2.50 goes pwd

#### Command

    > goes pwd

#### Description

     Print working directory.
    
    > goes pwd help
    
     usage: pwd [-L]
    
    >]

#### Examples

    > pwd
    
     /root
    
    >

### 2.2.51 goes qsfp

#### Command

    > goes qsfp

#### Description

     Qsfp monitoring daemon, publishes to redis.
    
    > goes qsfp help
    
     usage: qsfp
    
    >]

#### Examples

    > goes qsfp

### 2.2.52 goes reboot

#### Command

    > goes reboot

#### Description

     Reboot the system.
    
    > goes reboot help
    
     usage: reboot
    
    >]

#### Examples

    > goes reboot

### 2.2.53 goes redisd

#### Command

    > goes redisd

#### Description

     A redis server.
    
    > goes redisd help
    
     usage: redisd [-port PORT] [-set FIELD=VALUE]... [DEVICE]...
    
    >]

#### Examples

    > goes redisd
    
listen unix @platina-mk1/redisd: bind: address already in use
    
    >

### 2.2.54 goes reload

#### Command

    > goes reload

#### Description

     SIGHUP this goes machine.
    
    > goes reload help
    
     usage: reload
    
    >]

#### Examples

    > goes reload

### 2.2.55 goes restart

#### Command

    > goes restart

#### Description

     Stop, then start this goes machine.
    
    > goes restart help
    
     usage: restart [STOP, STOP, and REDISD OPTIONS]...
    
    >]

#### Examples

    > goes restart
    
    >

### 2.2.56 goes rm

#### Command

    > goes rm

#### Description

     Remove files or directories.
    
    > goes rm help
    
     usage: rm [OPTION]... FILE...
    
    >]

#### Examples

    > goes ls
    
     goes-platina-mk1-installer
    
     goesd-platina-mk1-modprobe.conf
    
     goesd-platina-mk1-modules.conf
    
     goesd-platina-mk1-sysctl.conf
    
     linux-image-platina-mk1-4.13.0.deb
    
     .ansible .profile cb.rom
    
     .bash_history .ssh coreboot-platina-mk1.rom
    
     .bashrc .vim testfile.txt
    
     .config .viminfo testfile55.txt
    
     .history_quagga .vimrc tools
    
     .lesshst .vimrc.swp volumes
    
    > goes rm testfile55.txt
    
    >
    
    > goes ls
    
     goes-platina-mk1-installer
    
     goesd-platina-mk1-modprobe.conf
    
     goesd-platina-mk1-modules.conf
    
     goesd-platina-mk1-sysctl.conf
    
     linux-image-platina-mk1-4.13.0.deb
    
     .ansible .profile cb.rom
    
     .bash_history .ssh coreboot-platina-mk1.rom
    
     .bashrc .vim testfile.txt
    
     .config .viminfo tools
    
     .history_quagga .vimrc volumes
    
     .lesshst .vimrc.swp
    
    >

### 2.2.57 goes rmmod

#### Command

    > goes rmmod

#### Description

     Remove a module from the Linux Kernel.
    
    > goes rmmod help
    
     usage: rmmod [OPTION]... MODULE...
    
    >]

#### Examples

    > goes rmmod

### 2.2.58 goes show

#### Command

    > goes show

#### Description

     Print stuff.
    
    > goes show help
    
     usage: show OBJECT
    
    >]

#### Examples

    > goes show

### 2.2.59 goes sleep

#### Command

    > goes sleep

#### Description

     Suspend execution for an interval of time.
    
    > goes sleep help
    
     usage: sleep SECONDS
    
    >]

#### Examples

    > goes sleep

### 2.2.60 goes source

#### Command

    > goes source

#### Description

     Import command script.
    
    > goes source help
    
     usage: source [-x] FILE
    
    >]

#### Examples

    > goes source

### 2.2.61 goes start

#### Command

    > goes start

#### Description

     Start this goes machine.
    
    > goes start help
    
     usage: start [-start=URL] [REDIS OPTIONS]...
    
    >]

#### Examples

    > goes start

### 2.2.62 goes status

#### Command

    > goes status

#### Description

     Print status of goes daemons.
    
    > goes status help
    
     usage: status
    
    >]

#### Examples

     root@invader29:~# goes status
    
     GOES status
    
     ======================
    
     Mode - SRIOV
    
     PCI - OK
    
     Check daemons - OK
    
     Check Redis - OK
    
     Check vnet - OK
    
     root@invader29:~#

### 2.2.63 goes stop

#### Command

    > goes stop

#### Description

     Stop this goes machine.
    
    > goes stop help
    
     usage: stop [-stop=URL] [SIGNAL]
    
    >]

#### Examples

    > goes stop

### 2.2.64 goes stty

#### Command

    > goes stty

#### Description

     Print info for given or current TTY.
    
    > goes stty help
    
     usage: stty [DEVICE]
    
    >]

#### Examples

    > goes stty
    
     name /dev/pts/2
    
     speed 38400 baud; rows ; columns ; line 0;
    
     intr = ^C; quit = ^; erase = ^?; kill = ^U; eof = ^D; min =
     ^A; start = ^Q; stop = ^S; susp = ^Z; reprint = ^R; discard =
     ^O; werase = ^W; lnext = ^V;
    
-parenb -parodd -cmspar cs8 -hupcl -cstopb cread -clocal -crtscts
    
-ignbrk -brkint -ignpar -parmrk -inpck -istrip -inlcr -igncr icrnl
     ixon -ixoff -iuclc -ixany -imaxbel -iutf8
    
opost -olcuc -ocrnl onlcr -onocr -onlret -ofill -ofdel
    
isig icanon iexten echo echoe echok -echonl -noflsh -xcase -tostop
     echoctl -echoprt echoke -flusho -pendin
    
    >

### 2.2.65 goes subscribe

#### Command

    > goes subscribe

#### Description

Print messages published to given redis channel.
    
    > goes subscribe help
    
     usage: subscribe CHANNEL
    
    >]

#### Examples

    > goes subscribe

### 2.2.66 goes sync

#### Command

    > goes sync

#### Description

     Flush the file system buffers.
    
    > goes sync help
    
     usage: sync
    
    >]

#### Examples

    > goes sync

### 2.2.67 goes tempd

#### Command

    > goes tempd

#### Description

Temperature monitoring daemon, publishes to redis.
    
    > goes tempd help
    
     usage: tempd
    
    >]

#### Examples

    > goes tempd

### 2.2.68 goes then

#### Command

    > goes then

#### Description

     Conditionally execute commands.
    
    > goes then help
    
usage: if COMMAND ; then COMMAND else COMMAND endif
    
    >]

#### Examples

    > goes then

### 2.2.69 goes toggle

#### Command

    > goes toggle

#### Description

     Toggle console port between x86 and BMC.
    
    > goes toggle help
    
     usage: toggle SECONDS
    
    >]

#### Examples

    > goes toggle

### 2.2.70 goes true

#### Command

    > goes true

#### Description

     Returns 'successful' not matter what.
    
    > goes true help
    
     usage: true
    
    >]

#### Examples

    > goes true

### 2.2.71 goes umount

#### Command

    > goes umount

#### Description

     Deactivate file-systems.
    
    > goes umount help
    
     usage: umount [OPTION]... FILESYSTEM|DIR
    
    >]

#### Examples

    > goes umount

### 2.2.72 goes uninstall

#### Command

    > goes uninstall

#### Description

     Uninstall this goes machine.
    
    > goes uninstall help
    
     usage: uninstall
    
    >]

#### Examples

    > goes uninstall

### 2.2.73 goes upgrade

#### Command

    > goes upgrade

#### Description

     Upgrade images.
    
    > goes upgrade help
    
     usage: upgrade [-v VER] [-s SERVER[/dir]] [-r] [-l] [-t]
     [-a | -g -k -c] [-f]
    
    >]

#### Examples

    > goes upgrade

### 2.2.74 goes uptimed

#### Command

    > goes uptimed

#### Description

     Record system uptime in redis.
    
    > goes uptimed help
    
     usage: uptimed
    
    >]

#### Examples

    > goes uptimed

### 2.2.75 goes vnet

#### Command

    > goes vnet

#### Description

     Send commands to hidden CLI.
    
    > goes vnet help
    
     usage:
    
    >]

#### Examples

    > goes vnet

### 2.2.75.1 goes vnet show event-log

#### Command

    > goes vnet show event-log

#### Description

     Event log commands.

#### Examples

    > goes vnet show event-log
    
    >

### 2.2.75.2 goes vnet show buffers

#### Command

    > goes vnet show buffers

#### Description

     Shows DMA(direct memory access) buffer usage.

#### Examples

    > goes vnet show buffers
    
     DMA heap: used 5.27M, free 73.88K, capacity 256M
    
     Pool Size Free Used
    
     default 1088 0 0
    
     fe1-rx 1088 217.31K 4.94M
    
     pg0 1088 0 0
    
    >

### 2.2.75.3 goes vnet show errors

#### Command

    > goes vnet show errors

#### Description

     Shows Error counters..

#### Examples

    > goes vnet show errors
    
     Node Error Count
    
     fe1-rx management duplicate 9
    
     fe1-rx-vlan-redirect not vlan tagged 16
    
    >

### 2.2.75.4 goes vnet show fe1 adj

#### Command

    > goes vnet show fe1 adj

#### Description

     Shows adjacencies.
    
     There are different options that can be used:

12. goes vnet show fe1 adj detail

13. goes vnet show fe1 adj sw

14. goes vnet show fe1 adj raw

15. goes vnet show fe1 adj detail eth-1-1

16. goes vnet show fe1 adj sw eth-1-1

17. goes vnet show fe1 adj raw eth-1-1

#### Examples

    > goes vnet show fe1 adj sw eth-1-1
    
sw/asic rx_pipe index free/used ipAdj type port drop copyCpu ...
tx_pipe type dstRewrite srcRewrite classID dstAddr if_index si_name
    
hardware 0 2 used nil unicast meth-1 false false ... 1 l3_unicast
     false false 0 00:00:00:00:00:00 2 eth-1-1
    
software 0 2 used nil unicast meth-1 false false ... 1 l3_unicast
     false false 0 00:00:00:00:00:00 2 eth-1-12
    
    >

### 2.2.75.5 goes vnet show fe1 debug-events

#### Command

    > goes vnet show fe1 debug-events

#### Description

     Shows Debug events.

#### Examples

    > goes vnet show fe1 debug-events
    
     no debug events
    
    >

### 2.2.75.6 goes vnet show fe1 eye-summary

#### Command

    > goes vnet show fe1 eye-summary

#### Description

     Shows Eye-Summary.
    
     This command has 2 options

18. goes vnet show fe1 eye-summary :- Shows full eye-summary

19. goes vnet show fe1 eye-summary [Interface] :-Shows eye-summary of
    that interface

#### Examples

    > goes vnet show fe1 eye-summary eth-2-1
    
Port Lane Horizontal/Vertical Raw Left/Lower Raw Right/Upper
     Width/Height Center
    
     eth-2-1 lane1 Horizontal 0x0 0x0 0 mUI 0%
    
     eth-2-1 lane1 Vertical 0x0 0x0 0 mV 0%
    
     eth-2-1 lane2 Horizontal 0x0 0x0 0 mUI 0%
    
     eth-2-1 lane2 Vertical 0x0 0x0 0 mV 0%
    
     eth-2-1 lane3 Horizontal 0x0 0x0 0 mUI 0%
    
     eth-2-1 lane3 Vertical 0x0 0x0 0 mV 0%
    
     eth-2-1 lane4 Horizontal 0x0 0x0 0 mUI 0%
    
     eth-2-1 lane4 Vertical 0x0 0x0 0 mV 0%
    
    >

### 2.2.75.7 goes vnet show fe1 eyescan

#### Command

    > goes vnet show fe1 eyescan

#### Description

     Shows eye-scanning results.

#### Examples

    > goes vnet show fe1 eyescan
    
     Starting eye scan for port eth-1-1, lane 0
    
     Failed: pmd not locked
    
     Starting eye scan for port eth-1-1, lane 1
    
     Failed: pmd not locked
    
     Starting eye scan for port eth-1-1, lane 2
    
     Failed: pmd not locked
    
     Starting eye scan for port eth-1-1, lane 3
    
     Failed: pmd not locked
    
     ...
    
     ...
    
    >

### 2.2.75.8 goes vnet show fe1 interrupt

#### Command

    > goes vnet show fe1 interrupt

#### Description

     Shows corresponding interrupts.

#### Examples

    > goes vnet show fe1 interrupt
    
     Interrupt Count Rate
    
     packet dma ch 0 desc controlled 103 4.14e-04
    
     packet dma ch 0 desc done 103 4.14e-04
    
     phy scan link status 18 7.24e-05
    
     pio done 34 1.37e-04
    
     sbus dma ch 0 done 13209373 5.31e+01
    
     sbus dma ch 1 done 13209374 5.31e+01
    
     sbus dma ch 2 done 13209373 5.31e+01
    
     polls 39495285 1.59e+02
    
    >

### 2.2.75.9 goes vnet show fe1 l2

#### Command

    > goes vnet show fe1 l2

#### Description

     [not sure]

#### Examples

    > goes vnet show fe1 l2
    
    >

### 2.2.75.10 goes vnet show fe1 temperature

#### Command

    > goes vnet show fe1 temperature

#### Description

     Shows Temperature data.

#### Examples

    > goes vnet show fe1 temperature
    
     0: &{current:58 max:61 min:55}
    
     1: &{current:57 max:62 min:55}
    
     2: &{current:59 max:63 min:57}
    
     3: &{current:58 max:62 min:56}
    
     4: &{current:57 max:61 min:55}
    
     5: &{current:58 max:63 min:57}
    
     6: &{current:57 max:61 min:55}
    
     7: &{current:57 max:60 min:54}
    
    >

### 2.2.75.11 goes vnet show fe1 switches

#### Command

    > goes vnet show fe1 switches

#### Description

     Shows Switches.

#### Examples

    > goes vnet show fe1 switches
    
     0000:04:00.0: id 0xb960, rev 0x12
    
    >

### 2.2.75.12 goes vnet show fe1 phy event-log

#### Command

    > goes vnet show fe1 phy event-log

#### Description

     Shows event logs on every physical port.

#### Examples

    > goes vnet show fe1 phy event-log
    
     port-2(ce01): event log
    
     7.0000e-05: lane 0, uc_entry_to_core_reset
    
     1.4372e+02: lane 0, uc_stop_event_log
    
     port-2(ce01): event log
    
     5.1449e-01: lane 0, uc_stop_event_log
    
     port-2(ce01): event log
    
     8.4147e-01: lane 0, uc_stop_event_log
    
     port-2(ce01): event log
    
     5.1418e-01: lane 0, uc_stop_event_log
    
     port-1(ce00): event log
    
     8.0000e-05: lane 0, uc_entry_to_core_reset
    
     1.4502e+02: lane 0, uc_stop_event_log
    
     ...
    
     ...
    
    >

### 2.2.75.13 goes vnet show fe1 pipe-counters

#### Command

    > goes vnet show fe1 pipe-counters

#### Description

     ??.

#### Examples

    > goes vnet show fe1 pipe-counters
    
     Name Pipe Packets Bytes
    
     eth-17-1 rx l3 interface 2 9 1034
    
     eth-17-1 rxf rule 2 7 774
    
     key: l3 if: 0x1/ffff
    
     rx port: 0x64/ff
    
     redirect: unicast port eth-17-1
    
     l3 switch change: l2 switch 0
    
     eth-19-1 rx l3 interface 2 9 1034
    
     eth-19-1 rxf rule 2 7 774
    
     key: l3 if: 0x3/ffff
    
     rx port: 0x64/ff
    
     redirect: unicast port eth-19-1
    
     l3 switch change: l2 switch 0
    
     ...
    
     ...
    
     meth-2 vlan-translate 2 7 774
    
     key: outer id 118=0x76, src: src port meth-2
    
     result: outer: 0, inner: 0, l3 if: 0xd, tag action: 1
    
     meth-2 vlan-translate 2 7 774
    
     key: outer id 126=0x7e, src: src port meth-2
    
     result: outer: 0, inner: 0, l3 if: 0xf, tag action: 1
    
    >

### 2.2.75.14 goes vnet show fe1 port-status mac

#### Command

    > goes vnet show fe1 port-status mac

#### Description

     Shows MAC Port-Status.

#### Examples

    > sudo goes vnet show fe1 port-status mac
    
Name Link Up Signal Detect Pmd Lock RemoteF Eq Link Down LocalF Eq
     Link Down
    
     eth-1-1 false false false true false
    
     eth-2-1 false false false true false
    
     eth-3-1 false false false true false
    
     ...
    
     ...
    
     meth-1 true true true true false
    
     meth-2 true true true true false
    
    >

### 2.2.75.15 goes vnet show fe1 port-status phy

#### Command

    > goes vnet show fe1 port-status phy

#### Description

     Shows Physical Port-Status.
    
     This command has 3 options:
    
     1) goes vnet show fe1 port-status phy 
    
     2) goes vnet show fe1 port-status phy [Interface]
    
3) goes vnet show fe1 port-status phy [Interface] detail

#### Examples

    > goes vnet show fe1 port phy eth-1-1
    
Port Lane Live Link Signal Detect Admin Down Pmd Lock Sigdet Sts Speed
     Autonegotiate Cl72 Fec
    
     eth-1-1 0 true true false true 0x211 100G X4 CR done ready cl91
    
     eth-1-1 1 false true false true 0x211 100G X4 CR ready cl91
    
     eth-1-1 2 false true false true 0x211 100G X4 CR ready cl91
    
     eth-1-1 3 false true false true 0x211 100G X4 CR ready cl91
    
    >
    
    > goes vnet show fe1 port phy eth-1-1 detail
    
Port Lane Live Link Signal Detect Admin Down Pmd Lock Sigdet Sts Speed
     Autonegotiate Cl72 Fec
    
     eth-1-1 0 true true false true 0x211 100G X4 CR done ready cl91
    
     eth-1-1 1 false true false true 0x211 100G X4 CR ready cl91
    
     eth-1-1 2 false true false true 0x211 100G X4 CR ready cl91
    
     eth-1-1 3 false true false true 0x211 100G X4 CR ready cl91
    
     eth-1-1
    
     cl82 stats
    
     deskew_align_aquired : false
    
     deskew_loss_of_alignment : true
    
     deskew_r_type : 0x10
    
     deskew_rxsm_state : 0x2
    
     cl82_ber : 0
    
     cl82_tx_idle_del_underflow : false
    
     cl82_t_type_coded : 4
    
     cl82_txsm_state : 1
    
     pcs stats : lane0 lane1 lane2 lane3
    
     live_link : true false false false
    
     rx_deskew_achieved : true false false false
    
     rx_high_bit_error : false false false false
    
rx_low_power_idle_received : false false false false
    
     rx_link_interrupt : false false false false
    
     rx_remote_fault : false false false false
    
     rx_local_fault : false false false false
    
tx_low_power_idle_received : false false false false
    
     tx_link_interrupt : false false false false
    
     tx_remote_fault : false false false false
    
     tx_local_fault : false false false false
    
     sc stats
    
     cl72_en : true true true true
    
     os_mode : 0 0 0 0
    
     t_fifo_mode : 0x2 0x2 0x2 0x2
    
     t_enc_mode : 0x2 0x2 0x2 0x2
    
     t_hg2_en : false false false false
    
     t_pma_btmx_mode : 0x0 0x0 0x0 0x0
    
     scr_mode : 0x3 0x3 0x3 0x3
    
     descr_mode : 0x2 0x2 0x2 0x2
    
     dec_tl_mode : 0x2 0x2 0x2 0x2
    
     deskew_mode : 0x6 0x6 0x6 0x6
    
     dec_fsm_mode : 0x2 0x2 0x2 0x2
    
     r_hg2_en : false false false false
    
     block_sync_mode : clause82 clause82 clause82 clause82
    
     block_sync_enable : false false false false
    
     block_dist_mode : 0 0 0 0
    
     block_bitmux_mode : 0 0 0 0
    
     sc_fsm_status : done start start start
    
    >

### 2.2.75.16 goes vnet show fe1 serdes-param

#### Command

    > goes vnet show fe1 serdes-param

#### Description

     ???.

#### Examples

    > goes vnet show fe1 serdes-param
    
Name Lane Main Pre Post1 Post2 Post3 Amplitude Drive sdk TxDis pmd
     TxEna
    
     eth-1-1 lane1 100 12 44 0 0 8 0 0 1
    
     eth-1-1 lane2 100 12 44 0 0 8 0 0 1
    
     eth-1-1 lane3 100 12 44 0 0 8 0 0 1
    
     eth-1-1 lane4 100 12 44 0 0 8 0 0 1
    
     ...
    
     ...
    
interface conversion: m.Phyer is *phy.FortyGig, not *phy.HundredGig
    
    >
    
    > goes vnet show fe1 serdes-param
     eth-1-1
    
Name Lane Main Pre Post1 Post2 Post3 Amplitude Drive sdk TxDis pmd
     TxEna
    
     eth-1-1 lane1 100 12 44 0 0 8 0 1 1
    
     eth-1-1 lane2 100 12 44 0 0 8 0 1 1
    
     eth-1-1 lane3 100 12 44 0 0 8 0 1 1
    
     eth-1-1 lane4 100 12 44 0 0 8 0 1 1
    
    >

### 2.2.75.17 goes vnet show fe1 l3-interface

#### Command

    > goes vnet show fe1 l3-interface

#### Description

     Shows Interface details.
    
     This command has different options:

20. goes vnet show fe1 l3-interface 

21. goes vnet show fe1 l3-interface tx

22. goes vnet show fe1 l3-interface rx 

23. goes vnet show fe1 l3-interface tx pipe [Interface]

24. goes vnet show fe1 l3-interface tx pipe raw eth-1-1

25. goes vnet show fe1 l3-interface rx pipe eth-1-1

26. goes vnet show fe1 l3-interface rx pipe raw eth-1-1

#### Examples

    > goes vnet show fe1 l3-interface
    
     rx pipe 0=======================================
    
software_intf:pg0 tunnel:no tunnel is_tx_punt:false if_index:0
     ref_count:0 disposition:0
    
     rx_vlan_translate: no
    
     punt_redirect: rx_l3_if_index:0, rxf_class_id:nil
    
     punt_config: not punt
    
     adjacency: is_ecmp:false, rxf_class_id:nil, index:0
    
     rx_l3_interface_entry 0: software / hardware
    
     vrf: 0 / 0
    
     rx_l3_interface_profile_index: 0 / 0
    
     rxf_class_id: 0 / 0
    
     ip_multicast_interface_for_lookup_keys: 0 / 0
    
     ip_option_profile_index: 0 / 0
    
     active_rx_l3_interface_profile_index: 0 / 0
    
     src_realm_id: 0 / 0
    
     tunnel_termination_ecn_decap_mapping_pointer: 0 / 0
    
     pipe_counter_ref: invalid / invalid
    
software_intf:fe1-cpu tunnel:no tunnel is_tx_punt:false if_index:1
     ref_count:1 disposition:0
    
     rx_vlan_translate: no
    
     punt_redirect: rx_l3_if_index:1, rxf_class_id:punt
    
     punt_config: next:punt, RefOpaque{intf:fe1-cpu, aux:0},
     advanceL3Header:false, NReplaceVlanHeaders:0,
     ReplacementTypeAndTag:[{0x0 0} {0x0 0}]
    
     adjacency: is_ecmp:false, rxf_class_id:punt, index:1
    
     rx_l3_interface_entry 1: software / hardware
    
     vrf: 0 / 0
    
     rx_l3_interface_profile_index: 0 / 0
    
     rxf_class_id: 0 / 0
    
     ip_multicast_interface_for_lookup_keys: 0 / 0
    
     ip_option_profile_index: 0 / 0
    
     active_rx_l3_interface_profile_index: 0 / 0
    
     src_realm_id: 0 / 0
    
     tunnel_termination_ecn_decap_mapping_pointer: 0 / 0
    
pipe_counter_ref: {pool_index:0, mode:0, index:1} / {pool_index:0,
     mode:0, index:1}
    
software_intf:eth-1-1 tunnel:no tunnel is_tx_punt:false if_index:2
     ref_count:1 disposition:0
    
     rx_vlan_translate: no
    
     punt_redirect: rx_l3_if_index:2, rxf_class_id:punt
    
     punt_config: next:punt, RefOpaque{intf:eth-1-1, aux:0},
     advanceL3Header:false, NReplaceVlanHeaders:0,
     ReplacementTypeAndTag:[{0x0 0} {0x0 0}]
    
     adjacency: is_ecmp:false, rxf_class_id:punt, index:2
    
     rx_l3_interface_entry 2: software / hardware
    
     vrf: 0 / 0
    
     rx_l3_interface_profile_index: 0 / 0
    
     rxf_class_id: 0 / 0
    
     ip_multicast_interface_for_lookup_keys: 0 / 0
    
     ip_option_profile_index: 0 / 0
    
     active_rx_l3_interface_profile_index: 0 / 0
    
     src_realm_id: 0 / 0
    
     tunnel_termination_ecn_decap_mapping_pointer: 0 / 0
    
pipe_counter_ref: {pool_index:0, mode:0, index:2} / {pool_index:0,
     mode:0, index:2}
    
     ...
    
     ...
    
software_intf:eth-32-1 tunnel:no tunnel is_tx_punt:false
     if_index:8 ref_count:1 disposition:0
    
     rx_vlan_translate: no
    
     punt_redirect: rx_l3_if_index:8, rxf_class_id:punt
    
     punt_config: next:punt, RefOpaque{intf:eth-32-1, aux:0},
     advanceL3Header:false, NReplaceVlanHeaders:0,
     ReplacementTypeAndTag:[{0x0 0} {0x0 0}]
    
     adjacency: is_ecmp:false, rxf_class_id:punt, index:16
    
     rx_l3_interface_entry 8: software / hardware
    
     vrf: 0 / 0
    
     rx_l3_interface_profile_index: 0 / 0
    
     rxf_class_id: 0 / 0
    
     ip_multicast_interface_for_lookup_keys: 0 / 0
    
     ip_option_profile_index: 0 / 0
    
     active_rx_l3_interface_profile_index: 0 / 0
    
     src_realm_id: 0 / 0
    
     tunnel_termination_ecn_decap_mapping_pointer: 0 / 0
    
pipe_counter_ref: {pool_index:0, mode:0, index:8} / {pool_index:0,
     mode:0, index:8}
    
    >
    
    > goes vnet show fe1 l3-interface tx pipe
     eth-1-1
    
     tx pipe 0=======================================
    
     tx pipe 1=======================================
    
software_intf:eth-1-1 tunnel:no tunnel is_tx_punt:true if_index:2
     ref_count:1 disposition:0 punt_disposition:0
    
     punt_config: next:punt, RefOpaque{intf:eth-1-1, aux:0},
     advanceL3Header:false, NReplaceVlanHeaders:0,
     ReplacementTypeAndTag:[{0x0 0} {0x0 0}]
    
     tx_l3_interface_entry 2: software / hardware
    
     ip_tunnel_index: 65535 / 8191
    
     ip_ttl_expired_threshold: 0 / 0
    
     ip_dscp_or_mapping_pointer: 0 / 0
    
     ip_dscp_select: 0 / 0
    
     txf_class_id: 0 / 0
    
     l2_switch: false / false
    
     inner_vlan: {id:0, priority_spec:{0 false 0 0}} / {id:0,
     priority_spec:{0 false 0 0}}
    
     outer_vlan: {id:6, priority_spec:{0 false 0 0}} / {id:6,
     priority_spec:{0 false 0 0}}
    
     inner_vlan_present_action: 0 / 0
    
     inner_vlan_add_if_absent: false / false
    
     src_ethernet_address: 50:18:4c:00:15:9c / 50:18:4c:00:15:9c
    
     tx pipe 2=======================================
    
     tx pipe 3=======================================
    
    >

### 2.2.75.18 goes vnet show fe1 port-map

#### Command

    > goes vnet show fe1 port-map

#### Description

     Shows Mapping of Ports.
    
     This command has different options:

27. goes vnet show fe1 port-map

28. goes vnet show fe1 port-map mmu

29. goes vnet show fe1 port-map phy

30. goes vnet show fe1 port-map vlan

31. goes vnet show fe1 port-map pipe

32. goes vnet show fe1 port-map [Interface]

33. goes vnet show fe1 port-map all

#### Examples

    > sudo goes vnet show fe1 port-map
    
     eth-2-1( 1) eth-1-1( 5) eth-4-1( 9) eth-3-1( 13)
    
     eth-6-1( 17) eth-5-1( 21) eth-8-1( 25) eth-7-1( 29)
    
     eth-10-1( 34) eth-9-1( 38) eth-12-1( 42) eth-11-1( 46)
    
     eth-14-1( 50) eth-13-1( 54) eth-16-1( 58) eth-15-1( 62)
    
     eth-18-1( 68) eth-17-1( 72) eth-20-1( 76) eth-19-1( 80)
    
     eth-22-1( 84) eth-21-1( 88) eth-24-1( 92) eth-23-1( 96)
    
     eth-26-1(102) eth-25-1(106) eth-28-1(110) eth-27-1(114)
    
     eth-30-1(118) eth-29-1(122) eth-32-1(126) eth-31-1(130)
    
     meth-1( 66) meth-2(100)
    
    >
    
    > goes vnet show fe1 port-map mmu
    
     eth-2-1( 0) eth-1-1( 2) eth-4-1( 4) eth-3-1( 6)
    
     eth-6-1( 8) eth-5-1( 10) eth-8-1( 12) eth-7-1( 14)
    
     eth-10-1( 64) eth-9-1( 66) eth-12-1( 68) eth-11-1( 70)
    
     eth-14-1( 72) eth-13-1( 74) eth-16-1( 76) eth-15-1( 78)
    
     eth-18-1(128) eth-17-1(130) eth-20-1(132) eth-19-1(134)
    
     eth-22-1(136) eth-21-1(138) eth-24-1(140) eth-23-1(142)
    
     eth-26-1(192) eth-25-1(194) eth-28-1(196) eth-27-1(198)
    
     eth-30-1(200) eth-29-1(202) eth-32-1(204) eth-31-1(206)
    
     meth-1( 96) meth-2(160)
    
    >
    
    > goes vnet show fe1 port-map vlan
    
     eth-2-1( 2) eth-1-1( 6) eth-4-1( 10) eth-3-1( 14)
    
     eth-6-1( 18) eth-5-1( 22) eth-8-1( 26) eth-7-1( 30)
    
     eth-10-1( 34) eth-9-1( 38) eth-12-1( 42) eth-11-1( 46)
    
     eth-14-1( 50) eth-13-1( 54) eth-16-1( 58) eth-15-1( 62)
    
     eth-18-1( 66) eth-17-1( 70) eth-20-1( 74) eth-19-1( 78)
    
     eth-22-1( 82) eth-21-1( 86) eth-24-1( 90) eth-23-1( 94)
    
     eth-26-1( 98) eth-25-1( 102) eth-28-1( 106) eth-27-1( 110)
    
     eth-30-1( 114) eth-29-1( 118) eth-32-1( 122) eth-31-1( 126)
    
     meth-1( 130) meth-2( 132)
    
    >

### 2.2.75.19 goes vnet show fe1 port-table

#### Command

    > goes vnet show fe1 port-table

#### Description

     Shows Port tables of different Interfaces.
    
     This command has 2 options
    
     1) goes vnet show fe1 port-table
    
     2) goes vnet show fe1 port-table dev [Interface]

#### Examples

    > goes vnet show fe1 port-table dev eth-3-1
    
port-table[ 13] oper:2 disc:false rxf:true vt:true ipv4:true cml:0
tag_action_idx:0 ovid:14 ivid:0 trust:false use_ivid:false
    
lport-table[ 13] oper:2 disc:false rxf:true vt:true ipv4:true cml:0
tag_action_idx:0 ovid:14 ivid:0 trust:false use_ivid:false
    
    >

### 2.2.75.20 goes vnet show fe1 prefix-pool

#### Command

    > goes vnet show fe1 prefix-pool

#### Description

Show ip46-prefix counter pool entries for all pipes.

#### Examples

    > goes vnet show fe1 prefix-pool
    
     ip46_prefix_counter_pool.entries for pipe 0
    
     0: class-id punt, vrf 0, 10.0.1.0/24, rule 6/28
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     1: class-id punt, vrf 0, 10.0.1.29/32, rule 6/0
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ...
    
     ...
    
     ...
    
     ...
    
     62: class-id punt, vrf 0, 10.0.32.0/24, rule 6/63
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     63: class-id punt, vrf 0, 10.0.32.29/32, rule 6/42
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ip46_prefix_counter_pool.entries for pipe 1
    
     0: class-id punt, vrf 0, 10.0.1.0/24, rule 6/28
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     1: class-id punt, vrf 0, 10.0.1.29/32, rule 6/0
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ...
    
     ...
    
     ...
    
     62: class-id punt, vrf 0, 10.0.32.0/24, rule 6/63
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     63: class-id punt, vrf 0, 10.0.32.29/32, rule 6/42
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ip46_prefix_counter_pool.entries for pipe 2
    
     0: class-id punt, vrf 0, 10.0.1.0/24, rule 6/28
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     1: class-id punt, vrf 0, 10.0.1.29/32, rule 6/0
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ...
    
     ...
    
     ...
    
     ...
    
     63: class-id punt, vrf 0, 10.0.32.29/32, rule 6/42
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ip46_prefix_counter_pool.entries for pipe 3
    
     0: class-id punt, vrf 0, 10.0.1.0/24, rule 6/28
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.0/24
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     1: class-id punt, vrf 0, 10.0.1.29/32, rule 6/0
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.1.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
     ...
    
     ...
    
     ...
    
     ...
    
     63: class-id punt, vrf 0, 10.0.32.29/32, rule 6/42
    
     rxf rule
    
     key: class id c [7:0]: 0x1/ff
    
     ip dst [31:0]: 10.0.32.29/32
    
     l3 type [3:0]: 0x0/c
    
     vrf [11:8]: 0x1/f
    
     vrf [7:0]: 0x0/ff
    
    >

### 2.2.75.21 goes vnet show fe1 station-tcam

#### Command

    > goes vnet show fe1 station-tcam

#### Description

     ???.
    
     This command has 2 options:

34. goes vnet show fe1 station-tcam

35. goes vnet show fe1 station-tcam detail

#### Examples

    > goes vnet show fe1 station-tcam
    
     MY_STATION_TCAM[0] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     MY_STATION_TCAM[1] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     MY_STATION_TCAM[2] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     ...
    
     ...
    
     MY_STATION_TCAM[36] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     MY_STATION_TCAM[37] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     MY_STATION_TCAM[38] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     total: 39 entries
    
    >
    
    > goes vnet show fe1 station-tcam detail
    
{result:{copy_to_cpu:false drop:false ip4_unicast_enable:true
     ip6_unicast_enable:true ip4_multicast_enable:true
ip6_multicast_enable:true mpls_enable:true arp_rarp_enable:true
fcoe_enable:false trill_enable:false mac_in_mac_enable:false}
key:{LogicalPort:{isTrunk:false module:0 number:0} Vlan:0
EthernetAddress:[0 0 0 0 0 0]} mask:{LogicalPort:{isTrunk:false
     module:0 number:0} Vlan:0 EthernetAddress:[0 0 0 0 0 0]} valid:true}
    
     MY_STATION_TCAM[0] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
{result:{copy_to_cpu:false drop:false ip4_unicast_enable:true
     ip6_unicast_enable:true ip4_multicast_enable:true
ip6_multicast_enable:true mpls_enable:true arp_rarp_enable:true
fcoe_enable:false trill_enable:false mac_in_mac_enable:false}
key:{LogicalPort:{isTrunk:false module:0 number:0} Vlan:0
EthernetAddress:[0 0 0 0 0 0]} mask:{LogicalPort:{isTrunk:false
     module:0 number:0} Vlan:0 EthernetAddress:[0 0 0 0 0 0]} valid:true}
    
     ...
    
     ...
    
     MY_STATION_TCAM[36] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
{result:{copy_to_cpu:false drop:false ip4_unicast_enable:true
     ip6_unicast_enable:true ip4_multicast_enable:true
ip6_multicast_enable:true mpls_enable:true arp_rarp_enable:true
fcoe_enable:false trill_enable:false mac_in_mac_enable:false}
key:{LogicalPort:{isTrunk:false module:0 number:0} Vlan:0
EthernetAddress:[0 0 0 0 0 0]} mask:{LogicalPort:{isTrunk:false
     module:0 number:0} Vlan:0 EthernetAddress:[0 0 0 0 0 0]} valid:true}
    
     MY_STATION_TCAM[37] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
{result:{copy_to_cpu:false drop:false ip4_unicast_enable:true
     ip6_unicast_enable:true ip4_multicast_enable:true
ip6_multicast_enable:true mpls_enable:true arp_rarp_enable:true
fcoe_enable:false trill_enable:false mac_in_mac_enable:false}
key:{LogicalPort:{isTrunk:false module:0 number:0} Vlan:0
EthernetAddress:[0 0 0 0 0 0]} mask:{LogicalPort:{isTrunk:false
     module:0 number:0} Vlan:0 EthernetAddress:[0 0 0 0 0 0]} valid:true}
    
     MY_STATION_TCAM[38] port={false 0 0} vlan=0 mac=00:00:00:00:00:00
    
     total: 39 entries
    
    >

### 2.2.75.22 goes vnet show fe1 tcam

#### Command

    > goes vnet show fe1 tcam

#### Description

     ????.
    
     This command has 3 options:
    
     1) goes vnet show fe1 tcam
    
     2) goes vnet show fe1 tcam details
    
     3) goes vnet show fe1 tcam sw

#### Examples

    > goes vnet show fe1 tcam
    
pipe_num prefix_index prefix_len n_half_entries base_index
    
     0 0 /32 64 0
    
     0 1 /31 0 64
    
     0 2 /30 0 64
    
     0 3 /29 0 64
    
     0 4 /28 0 64
    
     ...
    
     ...
    
     0 30 /2 0 128
    
     0 31 /1 0 128
    
     0 32 /0 0 128
    
     ...
    
     ...
    
pipe_num prefix_index prefix_len n_half_entries base_index
    
     3 0 /32 64 0
    
     3 1 /31 0 64
    
     3 2 /30 0 64
    
     ...
    
     ...
    
     3 28 /4 0 128
    
     3 29 /3 0 128
    
     3 30 /2 0 128
    
     3 31 /1 0 128
    
     3 32 /0 0 128
    
    >
    
    > goes vnet show fe1 tcam sw
    
pipe_num prefix_index prefix_len n_half_entries base_index
    
     0 0 /32 64 0
    
     0 1 /31 0 64
    
     0 2 /30 0 64
    
     ...
    
     ...
    
     0 30 /2 0 128
    
     0 31 /1 0 128
    
     0 32 /0 0 128
    
     ...
    
     ...
    
     ...
    
pipe_num prefix_index prefix_len n_half_entries base_index
    
     3 0 /32 64 0
    
     3 1 /31 0 64
    
     3 2 /30 0 64
    
     3 3 /29 0 64
    
     ...
    
     ...
    
     3 30 /2 0 128
    
     3 31 /1 0 128
    
     3 32 /0 0 128
    
    >

### 2.2.75.23 goes vnet show fe1 tx-pipe-buffer

#### Command

    > goes vnet show fe1 tx-pipe-buffer

#### Description

     ???.

#### Examples

    > goes vnet show fe1 tx-pipe-buffer
    
port isProvisioned (phys)current_usage pipe_port ( mi, mp) mmu_port
mmu_cell_req_outstanding buffer_port_enable buffer_soft_reset
     all_port
    
     fe1-cpu false ( 0) 0 0 ( 32, 0) 32 11 0x1 0x0 true
    
     eth-2-1 true ( 1) 0 1 ( 0, 0) 0 36 0x1 0x0 true
    
     eth-2-2 false ( 2) 0 2 ( 16, 0) 16 0 0x0 0x0 false
    
     ...
    
     ...
    
     eth-31-3 false (127) 0 132 ( 15, 3) 207 0 0x0 0x0 false
    
     eth-31-4 false (128) 0 133 ( 31, 3) 223 0 0x0 0x0 false
    
     meth-1 true (129) 0 66 ( 32, 1) 96 11 0x1 0x0 true
    
     meth-2 true (131) 0 100 ( 32, 2) 160 11 0x1 0x0 true
    
     fe1-pipe0-loopback false (132) 0 33 ( 33, 0) 33 36 0x1 0x0 false
    
     fe1-pipe1-loopback false (133) 0 67 ( 33, 1) 97 36 0x1 0x0 false
    
     fe1-pipe2-loopback false (134) 0 101 ( 33, 2) 161 36 0x1 0x0 false
    
     fe1-pipe3-loopback false (135) 0 135 ( 33, 3) 225 36 0x1 0x0 false
    
    >

### 2.2.75.24 goes vnet show fe1 visibility

#### Command

    > goes vnet show fe1 visibility

#### Description

     ????.

#### Examples

    > goes vnet show fe1 visibility
    
     pipe 0: &{[0 0] [0 0 0 0] [0 0 0 0 0 0 0 0]}
    
     pipe 1: &{[0 0] [0 0 0 0] [0 0 0 0 0 0 0 0]}
    
     pipe 2: &{[0 0] [0 0 0 0] [0 0 0 0 0 0 0 0]}
    
     pipe 3: &{[0 0] [0 0 0 0] [0 0 0 0 0 0 0 0]}
    
    >

### 2.2.75.25 goes vnet show fe1 vlan

#### Command

    > goes vnet show fe1 vlan

#### Description

     Shows VLAN information.

#### Examples

    > goes vnet show fe1 vlan
    
     vlan_tab[1]( fe1-cpu meth-1) egr_vlan[1]( fe1-cpu meth-1) ubm()
    
     vlan_tab[2]( fe1-cpu eth-2-1 meth-1) egr_vlan[2]( fe1-cpu
     eth-2-1 meth-1) ubm()
    
     vlan_tab[3]( fe1-cpu eth-2-2 meth-1) egr_vlan[3]( fe1-cpu
     eth-2-2 meth-1) ubm()
    
     vlan_tab[4]( fe1-cpu eth-2-3 meth-1) egr_vlan[4]( fe1-cpu
     eth-2-3 meth-1) ubm()
    
     vlan_tab[5]( fe1-cpu eth-2-4 meth-1) egr_vlan[5]( fe1-cpu
     eth-2-4 meth-1) ubm()
    
     vlan_tab[6]( fe1-cpu eth-1-1 meth-1) egr_vlan[6]( fe1-cpu
     eth-1-1 meth-1) ubm()
    
     vlan_tab[7]( fe1-cpu eth-1-2 meth-1) egr_vlan[7]( fe1-cpu
     eth-1-2 meth-1) ubm()
    
     vlan_tab[8]( fe1-cpu eth-1-3 meth-1) egr_vlan[8]( fe1-cpu
     eth-1-3 meth-1) ubm()
    
     vlan_tab[9]( fe1-cpu eth-1-4 meth-1) egr_vlan[9]( fe1-cpu
     eth-1-4 meth-1) ubm()
    
     ...
    
     ...
    
     vlan_tab[126]( fe1-cpu meth-2 eth-31-1) egr_vlan[126]( fe1-cpu
     meth-2 eth-31-1) ubm()
    
     vlan_tab[127]( fe1-cpu meth-2 eth-31-2) egr_vlan[127]( fe1-cpu
     meth-2 eth-31-2) ubm()
    
     vlan_tab[128]( fe1-cpu meth-2 eth-31-3) egr_vlan[128]( fe1-cpu
     meth-2 eth-31-3) ubm()
    
     vlan_tab[129]( fe1-cpu meth-2 eth-31-4) egr_vlan[129]( fe1-cpu
     meth-2 eth-31-4) ubm()
    
     vlan_tab[130]( fe1-cpu meth-1) egr_vlan[130]( fe1-cpu meth-1)
     ubm()
    
     vlan_tab[132]( fe1-cpu meth-2) egr_vlan[132]( fe1-cpu meth-2)
     ubm()
    
     vlan_tab[133]( fe1-cpu fe1-pipe0-loopback meth-1) egr_vlan[133](
     fe1-cpu fe1-pipe0-loopback meth-1) ubm()
    
     vlan_tab[134]( fe1-cpu meth-1 fe1-pipe1-loopback) egr_vlan[134](
     fe1-cpu meth-1 fe1-pipe1-loopback) ubm()
    
     vlan_tab[135]( fe1-cpu meth-2 fe1-pipe2-loopback) egr_vlan[135](
     fe1-cpu meth-2 fe1-pipe2-loopback) ubm()
    
     vlan_tab[136]( fe1-cpu meth-2 fe1-pipe3-loopback) egr_vlan[136](
     fe1-cpu meth-2 fe1-pipe3-loopback) ubm()
    
    >

### 2.2.75.26 goes vnet show hardware-interfaces

#### Command

    > goes vnet show hardware-interfaces

#### Description

     Shows hardware Interface details.

#### Examples

    > goes vnet show hardware-interfaces
    
     Name Address Link Counter Count
    
     eth-1-1 50:18:4c:00:15:9c up tx pipe unicast queue cos0 packets 7
    
     tx pipe unicast queue cos0 bytes 746
    
     tx pipe port table packets 7
    
     tx pipe port table bytes 746
    
     port tx packets 7
    
     port tx bytes 774
    
     port tx 65 to 127 byte packets 5
    
     port tx 128 to 255 byte packets 2
    
     port tx good packets 7
    
     port tx multicast packets 7
    
     eth-3-1 50:18:4c:00:15:a4 up tx pipe unicast queue cos0 packets 7
    
     tx pipe unicast queue cos0 bytes 746
    
     tx pipe port table packets 7
    
     tx pipe port table bytes 746
    
     port tx packets 7
    
     port tx bytes 774
    
     port tx 65 to 127 byte packets 5
    
     port tx 128 to 255 byte packets 2
    
     port tx good packets 7
    
     port tx multicast packets 7
    
     ...
    
     ...
    
     meth-2 00:00:00:00:00:00 up port rx packets 58
    
     port rx bytes 6724
    
     port rx 65 to 127 byte packets 40
    
     port rx 128 to 255 byte packets 18
    
     port rx good packets 58
    
     port rx multicast packets 58
    
     port rx 1tag vlan packets 56
    
     rx pipe port table packets 58
    
     rx pipe port table bytes 6492
    
     rx pipe ip6 header errors 17179869208
    
     rx pipe multicast drops 12884901946
    
     rx pipe vlan tagged packets 12884901944
    
     fe1-cpu 00:00:00:00:00:00 up tx pipe vlan tagged packets 125
    
     tx pipe cpu vlan-redirect packets 125
    
     tx pipe cpu vlan-redirect bytes 14370
    
     tx pipe port table packets 125
    
     tx pipe port table bytes 14370
    
    >

### 2.2.75.27 goes vnet show interfaces

#### Command

    > goes vnet show interfaces

#### Description

     Shows Interface details.

#### Examples

    > goes vnet show interfaces
    
     Name State Counter Count
    
     eth-1-1 up rx pipe l3 interface packets 7
    
     rx pipe l3 interface bytes 774
    
     eth-3-1 up rx pipe l3 interface packets 7
    
     rx pipe l3 interface bytes 774
    
     eth-5-1 up rx pipe l3 interface packets 7
    
     rx pipe l3 interface bytes 774
    
     eth-7-1 up rx pipe l3 interface packets 7
    
     rx pipe l3 interface bytes 774
    
     ...
    
     ...
    
     eth-31-1 up rx packets 2
    
     rx bytes 268
    
     rx pipe l3 interface packets 9
    
     rx pipe l3 interface bytes 1034
    
     meth-1 up rx packets 51
    
     rx bytes 5726
    
     meth-2 up rx packets 58
    
     rx bytes 6500
    
    >

### 2.2.75.28 goes vnet show ip fib

#### Command

    > goes vnet show ip fib

#### Description

     Shows IP Forwarding-information-base tables.

#### Examples

    > goes vnet show ip fib
    
     Table Destination Adjacency
    
     default 10.0.1.0/24 3: glean eth-1-1
    
     default 10.0.1.29/32 4: local eth-1-1
    
     default 10.0.2.0/24 5: glean eth-2-1
    
     default 10.0.2.29/32 6: local eth-2-1
    
     default 10.0.3.0/24 7: glean eth-3-1
    
     default 10.0.3.29/32 8: local eth-3-1
    
     default 10.0.4.0/24 9: glean eth-4-1
    
     default 10.0.4.29/32 10: local eth-4-1
    
     default 10.0.5.0/24 11: glean eth-5-1
    
     default 10.0.5.29/32 12: local eth-5-1
    
     ...
    
     ...
    
     default 10.0.29.29/32 60: local eth-29-1
    
     default 10.0.30.0/24 61: glean eth-30-1
    
     default 10.0.30.29/32 62: local eth-30-1
    
     default 10.0.31.0/24 63: glean eth-31-1
    
     default 10.0.31.29/32 64: local eth-31-1
    
     default 10.0.32.0/24 65: glean eth-32-1
    
     default 10.0.32.29/32 66: local eth-32-1
    
    >

### 2.2.75.29 goes vnet show netlink namespaces

#### Command

    > goes vnet show netlink namespaces

#### Description

     Shows netlink-namespaces.

#### Examples

    > goes vnet show netlink namespaces
    
     Interface Type Namespace NSID Si
    
     eth-1-1 vlan default 0x5
    
     eth-2-1 vlan default 0x1
    
     eth-3-1 vlan default 0xd
    
     eth-4-1 vlan default 0x9
    
     eth-5-1 vlan default 0x15
    
     eth-6-1 vlan default 0x11
    
     eth-7-1 vlan default 0x1d
    
     eth-8-1 vlan default 0x19
    
     eth-9-1 vlan default 0x25
    
     eth-10-1 vlan default 0x21
    
     eth-11-1 vlan default 0x2d
    
     eth-12-1 vlan default 0x29
    
     eth-13-1 vlan default 0x35
    
     eth-14-1 vlan default 0x31
    
     eth-15-1 vlan default 0x3d
    
     eth-16-1 vlan default 0x39
    
     eth-17-1 vlan default 0x45
    
     eth-18-1 vlan default 0x41
    
     eth-19-1 vlan default 0x4d
    
     eth-20-1 vlan default 0x49
    
     eth-21-1 vlan default 0x55
    
     eth-22-1 vlan default 0x51
    
     eth-23-1 vlan default 0x5d
    
     eth-24-1 vlan default 0x59
    
     eth-25-1 vlan default 0x65
    
     eth-26-1 vlan default 0x61
    
     eth-27-1 vlan default 0x6d
    
     eth-28-1 vlan default 0x69
    
     eth-29-1 vlan default 0x75
    
     eth-30-1 vlan default 0x71
    
     eth-31-1 vlan default 0x7d
    
     eth-32-1 vlan default 0x79
    
    >

### 2.2.75.30 goes vnet show netlink summary

#### Command

    > goes vnet show netlink summary

#### Description

     Shows Netlink Summary.

#### Examples

    > goes vnet show netlink summary
    
     Type Ignored Handled
    
     RTM_NEWADDR 7 47
    
     RTM_NEWLINK 7 47
    
     RTM_NEWNEIGH 3090 0
    
     RTM_NEWROUTE 57 158
    
     Total 3161 252
    
    >

### 2.2.75.31 goes vnet show runtime

#### Command

    > goes vnet show runtime

#### Description

     Shows runtime details.
    
     This command has 4 options:

36. goes vnet show runtime

37. goes vnet show runtime detail

38. goes vnet show runtime event

39. goes vnet show runtime next

#### Examples

    > goes vnet show runtime
    
     Vectors: 125, Vectors/sec: 4.85e-04, Clocks/vector: 21187.40,
     Vectors/call 0.62
    
     Name Index Calls Vectors Suspends Clocks
    
     error 4 100 125 0 5983.99
    
     fe1-rx 9 203 125 0 11709.84
    
     fe1-rx-vlan-redirect out 10 16 16 0 27293.50
    
    >
    
    > goes vnet show runtime event
    
     Events: 71367, Events/sec: 1.16e+00, Clocks/event: 14696320.87
    
     Name Events Suspends Clocks
    
     vnet 71367 2748374 14696320.87
    
    >

### 2.2.75.32 goes vnet show packet-generator

#### Command

    > goes vnet show packet-generator

#### Description

     Shows Packet-generator details.

#### Examples

    > goes vnet show packet-generator
    
     Node Name Limit Sent
    
    >

### 2.2.75.33 goes vnet clear errors

#### Command

    > goes vnet clear errors

#### Description

     Clears error counters.

#### Examples

    > goes vnet clear errors
    
    >

### 2.2.75.34 goes vnet clear event-log

#### Command

    > goes vnet clear event-log

#### Description

     Clears events in event log

#### Examples

    > goes vnet clear event-log
    
    >

### 2.2.75.35 goes vnet clear fe1 interrupt

#### Command

    > goes vnet clear fe1 interrupt

#### Description

     Clears Interrupt Configurations.

#### Examples

    > goes vnet clear fe1 interrupt
    
    >

### 2.2.75.36 goes vnet clear fe1 pipe-counters

#### Command

    > goes vnet clear fe1 pipe-counters

#### Description

     Clears fe1 pipe counters.

#### Examples

    > goes vnet clear fe1 pipe-counters
    
    >

### 2.2.75.37 goes vnet clear interfaces

#### Command

    > goes vnet clear interfaces

#### Description

     Clears interfaces statistics.

#### Examples

    > goes vnet clear interfaces
    
    >

### 2.2.75.38 goes vnet clear ip fib

#### Command

    > goes vnet clear ip fib

#### Description

     Clears IP4 forwarding table statistics.

#### Examples

    > goes vnet clear ip fib
    
    >

### 2.2.75.39 goes vnet clear netlink summary

#### Command

    > goes vnet clear netlink summary

#### Description

     Clears netlink summary counters.

#### Examples

    > goes vnet clear netlink summary
    
    >

### 2.2.75.40 goes vnet clear runtime

#### Command

    > goes vnet clear runtime

#### Description

     Clears main-loop runtime statistics.

#### Examples

    > goes vnet clear runtime
    
    >

### 2.2.75.41 goes vnet event-log

#### Command

    > goes vnet event-log

#### Description

     Event log commands.

#### Examples

    > goes vnet event-log
    
    >

### 2.2.75.42 goes vnet exec

#### Command

    > goes vnet exec

#### Description

     Executes the CLI commands from given file(s).

#### Examples

    > goes vnet exec
    
    >

### 2.2.75.43 goes vnet fe1 activate mirror

#### Command

    > goes vnet fe1 activate mirror

#### Description

     Fe1 activates the mirror profile <session name>.

#### Examples

    > goes vnet fe1 activate mirror
    
    >

### 2.2.75.44 goes vnet fe1 deactivate mirror

#### Command

    > goes vnet fe1 deactivate mirror

#### Description

     Fe1 deactivates the mirror profile <session name>.

#### Examples

    > goes vnet fe1 deactivate mirror
    
    >

### 2.2.75.45 goes vnet fe1 delete mirror

#### Command

    > goes vnet fe1 delete mirror

#### Description

     Fe1 deletes the mirror profile <session name>.

#### Examples

    > goes vnet fe1 delete mirror
    
    >

### 2.2.75.46 goes vnet fe1 disable sflow

#### Command

    > goes vnet fe1 disable sflow

#### Description

     Fe1 disables the sflow source<ingress if name>.

#### Examples

    > goes vnet fe1 disable sflow
    
    >

### 2.2.75.47 goes vnet fe1 enable ingress mirroring

#### Command

    > goes vnet fe1 enable ingress mirroring

#### Description

Fe1 enables ingress mirroring session <mirror-session> source
     <if-name> destination <if-name>.

#### Examples

    > goes vnet fe1 enable ingress mirroring
    
    >

### 2.2.75.48 goes vnet fe1 enable sflow 

#### Command

    > goes vnet fe1 enable sflow 

#### Description

Fe1 enables sflow source<ingress if name> cpu <true/false> mirror
     <true/false>.

#### Examples

    > goes vnet fe1 enable sflow 
    
    >

### 2.2.75.49 goes vnet fe1 get sflow counters

#### Command

    > goes vnet fe1 get sflow counters

#### Description

     Efe1 gets sflow counters.

#### Examples

    > goes vnet fe1 get sflow counters
    
    >

### 2.2.75.50 goes vnet fe1 set mirror encapsulation

#### Command

    > goes vnet fe1 set mirror encapsulation

#### Description

Fe1 sets mirror encapsulation session<session-name>
     destination<if-name> encap<SFLOW/RSPAN/ERSPAN>.

#### Examples

    > goes vnet fe1 set mirror encapsulation
    
    >

### 2.2.75.51 goes vnet fe1 set sflow default 

#### Command

    > goes vnet fe1 set sflow default 

#### Description

     Fe1 sets sflow default.

#### Examples

    > goes vnet fe1 set sflow default
    
    >

### 2.2.75.52 goes vnet fe1 set sflow target

#### Command

    > goes vnet fe1 set sflow target

#### Description

     Fe1 sets sflow target destination <session name>.

#### Examples

    > goes vnet fe1 set sflow target
    
    >

### 2.2.75.53 goes vnet fe1 show mirror

#### Command

    > goes vnet fe1 show mirror

#### Description

     Fe1 shows mirror session <session-name>.

#### Examples

    > goes vnet fe1 show mirror
    
    >

### 2.2.75.54 goes vnet fe1 show sflow details

#### Command

    > goes vnet fe1 show sflow details

#### Description

     Shows s-flow details of all ports.

#### Examples

    > goes vnet fe1 show sflow details
    
    >

### 2.2.75.55 goes vnet i2c scan

#### Command

    > goes vnet i2c scan

#### Description

     ????????????.

#### Examples

    > goes vnet i2c scan
    
    >

### 2.2.75.56 goes vnet ip interface

#### Command

    > goes vnet ip interface

#### Description

     IP interface commands.

#### Examples

    > goes vnet ip interface
    
    >

### 2.2.75.57 goes vnet ip route

#### Command

    > goes vnet ip route

#### Description

     Adds/deletes the Ip4/Ip6 routes.

#### Examples

    > goes vnet ip route
    
    >

### 2.2.75.58 goes vnet netlink log

#### Command

    > goes vnet netlink log

#### Description

     Enables/disables the netlink message logging.

#### Examples

    > goes vnet netlink log
    
    >

### 2.2.75.59 goes vnet netlink route

#### Command

    > goes vnet netlink route

#### Description

     Adds/deletes the Ip4/Ip6 routes via netlink.

#### Examples

    > goes vnet netlink route
    
    >

### 2.2.75.60 goes vnet packet-generator

#### Command

    > goes vnet packet-generator

#### Description

     Edits or creates the packet generator streams.

#### Examples

    > goes vnet packet-generator
    
    >

### 2.2.75.61 goes vnet set fe1 l_train_restart

#### Command

    > goes vnet set fe1 l_train_restart

#### Description

     Restarts the link training (initiate to remote).

#### Examples

    > goes vnet set fe1 l_train_restart
    
    >

### 2.2.75.62 goes vnet set fe1 pmd_restart

#### Command

    > goes vnet set fe1 pmd_restart

#### Description

     Restarts the Rx equalization.

#### Examples

    > goes vnet set fe1 pmd_restart
    
    >

### 2.2.75.63 goes vnet set fe1 port-config

#### Command

    > goes vnet set fe1 port-config

#### Description

     Sets fe1 port mac [rf|lf] [1|0].

#### Examples

    > goes vnet set fe1 port-config
    
    >

### 2.2.75.64 goes vnet set fe1 reset_port

#### Command

    > goes vnet set fe1 reset_port

#### Description

     Resets the port.

#### Examples

    > goes vnet set fe1 reset_port
    
    >

### 2.2.75.65 goes vnet set fe1 serdes-param

#### Command

    > goes vnet set fe1 serdes-param

#### Description

     Example: set fe1 serdes eth-1-1 lane 2 main 100 pre 12 post1 44 post2
     0 post3 0 amp 8 drv 0 lte 0. (Description??????)

#### Examples

    > goes vnet set fe1 serdes-param
    
    >

### 2.2.75.66 goes vnet set hardware-interface

#### Command

    > goes vnet set hardware-interface

#### Description

     Sets the hardware interface commands.

#### Examples

    > goes vnet set hardware-interface
    
    >

### 2.2.75.67 goes vnet set interface

#### Command

    > goes vnet set interface

#### Description

     Sets the interface commands.

#### Examples

    > goes vnet set interface
    
    >

### 2.2.76 goes vnetd

#### Command

    > goes vnetd

#### Description

     FIXME.
    
    > goes vnetd help
    
     usage: vnetd
    
    >]

#### Examples

    > goes vnetd

### 2.2.77 goes wget

#### Command

    > goes wget

#### Description

     A non-interactive network downloader.
    
    > goes wget help
    
     usage: wget URL...
    
    >]

#### Examples

    > goes wget


## 4.1 Redis hget commands

### 4.1.1 bmc.temperature

#### Command

    > sudo redis-cli -h 172.17.3.44 hget platina
    
     "bmc.temperature.units.C"

#### Description

     Shows BMC processor current temperature.

#### Examples

    > sudo redis-cli -h 172.17.3.44 hget platina
     "bmc.temperature.units.C"
    
     "43.15"
    
    >

### 4.1.2 cmdline.console

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.console"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.console"
    
     "ttymxc0,115200n8"
    
    >

### 4.1.3 cmdline.init

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.init"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.init"
    
     "/init"
    
    >

### 4.1.4 cmdline.initrd

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.initrd"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.initrd"
    
     "0x89000000,3M"
    
    >

### 4.1.5 cmdline.ip

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.ip"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.ip"
    
     "172.17.3.45::172.17.2.1:255.255.254.0::eth0:on"
    
    >

### 4.1.6 cmdline.mem

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.mem"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.mem"
    
     "1024m"
    
    >

### 4.1.7 cmdline.quiet

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.quiet"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.quiet"
    
     "true"
    
    >

### 4.1.8 cmdline.root

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.root"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.root"
    
     "/dev/ram0"
    
    >

### 4.1.9 cmdline.rootfstype 

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rootfstype"

#### Description

     Shows BMC processor's file system type.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rootfstype"
    
     "ext4"
    
    >

### 4.1.10 cmdline.rootwait

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rootwait"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rootwait"
    
     "true"
    
    >

### 4.1.11 cmdline.rw

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rw"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.rw"
    
     "true"
    
    >

### 4.1.12 cmdline.start

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.start"

#### Description

     ?????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "cmdline.start"
    
     "true"
    
    >

### 4.1.13 eeprom.BaseEthernetAddress

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.BaseEthernetAddress"

#### Description

     Shows mac address of BMC management interface.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.BaseEthernetAddress"
    
     "50:18:4c:00:15:98"
    
    >

### 4.1.14 eeprom.Crc

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Crc"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Crc"
    
     "0xd2c4147a"
    
    >

### 4.1.15 eeprom.DeviceVersion

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.DeviceVersion"

#### Description

     Shows BMC eprom device version.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.DeviceVersion"
    
     "0x0a"
    
    >

### 4.1.16 eeprom.ManufactureDate

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.ManufactureDate"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.ManufactureDate"
    
     "2017/06/27 13:38:39"
    
    >

### 4.1.17 eeprom.NEthernetAddress

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.NEthernetAddress"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.NEthernetAddress"
    
     "132"
    
    >

### 4.1.18 eeprom.Onie.Data

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Onie.Data"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Onie.Data"
    
     ""TlvInfox00" 0x546c76496e666f00"
    
    >

### 4.1.19 eeprom.Onie.Version

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Onie.Version"

#### Description

     Shows eprom Onie Version.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Onie.Version"
    
     "0x01"
    
    >

### 4.1.20 eeprom.PartNumber

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.PartNumber"

#### Description

     Shows eeprom part info.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.PartNumber"
    
     "PS-3001-32C-AFA"
    
    >

### 4.1.21 eeprom.PlatformName

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.PlatformName"

#### Description

     Shows eeprom platform name.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.PlatformName"
    
     "X86-BDE-4C-16GB-128GB"
    
    >

### 4.1.22 eeprom.SerialNumber

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.SerialNumber"

#### Description

     Shows eeprom serial number.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.SerialNumber"
    
     "FDU1757A0000B"
    
    >

### 4.1.23 eeprom.Vendor

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Vendor"

#### Description

     Shows eeprom vendor name.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.Vendor"
    
     "Platina Systems"
    
    >

### 4.1.24 eeprom.VendorExtension

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.VendorExtension"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "eeprom.VendorExtension"
    
""x00x00xbcePx01x00Qx01x00Rx01nSx0e900-000000-000Tx10main.HB1N7150012Tx0fcpu.HB2N7150007Tx0ffan.HB4N7100019""
    
    >

### 4.1.25 fan_tray.1.1.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.1.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.1.speed.units.rpm"
    
     "4066"
    
    >

### 4.1.26 fan_tray.1.2.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.2.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.2.speed.units.rpm"
    
     "4029"
    
    >

### 4.1.27 fan_tray.1.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.status"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.1.status"
    
     "ok.front->back"
    
    >

### 4.1.28 fan_tray.2.1.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.1.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.1.speed.units.rpm"
    
     "4115"
    
    >

### 4.1.29 fan_tray.2.2.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.2.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.2.speed.units.rpm"
    
     "4103"
    
    >

### 4.1.30 fan_tray.2.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.status"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.2.status"
    
     "ok.front->back"
    
    >

### 4.1.31 fan_tray.3.1.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.1.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.1.speed.units.rpm"
    
     "4258"
    
    >

### 4.1.32 fan_tray.3.2.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.2.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.2.speed.units.rpm"
    
     "4128"
    
    >

### 4.1.33 fan_tray.3.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.status"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.3.status"
    
     "ok.front->back"
    
    >

### 4.1.34 fan_tray.4.1.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.1.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.1.speed.units.rpm"
    
     "4218"
    
    >

### 4.1.35 fan_tray.4.2.speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.2.speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.2.speed.units.rpm"
    
     "4054"
    
    >

### 4.1.36 fan_tray.4.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.status"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.4.status"
    
     "ok.front->back"
    
    >

### 4.1.37 fan_tray.duty

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.duty"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.duty"
    
     "0x30"
    
    >

### 4.1.38 fan_tray.speed

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.speed"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "fan_tray.speed"
    
     "auto"
    
    >

### 4.1.39 host.temp.target.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "host.temp.target.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "host.temp.target.units.C"
    
     "70.00"
    
    >

### 4.1.40 host.temp.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "host.temp.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "host.temp.units.C"
    
     "50.00"
    
    >

### 4.1.41 hostname

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "hostname"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "hostname"
    
     "172.17.3.45"
    
    >

### 4.1.42 hwmon.front.temp.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "hwmon.front.temp.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "hwmon.front.temp.units.C"
    
     "46.250"
    
    >

### 4.1.43 hwmon.rear.temp.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     hwmon.rear.temp.units.C "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "hwmon.rear.temp.units.C"
    
     "50.250"
    
    >

### 4.1.44 hwmon.target.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "hwmon.target.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "hwmon.target.units.C"
    
     "50"
    
    >

### 4.1.45 machine

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "machine"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina "machine"
    
     "platina-mk1-bmc"
    
    >

### 4.1.46 psu.powercycle

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu.powercycle"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu.powercycle"
    
     "true"
    
    >

### 4.1.47 psu1.admin.state

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.admin.state"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.admin.state"
    
     "enabled"
    
    >

### 4.1.48 psu1.eeprom

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.eeprom"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.eeprom"
    
     "01000000010900f5010819c54757202020cb47572d4352505335353020ca58585858585858585858c3585846ce5053555134303030313132475720c0c0c10000000000000000003200021822c42602390319052823b036504620672f3f0c1f94c20000001f01020d09e701b0047404ec047800e803c8af01820d274982b0047404ec0478000000b80b2020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020"
    
    >

### 4.1.49 psu1.fan_direction

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu1.fan_direction "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.fan_direction"
    
     "front->back"
    
    >

### 4.1.50 psu1.fan_speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.fan_speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.fan_speed.units.rpm"
    
     "0"
    
    >

### 4.1.51 psu1.i_out.units.A

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.i_out.units.A"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.i_out.units.A"
    
     "0.019"
    
    >

### 4.1.52 psu1.mfg_id

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.mfg_id"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.mfg_id"
    
     "Great Wall"
    
    >

### 4.1.53 psu1.mfg_model

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.mfg_model"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.mfg_model"
    
     "CRPS550"
    
    >

### 4.1.54 psu1.p_in.units.W

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.p_in.units.W"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.p_in.units.W"
    
     "0.000"
    
    >

### 4.1.55 psu1.p_out.units.W

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.p_out.units.W"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.p_out.units.W"
    
     "0.000"
    
    >

### 4.1.56 psu1.sn

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.sn"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina "psu1.sn"
    
     "PSUQ4000112GW"
    
    >

### 4.1.57 psu1.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu1.status "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.status"
    
     "powered_off"
    
    >

### 4.1.58 psu1.temp1.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu1.temp1.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.temp1.units.C"
    
     "33.250"
    
    >

### 4.1.59 psu1.temp2.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu1.temp2.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.temp2.units.C"
    
     "29.344"
    
    >

### 4.1.60 psu1.v_in.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.v_in.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.v_in.units.V"
    
     "0.000"
    
    >

### 4.1.61 psu1.v_out.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.v_out.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu1.v_out.units.V"
    
     "0.000"
    
    >

### 4.1.62 psu2.admin.state

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.admin.state "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.admin.state "
    
     "enabled"
    
    >

### 4.1.63 psu2.eeprom

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.eeprom"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.eeprom"
    
     "
     01000000010900f5010819c54757202020cb47572d4352505335353020ca58585858585858585858c3585846ce5053555134303030313131475720c0c0c10000000000000000003300021822c42602390319052823b036504620672f3f0c1f94c20000001f01020d09e701b0047404ec047800e803c8af01820d274982b0047404ec0478000000b80b2020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020
     "
    
    >

### 4.1.64 psu2.fan_direction

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.fan_direction "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.fan_direction"
    
     "front->back"
    
    >

### 4.1.65 psu2.fan_speed.units.rpm

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.fan_speed.units.rpm"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.fan_speed.units.rpm"
    
     "4680"
    
    >

### 4.1.66 psu2.i_out.units.A

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.i_out.units.A"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.i_out.units.A"
    
     "11.031"
    
    >

### 4.1.67 psu2.mfg_id

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.mfg_id"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.mfg_id"
    
     "Great Wall"
    
    >

### 4.1.68 psu2.mfg_model

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.mfg_model"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.mfg_model"
    
     "CRPS550"
    
    >

### 4.1.69 psu2.p_in.units.W

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.p_in.units.W"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.p_in.units.W"
    
     "147.000"
    
    >

### 4.1.70 psu2.p_out.units.W

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.p_out.units.W"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.p_out.units.W"
    
     "133.000"
    
    >

### 4.1.71 psu2.sn

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.sn"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina "psu2.sn"
    
     "PSUQ4000111GW"
    
    >

### 4.1.72 psu2.status

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.status "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.status"
    
     "powered_on"
    
    >

### 4.1.73 psu2.temp1.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.temp1.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.temp1.units.C"
    
     "36.688"
    
    >

### 4.1.74 psu2.temp2.units.C

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     psu2.temp2.units.C"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.temp2.units.C"
    
     "42.125"
    
    >

### 4.1.75 psu2.v_in.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.v_in.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.v_in.units.V"
    
     "122.000"
    
    >

### 4.1.76 psu2.v_out.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.v_out.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "psu2.v_out.units.V"
    
     "12.047"
    
    >

### 4.1.77 redis.ready

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "redis.ready"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "redis.ready"
    
     "true"
    
    >

### 4.1.78 sys.cpu.load1

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load1"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load1"
    
     "0.08"
    
    >

### 4.1.79 sys.cpu.load10

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load10"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load10"
    
     "0.64"
    
    >

### 4.1.80 sys.cpu.load15

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load15"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.cpu.load15"
    
     "0.58"
    
    >

### 4.1.81 sys.mem.buffer

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.buffer"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.buffer"
    
     "352256"
    
    >

### 4.1.82 sys.mem.free

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.free"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.free"
    
     "938221568"
    
    >


4.1.83 sys.mem.shared

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.shared"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.shared"
    
     "4096"
    
    >

### 4.1.84 sys.mem.total

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.total"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.mem.total"
    
     "1058603008"
    
    >

### 4.1.85 sys.uptime

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "sys.uptime"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "sys.uptime"
    
     "5 weeks, 9 hours, 42 minutes"
    
    >

### 4.1.86 system.fan_direction

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "system.fan_direction"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "system.fan_direction"
    
     "front->back"
    
    >

### 4.1.87 vmon.1v0.tha.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     vmon.1v0.tha.units.V "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v0.tha.units.V"
    
     "1.04"
    
    >

### 4.1.88 vmon.1v0.thc.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina "
     vmon.1v0.thc.units.V "

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v0.thc.units.V"
    
     "1.016"
    
    >

### 4.1.89 vmon.1v2.ethx.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v2.ethx.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v2.ethx.units.V"
    
     "1.182"
    
    >

### 4.1.90 vmon.1v25.sys.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v25.sys.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v25.sys.units.V"
    
     "1.237"
    
    >

### 4.1.91 vmon.1v8.sys.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v8.sys.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.1v8.sys.units.V"
    
     "1.805"
    
    >

### 4.1.92 vmon.3v3.bmc.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.bmc.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.bmc.units.V"
    
     "3.32"
    
    >

### 4.1.93 vmon.3v3.sb.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.sb.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.sb.units.V"
    
     "3.334"
    
    >

### 4.1.94 vmon.3v3.sys.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.sys.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v3.sys.units.V"
    
     "3.276"
    
    >

### 4.1.95 vmon.3v8.bmc.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v8.bmc.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.3v8.bmc.units.V"
    
     "3.83"
    
    >

### 4.1.96 vmon.5v.sb.units.V

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.5v.sb.units.V"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.5v.sb.units.V"
    
     "5.051"
    
    >

### 4.1.97 vmon.poweroff.events

#### Command

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.poweroff.events"

#### Description

     ??????.

#### Examples

    > redis-cli -h 172.17.3.45 hget platina
     "vmon.poweroff.events"
    
     "1970-01-01T06:33:01Z.1970-01-01T07:07:34Z.1970-01-01T10:45:47Z.1970-01-01T02:43:25Z.1970-01-01T16:21:11Z"
    
    >

4.2 Redis hset commands

#### Command

     redis-cli -h 172.17.3.43 hset platina psu.powercycle true
    
     1

#### Description

     Power cycle host

#### Examples

    > redis-cli -h 172.17.3.43 hset platina
     psu.powercycle true
    
     1
