// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type chroot struct {
	setup    string
	teardown string
}

var bootstrap = chroot{
	setup: `set -x
mount -t proc none proc
mount -t devtmpfs none dev
`,
}

func (c *Command) doInChroot(root chroot, command string) (err error) {
	script := root.setup + command + "\n"
	if root.teardown != "" {
		script = script + root.teardown + "\n"
	}
	err = c.writeTemplateToFile("install.sh", script)
	if err != nil {
		return fmt.Errorf("Error writing script %s: %w", command, err)
	}

	err = c.g.Main("!", "-cd", "/", "-chroot", c.Target, "-m",
		"/bin/sh", "install.sh")
	if err != nil {
		return fmt.Errorf("Error running script %s: %w", command, err)
	}
	return nil
}

func (c *Command) doCommandsInChroot(root chroot, commands []string) (err error) {
	for _, command := range commands {
		if err := c.doInChroot(root, command); err != nil {
			return fmt.Errorf("Error executing %s: %w", command,
				err)
		}
	}
	return nil
}

func (c *Command) writeTemplateToFile(file string, script string) (err error) {
	t := template.Must(template.New("template").Parse(script))
	f, err := os.Create(filepath.Join(c.Target, file))
	if err != nil {
		return fmt.Errorf("Error creating %s: %w", file, err)
	}
	defer f.Close()
	err = t.Execute(f, c)
	if err != nil {
		return fmt.Errorf("Error executing template %s: %w", file, err)
	}
	return nil
}
