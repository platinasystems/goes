# PLATINA SYSTEMS CORPORATION
---
## INSTRUCTIONS FOR UPGRADING VERY OLD BMC
## 7/31/2018

## PROCEDURE 1 - Upgrading a very old bmc to latest code.

Instructions for a BMC with goes that has NO "upgrade" command.  u-boot will be used to upgrade.
This procedure is for a unit we have in the lab, i.e., we have the console port, access to tftp server.

Set WD so we don't reboot
=> run wd

Set active to qspi0
=> run qspi0                      we will do qspi0 in case something goes wrong

Set ip addresses
=> setenv ipaddr 192.168.101.198
=> setenv netmask 255.255.255.0
=> setenv gatewayip 192.168.101.2
=> setenv serverip 192.168.101.1  (platina1)
=> saveenv

read current qspi0 flash contents into memory
=> sf probe
=> sf read 80800000 0 00600000    read qspi0

Overwrite initrd and kernel in memory
=> run dw_initrd                  will pull from platina1: /srv/tftp 2285456 Jul 27 14:53 initrd.img.xz
=> run dw_kernel                  will pull from platina1: /srv/tftp 1649008 Jul 28 19:43 zImage

Burn memory to flash
=> run qspi                       writes ram to qspi0, writes env with new sizes
=> reset

NOW QSPI0 is running 7/27/2018 code

---

Use new upgrade command to do a full upgrade.

Pull version of choice
  Do full upgrade, all files, not just kernel and initrd, qspi0
  upgrade -v v0.41 -s 172.16.2.23:/downloads -f

Update qspi1 while we are at it
  Do full upgrade, all files, qspi1
  upgrade -v v0.41 -s 172.16.2.23:/downloads -f -1

NOW QSPI0 has v0.41 code
NOW QSPI1 has v0.41 code

DONE

---
---
---
## PROCEDURE 2 - Convert from Gen 1 to Gen 2 "upgrade" command

upgrading PURE remotely, in the event they have generation 1 upgrade format (tar file)
convert them to generation 2 upgrade format (zip file)

this assumes we have access only to the platinasystems.com server
this uses the old upgrade command to load old style format with new payload
after this is done the new upgrade command will be available.


Pure should have the latest upgrade format.  This would allow them to upgrade the bmc using  platinasystems.com.

There were two (2) formats for the upgrade command.  Pure initially had the original format, dated 6/9/2017.
A month later the new format was rolled out.  Pure was given a procedure to convert their units.  Not sure if they ever did.
Regardless, there is an easy path to convert units to the new format:

     upgrade -v LATEST -s 172.16.2.23

     This uses a file is on platina4:  /var/www/html/platina-mk1-bmc/platina-mk1-bmc-LATEST.tar
     This file is the new files stored in the old format.
     This only upgrades the kernel and initrd.
     A 2nd step is needed to do a full upgrade of all files, versions, dtb, u-boot, etc.
     This 2nd step also pulls in the latest version, v0.41 or whatever.


