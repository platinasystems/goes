//go:build !darwin

package sysctl

const (
	CTL_SYSCTL = iota
	CTL_KERN
	CTL_VM
	CTL_VFS
	CTL_NET
	CTL_DEBUG
	CTL_HW
	CTL_MACHDEP
	CTL_USER
	CTL_P1003_1B
)
