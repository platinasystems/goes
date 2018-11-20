// package kexec contains the kexec-related syscalls
// +build linux

package kexec

import (
	"fmt"
	"github.com/platinasystems/fdt"
	"github.com/platinasystems/go/internal/memmap"
	"io/ioutil"
	"os"
	"syscall"
)

func fileAddSegment(s []KexecSegment, f *os.File, a uintptr) ([]KexecSegment, error) {
	if f != nil {
		dat, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return SliceAddSegment(s, &dat, a), nil
	}
	return s, nil
}

func fileLoadFiles(k *os.File, i *os.File, cmdline string, flags uintptr) (err error) {
	m, err := memmap.FileToMap("/proc/iomem")
	if err != nil {
		return err
	}

	kCode, ok := m["Kernel code"]
	if !ok {
		return fmt.Errorf("Can't find kernel code in /proc/iomem")
	}
	kBase := kCode.Ranges[0].Start
	segments := make([]KexecSegment, 0, 3)

	segments, err = fileAddSegment(segments, k, kBase)
	if err != nil {
		return err
	}
	segments, err = fileAddSegment(segments, i, kBase+0x2000000)
	if err != nil {
		return err
	}
	t := fdt.DefaultTree()
	if t != nil {
		chosen := t.RootNode.Children["chosen"]
		if chosen == nil {
			chosen = &fdt.Node{Name: "chosen", Depth: 1}
			t.RootNode.Children["chosen"] = chosen
		}
		//need 64 bit support - parsing for #address-cells
		if i != nil {
			chosen.Properties["linux,initrd-start"] = t.PropUint32ToSlice(uint32(segments[1].Mem))
			chosen.Properties["linux,initrd-end"] = t.PropUint32ToSlice(uint32(segments[1].Mem) + uint32(segments[1].Bufsz) + 1)
		}
		chosen.Properties["bootargs"] = []byte(cmdline + "\x00")
		fdt := t.FlattenTreeToSlice()
		segments = SliceAddSegment(segments, &fdt, kBase+0x1000000)
	}
	err = SegmentLoad(uint64(kBase), &segments, 0)
	return err
}

func FileLoad(k *os.File, i *os.File, cmdline string, flags uintptr) (err error) {
	err = fileLoadSyscall(k, i, cmdline, flags)
	if err == syscall.ENOSYS {
		err = fileLoadFiles(k, i, cmdline, flags)
	}
	return err
}
