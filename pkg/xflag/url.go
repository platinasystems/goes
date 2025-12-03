// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import "net/url"

type URL url.URL

func (u *URL) URL() *url.URL {
	return (*url.URL)(u)
}

func (u *URL) Set(s string) error {
	p, err := url.Parse(s)
	if err == nil {
		*((*url.URL)(u)) = *p
	}
	return err
}

func (u *URL) String() string {
	return (*url.URL)(u).String()
}
