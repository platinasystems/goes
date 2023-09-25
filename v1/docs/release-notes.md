# Platina Systems Corporation
### RELEASE NOTES
##### Future releases will be named vX.Y.Z
---
### Release: v1.1.0 - Enhanced Functionality in Goes
#### New Features
  - MK1 X86 - Upgraded Link configuration, Enhanced Stability.

#### Description
  - Link Configuration
    - __ethtool__ is used for link configuration
    - Interface names for front panel ports are called '__xeth?__', for example, xeth1, xeth2...xeth32.
    - When ports are broken out into seprate lanes, Sub interfaces are called '__xethX-Y__', for example xeth1-1, xeth1-4,...,xeth32-1.. xeth-32-4
    - Supported Link Speeds
      - 100G - Interface names are xethX, where X = 1 .. 32
      - 50G  - Interface names are xethX-1,xethX-3, where X = 1 .. 32 
      - 40G  - Interface names are xethX, where X = 1 .. 32
      - 25G  - Interface names are xethX-Y, where X = 1 .. 32, Y = 1 .. 4
      - 10G  - Interface names are xethX-Y, where X = 1 .. 32, Y = 1 .. 4 
      - 1G   - Interface names are xethX-Y, where X = 1 .. 32, Y = 1 .. 4
    - Each individual port can be configured independently.
    - Please refer to [Platina Config Guide](https://github.com/platinasystems/go/docs/Platina_Config_Guide_v0.1.md) for breakout configuration and persistence. 
  - Routing and Protocol Support
    - Open Source routing stacks are supported. E.g. FRR, Quagga, GoBGP, Bird.
  - Network Slicing, VRF support
    - Network slicing and VRF are supported with linux namespaces.
 #### Compatible Versions
   - MK1 X86 (Goes) - v1.1.0
   - Coreboot -
   - Kernel -
---
### Release:  __20170910__ - BMC only, upgrade improvements
#### New Features
  - MK1 X86 - __No changes__
  - MK1 BMC - __Upgrade improvements.  Version support.  Verify.__

#### Description
```Upgrade improvements.  Version support.  Verify.```
 
### Version List
| MK1 _x86_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ | ------ | ------ |
| Goes | v0.2 | dcb42af | No changes from previous release |
| FE1 | v0.2 | cdb4a93 | No changes from previous release |
| FE1 Firmware | v0.2 | 60f3914 | No changes from previous release |
| Linux | v4.11 | bd1317f | Debian |
| Coreboot | v0.2 | 923fea6 | No changes from previous release |

| MK1 _BMC_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ |------| ------ |
| Initrd/Goes | v0.4 | ea9e25e | Upgrade improvements.  Version support.  Verify. |
| Linux | v4.11 | 1047c60 | Debian - I2C leakage fix.  Access of QSPI1 in goes. |
| DTB - devtree | v4.11 | 1047c60 | I2C leakage fix.  Access of QSPI1 in goes. |
| u-boot | v0.2 | 3825c3d | |
| u-boot env | v0.4 | 04f5e6c | Upgrade improvements.  Version support.  Verify. |
---
### Release:  __V0.3__ - BMC only, upgrade no longer requires MMC/SDcard
#### New Features
  - MK1 X86 - __No changes__
  - MK1 BMC - __Upgrade no longer requires MMC/SDcard__

#### Description
```BMC only, upgrade no longer requires MMC/SDCard```
 
### Version List
| MK1 _x86_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ | ------ | ------ |
| Goes | v0.2 | dcb42af | No changes from previous release |
| FE1 | v0.2 | cdb4a93 | No changes from previous release |
| FE1 Firmware | v0.2 | 60f3914 | No changes from previous release |
| Linux | v4.11 | bd1317f | Debian |
| Coreboot | v0.2 | 923fea6 | No changes from previous release |

| MK1 _BMC_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ |------| ------ |
| Initrd/Goes | v0.3 | e98d25f | Upgrade command works w/o MMC/SDcard |
| Linux | v4.11 | ebb5e88 | Debian - w/patch for qspi 4-byte addressing |
| DTB - devtree | v4.11 | ebb5e88 | |
| u-boot | v0.2 | 3825c3d | |
| u-boot env | --- |xxxxxxx | |
---

---
### Release:  __V0.2__ - Pre namespace collapse
#### New Features
  - MK1 X86 - __Initial Release - Pre namespace collapse__
  - MK1 BMC - __Initial Release__

#### Description
```This is the initial release```

### Version List
| MK1 _x86_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ |------| ------ |
| Goes | v0.2 | dcb42af | Initial Release |
| FE1 | v0.2 | cdb4a93 | Initial Release |
| FE1 Firmware | v0.2 | 60f3914 | Initial Release |
| Linux | v4.11 | bd1317f | Debian |
| Coreboot | v0.2 | 923fea6 | Initial Release |

| MK1 _BMC_ | Version Tag | Github SHA-1 | Changes |
| ------ | ------ |------| ------ |
| Initrd/Goes | v0.2 | dcb42af | Initial Release |
| Linux | v4.11 | bd1317f | Debian |
| DTB - devtree | v4.11 | bd1317f | Initial Release |
| u-boot | v0.2 | 3825c3d | Initial Release |
| u-boot env | --- | xxxxxxx | Initial Release |
---
### Installation
---
### Copyright
---
### License
