//go:build unix

package kill

import "syscall"

const KillDefaultSignal = syscall.SIGTERM
const KillDefaultSignalName = "term"

type Signal = syscall.Signal
