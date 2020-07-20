// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

import (
	"fmt"
)

var formatNoSwap = `label: gpt
device: /dev/sda
unit: sectors
sector-size: 512

/dev/sda1 : size=     1024000, type=C12A7328-F81F-11D2-BA4B-00A0C93EC93B, uuid={{ .UUIDEFI }}
/dev/sda2 :                    type=0FC63DAF-8483-4772-8E79-3D69D8477DE4
`
var fstab = `PARTUUID={{ .UUIDEFI }}	/boot/efi	vfat	umask=0077	0	1
UUID={{ .UUIDLinux }}	/	ext4	errors=remount-ro	0	1
`

func (c *Command) filesystemSetup() (err error) {
	err = c.writeTemplateToFile("sda.format", formatNoSwap)
	if err != nil {
		return fmt.Errorf("filesystemSetup: Error writing sda.format: %w", err)
	}

	err = c.writeTemplateToFile("fstab", fstab)
	if err != nil {
		return fmt.Errorf("filesystemSetup: Error writing fstab: %w", err)
	}

	err = c.doCommandsInChroot(bootstrap, []string{
		"sfdisk /dev/sda < sda.format",
		"mkfs.vfat /dev/sda1",
		"mkfs.ext4 -U {{ .UUIDLinux }} /dev/sda2 > /dev/null",
	})
	if err != nil {
		return fmt.Errorf("Error setting up filesystems: %w", err)
	}
	return nil
}
