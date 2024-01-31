// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

type Subs struct {
	sync.RWMutex
	name string
	head *X509
}

var Subscribers, Subscriptions Subs

var InitSubscribers = sync.OnceValue(func() error {
	Subscribers.name = *OptionalSubscribersFileName
	return Subscribers.Init()
})

var InitSubscriptions = sync.OnceValue(func() error {
	Subscriptions.name = *OptionalSubscriptionsFileName
	return Subscriptions.Init()
})

var DefaultSubscribersFileName = sync.OnceValue(func() string {
	return filepath.Join(xdg.StateHome(), program.Base(),
		"tlsx", "subscribers.pem")
})

var DefaultSubscriptionsFileName = sync.OnceValue(func() string {
	return filepath.Join(xdg.StateHome(), program.Base(),
		"tlsx", "subscriptions.pem")
})

var OptionalSubscribersFileName, OptionalSubscriptionsFileName *string

func ShowSubs(ctx context.Context, args []string) error {
	var x X509
	var fn string
	branch := goes.ContextBranch(ctx)
	switch last := branch[len(branch)-1]; last {
	case "subscribers":
		fn = *OptionalSubscribersFileName
	case "subscriptions":
		fn = *OptionalSubscriptionsFileName
	default:
		return fmt.Errorf("%q: %w", last, ErrInvalid)
	}
	if goes.ContextComplete(ctx) {
		return complete.Last(args, "*.pem")
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["filename"] = fn
		return goes.Usage(ctx, `
usage: {{branch .}}
Print x509 certifcates within {{filename}}.`)
	}
	r, err := goes.ContextInput(ctx, fn)
	if err != nil {
		return err
	}
	if _, err = x.ReadFrom(r); err != nil {
		return err
	}
	w := goes.ContextStdout(ctx)
	for p := &x; p != nil; p = p.Next {
		fmt.Fprint(w, p)
	}
	return nil
}

func (subs *Subs) Init() error {
	data, err := os.ReadFile(subs.name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = subs.head.UnmarshalText(data)
	if err == nil && subs.head.Block == nil {
		err = ErrInvalid
	}
	return err
}

func (subs *Subs) Append(sub *X509) error {
	subs.Lock()
	defer subs.Unlock()
	if subs.head == nil {
		subs.head = sub
	} else {
		subs.head.Append(sub)
	}
	const fflags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	fmode := fs.FileMode(0600)
	if fi, err := os.Stat(subs.name); err == nil {
		fmode = fi.Mode()
	}
	f, err := os.OpenFile(subs.name, fflags, fmode)
	if err != nil {
		return err
	}
	defer f.Close()
	subs.head.Range(func(xx *X509) bool {
		return pem.Encode(f, xx.Block) == nil
	})
	return nil
}

func (subs *Subs) Index(nameOrSKI string) int {
	subs.RLock()
	defer subs.RUnlock()
	return subs.head.Index(nameOrSKI)
}

func (subs *Subs) Match(nameOrSKI string) *X509 {
	subs.RLock()
	defer subs.RUnlock()
	return subs.head.WhoIs(nameOrSKI)
}

func (subs *Subs) Names() []string {
	subs.RLock()
	defer subs.RUnlock()
	return subs.head.Names()
}

func (subs *Subs) Range(f func(*X509) bool) {
	subs.RLock()
	defer subs.RUnlock()
	subs.head.Range(f)
}

func (subs *Subs) SKIs() []string {
	subs.RLock()
	defer subs.RUnlock()
	return subs.head.SKIs()
}
