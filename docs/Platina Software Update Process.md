# Platina Software Update Process

## Using GoES upgrade (recommended)

### Upgrading Coreboot

    sudo goes upgrade -c -v LATEST -s downloads.platinasystems.com

### Upgrading Linux Kernel

    sudo goes upgrade -k -v LATEST -s downloads.platinasystems.com

### Upgrading GoES

    sudo goes upgrade -g -v LATEST -s downloads.platinasystems.com

### Upgrading Coreboot, Linux Kernel, and GoES in one command

    sudo goes upgrade -a -v LATEST -s downloads.platinasystems.com

### Updating the BMC Firmware

---
You must upgrade the BMC using the GoES shell on the BMC console. It cannot be upgraded using the system GoES shell. For instructions on how to connect to the BMC console, see the section below titled _Getting to the BMC Console_.

From BMC console CLI, execute the command:

    upgrade -v LATEST -s downloads.platinasystems.com

This will automatically retrieve the upgrade file (platina-mk1-bmc.zip) from http://downloads.platinasystems.com/LATEST/, perform the update, then reboot the BMC to complete the upgrade. This BMC reboot only impacts the BMC and not the running system.

- _LATEST_ above may be replaced by any version control string (e.g. 'v0.2').
- _downloads.platinasystems.com_ above may be replaced with any reachable HTTP server hostname or IP address. For example: `upgrade -v v0.2 -s 192.168.101.127` will retrieve the URL http://192.168.101.127/v0.2/platina-mk1-bmc.zip.
- If the -s option is not specified, the default URL (http://downloads.platinasystems.com/) is used to retrieve the update file.
- To see a list of software revisions available through http://downloads.platinasystems.com/ execute the GoES command `upgrade -l`

---

## Manual upgrade (not recommended)
### Download the flash ROM and Coreboot images
To retrieve the flash ROM, execute the following Linux commands:

```
sudo bash
cd ~/
wget http://downloads.platinasystems.com/tools/flashrom
wget http://downloads.platinasystems.com/tools/platina-mk1.xml
chmod 655 flashrom
mv flashrom /usr/local/sbin/flashrom
mkdir -p /usr/local/share/flashrom/layouts
mv platina-mk1.xml /usr/local/share/flashrom/layouts
rm coreboot-platina-mk1.rom
wget http://downloads.platinasystems.com/LATEST/coreboot-platina-mk1.rom
```

### Install Coreboot
To update the boot loader, execute the following Linux commands on the host:
```
/usr/local/share/flashrom/layouts/platina-mk1.xml -i bios -w coreboot-platina-mk1.rom -A -V
```

### Update the Linux Kernel
To update the Linux kernel, execute the following Linux commands on the host:
```
sudo bash
wget http://downloads.platinasystems.com/LATEST/linux-image-platina-mk1-4.13.0.deb
dpkg -i linux-image-platina-mk1-4.13.0.deb
```

### Update GoES
To update the Platina GoES binary, execute the following Linux commands on the host:
```
sudo bash
wget http://downloads.platinasystems.com/LATEST/goes-platina-mk1-installer
chmod +x goes-platina-mk1-installer
./goes-platina-mk1-installer
```

## Additional Information

### Getting to the BMC Console
The RS-232 port on the front of the appliance is used to drive the system console of the Platina appliance _and_ the console of the Baseboard Management Controller (BMC). This section demonstrates how to toggle between the BMC console and the system console.

To switch to BMC console:

- At the system console (Linux) prompt, execute the command `sudo goes toggle`
- Hit return to redisplay the prompt
- The BMC console prompt should be displayed and looks like this:

    203.0.113.75>


To switch back to x86 console:

- At the BMC console prompt, execute the `toggle` GoES command.
- Hit return to display the Linux prompt on the system console

