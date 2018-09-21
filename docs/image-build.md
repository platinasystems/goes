# Platina Systems Corporation
### Building Images
---
## Building the BMC for internal use

1. cd system-build
2. scripts/make-mk1-bmc
3. output will be "platina-mk1-bmc.zip"
4. upgrade the bmc using the upgrade command in goes

---
## Building the BMC for release

1. make sure tags are in place
2. cd system-build
3. scripts/make-mk1-bmcrel
4. output will be "platina-mk1-bmc.zip"
5. upgrade the bmc using the upgrade command in goes

---
## Building the kernel .deb file

1. cd system-build
2. rm -rf linux
3. make -B linux/linux-image-4.13.0-platina-mk1_4.13-135-gf5d1b0298af7_amd64.deb
   where 135-gf5d1b0298af7 is the value from git describe
4. output will be "linux-image-4.13.0-platina-mk1_4.13-135-gf5d1b0298af7_amd64.deb"
   or similar

---
## Building coreboot (goes-boot)

1. cd system-build
2. scripts/make-coreboot
3. output will be "coreboot-platina-mk1.rom"

---
## Building goes-platina-mk1

1. cd go
2. go generate
3. either cd main/goes-platina-mk1, go build
4. or go run ./main/goes-build/main.go goes-platina-mk1

--- Building goes-platina-mk1-installer
1. cd go
2. go generate
3. go run ./main/goes-build/main.go goes-platina-mk1-installer

---
---
---
## Putting the images in downloads/TEST, downloads/LATEST, etc.

scp files to /var/www/html/downloads/TEST:

     platina-mk1-bmc.zip
     coreboot-platina-mk1.rom
     linux-image-4.13.0-platina-mk1_4.13-135-gf5d1b0298af7_amd64.deb
     goes-platina-mk1-installer
     goes-platina-mk1

rename linux-image-4.13.0-platina-mk1_4.13-135-gf5d1b0298af7_amd64.deb to linux-image-platina-mk1-4.13.0.deb

edit linux-image-platina-mk1, update the two lines:

     v4.13-139-gace1518-platina-mk1-amd64
     linux-image-platina-mk1-4.13.0.deb

the downloads/TEST directory will end up have six (6) files in it:

     coreboot-platina-mk1.rom
     goes-platina-mk1
     goes-platina-mk1-installer
     linux-image-platina-mk1
     linux-image-platina-mk1-4.13.0.deb
     platina-mk1-bmc.zip
