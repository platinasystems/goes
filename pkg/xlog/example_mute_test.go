// Copyright © 2021-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xlog

import (
	"log"
	"os"
)

func ExampleUnmute() {
	l := log.New(os.Stdout, "", 0)
	Unmute(l).Write([]byte("write slice\n"))
	Unmute(l).Println("print string")
	Unmute(Mute(l)).Println("UnmuteMutePrint")
	Unmute(Unmute(l)).Println("UnmuteUnmutePrint")
	Unmute(Mute(Unmute(l))).Println("UnmuteMuteUnmutePrint")
	// Output:
	// write slice
	// print string
	// UnmuteMutePrint
	// UnmuteUnmutePrint
	// UnmuteMuteUnmutePrint
}

func ExampleMute() {
	l := log.New(os.Stdout, "", 0)
	Mute(l).Write([]byte("write slice"))
	Mute(l).Println("print string")
	Mute(Unmute(l)).Println("MuteUnmutePrint")
	Mute(Mute(l)).Println("MutedMutePrint")
	Mute(Unmute(Mute(l))).Println("MuteUnmuteMutePrint")
	// Output:
}

func ExampleIsMuted() {
	l := log.New(os.Stdout, "", 0)
	l.Println("Unmute", IsMuted(Unmute(l)))
	l.Println("Unmute/Mute", IsMuted(Unmute(Mute(l))))
	l.Println("Mute", IsMuted(Mute(l)))
	l.Println("Mute/Unmute", IsMuted(Mute(Unmute(l))))
	// Output:
	// Unmute false
	// Unmute/Mute false
	// Mute true
	// Mute/Unmute true
}

func ExampleToggleMute() {
	mutable := ToggleMute(log.New(os.Stdout, "", 0))
	mutable.Println("unmuted")
	mutable = ToggleMute(mutable)
	mutable.Println("muted")
	mutable = ToggleMute(mutable)
	mutable.Println("unmuted")
	// Output:
	// unmuted
	// unmuted
}
