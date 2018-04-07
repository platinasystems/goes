// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootd

// BOOT STATES
const (
	BootStateUnknown = iota
	BootStateMachineOff
	BootStateCoreboot
	BootStateCBLinux
	BootStateCBGoes
	BootStateRegistrationStart
	BootStateRegistrationDone
	BootStateScriptStart
	BootStateScriptExecuting
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
		"Script-exec",
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

// INSTALL TYPES
const (
	Debian = iota
)

func installTypeText(i int) string {
	var installTypes = []string{
		"Debian",
	}
	return installTypes[i]
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

// CLIENT MESSAGES
const (
	BootRequestRegister = iota
	BootRequestDumpVars
	BootRequestDashboard
	BootRequestKernelNotFound
	BootRequestRebootLoop
)

//FIXME DEFINE THIS
type BootRequest struct {
	Request int
}

// SERVER REPLY MESSAGES
const (
	BootReplyNormal = iota
	BootReplyRunGoesScript
	BootReplyExecUsermode
	BootReplyExecKernel
	BootReplyReflashAndReboot
)

//FIXME DEFINE THIS
type BootReply struct {
	Reply   int
	Binary  []byte
	Payload []byte
}
