// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cavaliercoder/go-cpio"

	"github.com/platinasystems/url"

	"github.com/ulikunitz/xz"
	//	"golang.org/x/sys/unix"
)

func (c *Command) readArchive() (err error) {
	txz, err := url.Open(c.Archive)
	if err != nil {
		return fmt.Errorf("Error opening %s: %w", c.Archive, err)
	}
	defer txz.Close()

	tr, err := xz.NewReader(txz)
	if err != nil {
		return fmt.Errorf("Error in xz.NewReader: %w", err)
	}

	archive := cpio.NewReader(tr)

	for {
		file, err := archive.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("Error reading archive: %v", err)
		}
		path := filepath.Join(c.Target, file.Name)
		if (file.FileInfo().Mode() & os.ModeSymlink) != 0 {
			err := os.Symlink(file.Linkname, path)
			if err != nil {
				fmt.Printf("Error creating symlink %s->%s:%s\n",
					path, file.Linkname, err)
			}
			continue
		}
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, os.FileMode(file.Mode))
			if err != nil {
				fmt.Printf("Error creating directory %s: %s\n",
					path, err)
			}
			continue
		}
		if (file.FileInfo().Mode() & os.ModeDevice) != 0 {
			//			err := syscall.Mknod(path, uint32(file.FileInfo().Mode()),
			//	int(unix.Mkdev(uint32(file.Devmajor), uint32(file.Devminor))))
			//if err != nil {
			//	fmt.Printf("Error creating special file %s: %s",
			//		path, err)
			//}
			continue
		}
		if !file.FileInfo().Mode().IsRegular() {
			fmt.Printf("Can't create %x file %s\n",
				file.FileInfo().Mode(), path)
			continue
		}
		err = func(path string, archive io.Reader) (err error) {
			t, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				os.FileMode(file.Mode))
			if err != nil {
				return fmt.Errorf("Error creating %s: %w",
					path, err)
			}
			defer t.Close()
			if _, err := io.Copy(t, archive); err != nil {
				return err
			}
			return nil
		}(path, archive)
		if err != nil {
			return err
		}
	}
	return nil
}
