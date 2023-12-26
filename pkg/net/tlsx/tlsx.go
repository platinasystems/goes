// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

var Daemons = map[string]any{
	"exchange": Exchange,
	"tap":      TunTap,
	"tun":      TunTap,
}

var Root = map[string]any{
	"approve":   IPC,
	"deny":      IPC,
	"exec":      Rexec,
	"subscribe": Subscribe,
}

var Show = map[string]any{
	"cert": map[string]any{
		"pem":  Selfie.MarshalPEM,
		"text": Selfie.MarshalText,
	},
	"filenames":     Filenames,
	"registry":      IPC,
	"subscribers":   SubscribersNames,
	"subscriptions": SubscriptionsNames,
	"tenants":       IPC,
}
