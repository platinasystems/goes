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

type Command struct{}

func (Command) String() string { return "wget" }

func (Command) Usage() string { return "wget URL..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "a non-interactive network downloader",
	}
}

func (Command) Main(args ...string) error {
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
