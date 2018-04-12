// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootd

// BOOT STATES FIXME REWORK THIS
const (
	BootStateUnknown = iota
	BootStateMachineOff
	BootStateCoreboot
	BootStateCBLinux
	BootStateCBGoes
	BootStateRegistrationStart
	BootStateRegistrationDone
	BootStateScriptStart
	BootStateScriptRunning
	BootStateScriptDone
	BootStateBooting
	BootStateUp
)

func bootText(i int) string {
	var bootStates = []string{
		"Unknown",
		"Off",
		"Coreboot",
		"Coreboot-linux",
		"Coreboot-goes",
		"Register-start",
		"Registered",
		"Script-sent",
		"Script-run",
		"Script-done",
		"Booting-linux",
		"Up",
	}
	return bootStates[i]
}

// INSTALL STATES
const (
	InstallStateFactory = iota
	InstallStateInProgess
	InstallStateCompleted
	InstallStateFail
	InstallStateFactoryRestoreStart
	InstallStateFactoryRestoreDone
	InstallStateFactoryRestoreFail
)

func installText(i int) string {
	var installStates = []string{
		"Factory",
		"In-progress",
		"Completed",
		"Install-fail",
		"Restore-start",
		"Restore-done",
		"Restore-fail",
	}
	return installStates[i]
}

// DISTROS
const (
	Debian = iota
)

func distroText(i int) string {
	var distroTypes = []string{
		"Debian",
	}
	return distroTypes[i]
}

// SCRIPTS
const (
	ScriptBootLatest = iota
	ScriptBootKnownGood
	ScriptInstallDebian
)

func scriptText(i int) string {
	var scripts = []string{
		"Boot-latest",
		"Boot-known-good",
		"Debian-install",
	}
	return scripts[i]
}

//msgtype, mac, ip, slice of sda2, myip?
// CLIENT SIDE MESSAGE REQUESTS
const (
	BootRequestRegister = iota
	BootRequestDumpVars
	BootRequestDashboard
	BootRequestKernelNotFound
	BootRequestRebootLoop
)
const (
	Register  = "register"
	DumpVars  = "dumpvars"
	Dashboard = "dashboard"
)

type BootRequest struct {
	Request int
}
type BootReply struct {
	Reply   int
	Binary  []byte
	Payload []byte
}

//msg type,reply#,error, json(name, scripttype, script,binary)
// SERVER SIDE MESSAGE REPLIES
const (
	BootReplyNormal = iota
	BootReplyRunGoesScript
	BootReplyExecUsermode
	BootReplyExecKernel
	BootReplyReflashAndReboot
)
