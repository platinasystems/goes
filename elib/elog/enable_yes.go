// +build elog

package elog

func Enabled() bool { return DefaultBuffer.Enabled() }
