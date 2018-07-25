/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Netns uint64

var nameByInode = map[Netns]string{
	1: "default",
}

func (ns Netns) String() string {
	name, found := nameByInode[ns]
	if found {
		return name
	}
	filepath.Walk("/run",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if len(name) > 0 {
				return filepath.SkipDir
			}
			stat := info.Sys().(*syscall.Stat_t)
			if stat.Ino == uint64(ns) {
				name = info.Name()
				return filepath.SkipDir
			}
			return nil
		})
	if len(name) > 0 {
		nameByInode[ns] = name
		return name
	}
	return fmt.Sprintf("%#x", uint64(ns))
}
