// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

var setupChroot = chroot{
	setup: bootstrap.setup + `mkdir -p /debian
mount /dev/{{ .InstallDev }}2 /debian
`,
}

var debianChroot = chroot{
	setup: bootstrap.setup + `mount /dev/{{ .InstallDev }}2 /debian
chroot /debian /bin/sh << EOF
set -x
mount -t proc none proc
mount -t devtmpfs none dev
[ -L /dev/fd ] || ln -s /proc/self/fd /dev/fd
mount -t devpts none /dev/pts
mount -t sysfs none sys
mkdir -p /debian
mount /dev/{{ .InstallDev }}2 /debian
export PATH
`,
	teardown: "EOF",
}

func (c *Command) debianInstall() (err error) {
	for _, cmd := range []struct {
		root chroot
		cmds []string
	}{
		{setupChroot, []string{
			"mkdir -p /debian/etc",
			"cp etc/resolv.conf /debian/etc",
			"{{ .DebootstrapProgram }} --arch amd64 {{ .DebootstrapOptions }}{{ .DebianDistro }} /debian {{ .DebianDownload }}",
			"cp fstab /debian/etc/fstab",
			"mkdir -p /debian/etc/network/interfaces.d",
			"cp {{ .MgmtEth }} /debian/etc/network/interfaces.d",
		},
		},

		{debianChroot, []string{
			"apt-get update",
			"apt-get -y install grub-efi-amd64 apt-transport-https dirmngr initramfs-tools openssh-server sudo ca-certificates ifupdown resolvconf wget",
			"wget {{ .PlatinaDownload }}/platina-apt-source_{{ .PlatinaDistro }}.deb {{ .PlatinaDownload }}/platina-archive-keyring_{{ .PlatinaDistro }}.deb",
			"dpkg -i platina-apt-source_{{ .PlatinaDistro }}.deb platina-archive-keyring_{{ .PlatinaDistro }}.deb",
			"apt-get update",
			"apt-get -y install {{ .PlatinaRelease }}",
			"update-grub",
			`adduser --gecos "System Administrator" --disabled-password {{ .AdminUser }}`,
			"adduser {{ .AdminUser }} sudo",
			"echo {{ .AdminUser }}:{{ .AdminPass }}|chpasswd",
			"echo {{ .Hostname }}>/etc/hostname",
			`{{if .DNSAddr }} sed -i -e "s/^#DNS=$/DNS={{ .DNSAddr }}/" /etc/systemd/resolved.conf{{end}}`,
			`sed -i -e "s/^source-directory \/etc\/network\/interfaces.d$/source \/etc\/network\/interfaces.d\/*/" /etc/network/interfaces`,
			"systemctl enable systemd-resolved",
		},
		},
	} {
		if err := c.doCommandsInChroot(cmd.root, cmd.cmds); err != nil {
			return err
		}
	}
	return nil
}
