// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix && !darwin && !ios

package style

import "log/syslog"

var priorities = map[Level]syslog.Priority{
	Emergency: syslog.LOG_EMERG,
	Alert:     syslog.LOG_ALERT,
	Critical:  syslog.LOG_CRIT,
	Errata:    syslog.LOG_ERR,
	Warning:   syslog.LOG_WARNING,
	Notice:    syslog.LOG_NOTICE,
	Info:      syslog.LOG_INFO,
	Debug:     syslog.LOG_DEBUG,
}

// Log to syslog instead of Std{out|err}.
func (style Style) System() {
	pri := syslog.LOG_DAEMON | priorities[style.Level]
	if sl, err := syslog.New(pri, Base); err == nil {
		style.SetOutput(sl)
		style.SetPrefix("")
	}
}
