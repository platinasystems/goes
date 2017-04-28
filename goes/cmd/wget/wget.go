// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package wget

import (
	"fmt"

	"github.com/cavaliercoder/grab"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "wget"
	Apropos = "a non-interactive network downloader"
	Usage   = "wget URL..."
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	// validate command args
	if len(args) < 1 {
		return fmt.Errorf("URL: missing")
	}

	reqs := make([]*grab.Request, 0)
	for _, url := range args {
		req, err := grab.NewRequest(url)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}

	successes, err := url.FetchReqs(0, reqs)
	if successes == 0 && err != nil {
		return err
	}

	fmt.Printf("%d files successfully downloaded.\n", successes)
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
