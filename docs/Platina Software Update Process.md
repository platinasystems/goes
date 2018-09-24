# Platina Software Update Process
## Using GoES upgrade
### Upgrade Coreboot, Linux Kernel, GoES
---
 - sudo goes upgrade -c -v LATEST -s downloads.platinasystems.com
 - sudo goes upgrade -k -v LATEST -s downloads.platinasystems.com
 - sudo goes upgrade -g -v LATEST -s downloads.platinasystems.com
 - sudo goes upgrade -a -v LATEST -s downloads.platinasystems.com
---
### Upgrade BMC
---
#### Getting to the BMC
The BMC can be accessed via the front panel console port. Be default, the console port is set to the x86 processor’s console instead of the BMC’s
To switch to BMC console:
- At the x86 console Linux prompt, assuming GoES is running already, enter “sudo goes toggle”
- The console has switched to BMC at this point. Hit return to see the BMC CLI prompt, which should look something like:
192.168.101.211>
-To switch back to x86 console:
 At the BMC console prompt, enter “toggle”. The console has switched to x86 at this point. Hit return to see the Linux prompt.

#### General Upgrade Process
From BMC console CLI:
- Enter:
upgrade -v LATEST -s downloads.platinasystems.com
This will automatically go to http://downloads.platinasystems.com/LATEST/ to wget the zip file
platina-mk1-bmc.zip, install, and reboot the BMC to complete the upgrade. Reboot of BMC will not
affect the x86.

‘LATEST’ can be replaced by any version control string (e.g. 'v0.2').

'downloads.platinasystems.com' can be replaced by any server http or IP address reachable via
network that supports wget (e.g. 192.168.101.127).
Example: 'upgrade -v v0.2 -s 192.168.101.127' will go to 192.168.101.127/v0.2/ to wget the zip file
platina-mk1-bmc.zip.

By default if the -s option is left out, upgrade command will go to
http://downloads.platinasystems.com/ to look for the files.

To see a list of software revisions available for upgrade on http://downloads.platinasystems.com/,
enter 'upgrade -l':

---

### Manual upgrade
---
---
