// Copyright © 2021-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import (
	"log"
	"os"
)

func ExampleUnmuteWrite() {
	l := log.New(os.Stdout, "", 0)
	Unmute(l).Write([]byte("example\n"))
	// Output: example
}

func ExampleUnmutePrint() {
	l := log.New(os.Stdout, "", 0)
	Unmute(l).Print("example\n")
	// Output: example
}

func ExampleUnmuteMutePrint() {
	l := log.New(os.Stdout, "", 0)
	Unmute(Mute(l)).Print("example\n")
	// Output: example
}

func ExampleUnmuteUnmutePrint() {
	l := log.New(os.Stdout, "", 0)
	Unmute(Unmute(l)).Print("example\n")
	// Output: example
}

func ExampleUnmuteMuteUnmutePrint() {
	l := log.New(os.Stdout, "", 0)
	Unmute(Mute(Unmute(l))).Print("example\n")
	// Output: example
}

func ExampleMuteWrite() {
	l := log.New(os.Stdout, "", 0)
	Mute(l).Write([]byte("example\n"))
	// Output:
}

func ExampleMutePrint() {
	l := log.New(os.Stdout, "", 0)
	Mute(l).Print("example\n")
	// Output:
}

func ExampleMuteUnmutePrint() {
	l := log.New(os.Stdout, "", 0)
	Mute(Unmute(l)).Print("example\n")
	// Output:
}

func ExampleMutedMutePrint() {
	l := log.New(os.Stdout, "", 0)
	Mute(Mute(l)).Print("example\n")
	// Output:
}

func ExampleMuteUnmuteMutePrint() {
	l := log.New(os.Stdout, "", 0)
	Mute(Unmute(Mute(l))).Print("example\n")
	// Output:
}

func ExampleIsMutedUnmute() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsMuted(Unmute(l)))
	// Output: false
}

func ExampleIsMutedUnmuteMute() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsMuted(Unmute(Mute(l))))
	// Output: false
}

func ExampleIsMutedMute() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsMuted(Mute(l)))
	// Output: true
}

func ExampleIsMutedMuteUnmute() {
	l := log.New(os.Stdout, "", 0)
	l.Print(IsMuted(Mute(Unmute(l))))
	// Output: true
}
