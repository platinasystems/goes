// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package scp

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "scp" }

func (Command) Usage() string { return "scp [-t] [-f] DIR" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "securely copy a file",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	scp implemenets the server side of the SCP protocol.

OPTIONS
	-f	Specifies source mode
	-t	Specifies sink mode`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-d", "-f", "-p", "-r", "-t")
	i := 0
	if flag.ByName["-f"] {
		i++
	}
	if flag.ByName["-t"] {
		i++
	}
	if i != 1 {
		return fmt.Errorf("scp requires one of -f or -t")
	}
	if len(args) == 0 {
		return fmt.Errorf("scp requires a directory or filename")
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected %v", args)
	}
	isDir := false
	exists := false
	p := args[0]
	s, err := os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if flag.ByName["-d"] {
			return err
		}
		if strings.HasSuffix(p, string(os.PathSeparator)) {
			return fmt.Errorf("%s: is directory", p)
		}
		if flag.ByName["-r"] {
			isDir = true
		}
	} else {
		exists = true
		if s.IsDir() {
			isDir = true
		} else if !s.Mode().IsRegular() {
			return fmt.Errorf("%s: is not a regular file", p)
		}
	}
	if flag.ByName["-f"] {
		return sourceMode(flag, bufio.NewReader(os.Stdin), p)
	}
	return sinkMode(flag, exists, isDir, bufio.NewReader(os.Stdin), p)
}

func sourceMode(flag *flags.Flags, r *bufio.Reader, p string) error {
	s, err := os.Stat(p)
	if err != nil {
		return err
	}
	if s.IsDir() {
		dir, err := ioutil.ReadDir(p)
		if err != nil {
			return err
		}
		fmt.Printf("D%#o %d %s\n", s.Mode()&os.ModePerm, 0, s.Name())
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b != 0 {
			return fmt.Errorf("Unexpected response %#x", b)
		}
		for _, f := range dir {
			err := sourceMode(flag, r, filepath.Join(p, f.Name()))
			if err != nil {
				return err
			}
		}
		fmt.Printf("E\n")
		b, err = r.ReadByte()
		if err != nil {
			return err
		}
		if b != 0 {
			return fmt.Errorf("Unexpected response %#x", b)
		}
		return nil
	}
	if flag.ByName["-p"] {
		stat := s.Sys().(*syscall.Stat_t)
		fmt.Printf("T%d 0 %d 0\n", stat.Mtim.Sec, stat.Atim.Sec)
		b, err := r.ReadByte()
		if err != nil {
			return err
			if b != 0 {
				return fmt.Errorf("Unexpected response %#x", b)
			}
		}
	}
	fmt.Printf("C%#o %d %s\n", s.Mode()&os.ModePerm, s.Size(), s.Name())
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b != 0 {
		return fmt.Errorf("Unexpected response %#x", b)
	}
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	fmt.Printf("%s\x00", buf)

	b, err = r.ReadByte()
	if err != nil {
		return err
	}
	if b != 0 {
		return fmt.Errorf("Unexpected response %#x", b)
	}
	return nil
}

func sinkMode(flag *flags.Flags, exists, isDir bool, r *bufio.Reader, p string) error {
	dirStack := make([]string, 0)
	var atime, mtime time.Time
	fmt.Printf("\x00")
	s := bufio.NewScanner(r)
	for s.Scan() {
		c := s.Text()
		switch c[0] {
		case 'C':
			// Cmmmm <length> <filename>
			f := strings.SplitN(c[1:], " ", 3)
			m, err := strconv.ParseUint(f[0], 0, 32)
			if err != nil {
				return err
			}
			l, err := strconv.ParseUint(f[1], 0, 64)
			if err != nil {
				return err
			}
			fmt.Printf("\x00")
			b := make([]byte, l+1)
			_, err = io.ReadFull(r, b)
			if err != nil {
				return err
			}
			if b[l] != 0 {
				return fmt.Errorf("Unexpected trailing byte %02x\n", b[l])
			}
			b = b[:l]
			fn := ""
			if isDir {
				fn = filepath.Join(p, f[2])
			} else {
				fn = p
			}
			err = ioutil.WriteFile(fn, b, os.FileMode(m))
			if err != nil {
				log.Print("error writing ", fn, ": ", err)
				return err
			}
			if flag.ByName["-p"] {
				err = os.Chtimes(fn, atime, mtime)
				if err != nil {
					log.Print("error setting times ", fn, ": ", err)
					return err
				}
			}
			fmt.Printf("\x00")
			log.Print("successfully wrote ", fn)

		case 'D':
			// Dmmmm <length-ignore> <dirname>
			f := strings.SplitN(c[1:], " ", 3)
			m, err := strconv.ParseUint(f[0], 0, 32)
			if err != nil {
				return err
			}
			_, err = strconv.ParseUint(f[1], 0, 64)
			if err != nil {
				return err
			}
			fmt.Printf("\x00")
			dirStack = append(dirStack, p)
			if exists || len(dirStack) > 1 {
				p = filepath.Join(p, f[2])
			}
			log.Print("D command creating ", p)
			err = os.Mkdir(p, os.FileMode(m))
			if err != nil {
				return fmt.Errorf("mkdir %s: %s", p, err)
			}
		case 'E':
			if len(dirStack) == 0 {
				return fmt.Errorf("E command without matching D")
			}
			p = dirStack[len(dirStack)-1]
			dirStack = dirStack[:len(dirStack)-1]
			fmt.Printf("\x00")

		case 'T':
			//T<mtime> <mtime-usec> <atime> <atime-usec>
			if !flag.ByName["-p"] {
				return fmt.Errorf("T command without -p")
			}
			f := strings.SplitN(c[1:], " ", 4)
			mtimeSec, err := strconv.ParseUint(f[0], 0, 64)
			if err != nil {
				return err
			}
			mtimeUsec, err := strconv.ParseUint(f[1], 0, 64)
			if err != nil {
				return err
			}
			atimeSec, err := strconv.ParseUint(f[2], 0, 64)
			if err != nil {
				return err
			}
			atimeUsec, err := strconv.ParseUint(f[3], 0, 64)
			if err != nil {
				return err
			}
			mtime = time.Unix(int64(mtimeSec), int64(mtimeUsec)*1000)
			atime = time.Unix(int64(atimeSec), int64(atimeUsec)*1000)
			fmt.Printf("\x00")

		default:
			return fmt.Errorf("Unknown command %s", c)
		}
	}
	return nil
}
