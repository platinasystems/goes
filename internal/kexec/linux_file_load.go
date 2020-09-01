// package kexec contains the kexec-related syscalls
// +build linux

package kexec

import (
	"fmt"
	"github.com/platinasystems/fdt"
	"github.com/platinasystems/goes/internal/memmap"
	"io/ioutil"
	"os"
	"syscall"
)

func LoadSlices(kdat, idat []byte, cmdline string, flags uintptr) (err error) {
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

	segments = SliceAddSegment(segments, &kdat, kBase)

	segments = SliceAddSegment(segments, &idat, kBase+0x2000000)

	t := fdt.DefaultTree()
	if t != nil {
		chosen := t.RootNode.Children["chosen"]
		if chosen == nil {
			chosen = &fdt.Node{Name: "chosen", Depth: 1}
			t.RootNode.Children["chosen"] = chosen
		}
		//need 64 bit support - parsing for #address-cells
		if idat != nil {
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
		kdat, err := ioutil.ReadAll(k)
		if err != nil {
			return fmt.Errorf("Opening kernel %s: %w", k.Name(), err)
		}
		idat := []byte{}
		if i != nil {
			idat, err = ioutil.ReadAll(i)
			if err != nil {
				return fmt.Errorf("Opening initramfs %s: %w",
					i.Name(), err)
			}
		}
		return LoadSlices(kdat, idat, cmdline, flags)
	}
	return err
}
