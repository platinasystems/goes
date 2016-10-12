// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package url returns reader/writers for a given url.
package url

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func Open(path string) (io.ReadCloser, error) {
	var (
		u   *url.URL
		r   *http.Response
		err error
	)
	u, err = url.Parse(path)

	// It's not a URL its a file.
	if err != nil || (u.Scheme == "" && u.Host == "") {
		return os.Open(path)
	}

	// Handle file://... URLs
	if u.Scheme == "file" {
		return os.Open(u.Path)
	}

	// Get URL from server.
	r, err = http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s: %s", u, http.StatusText(r.StatusCode))
		return nil, err
	}
	return r.Body, err
}

func createFile(path string, isAppend bool) (io.WriteCloser, error) {
	flags := os.O_CREATE | os.O_WRONLY
	if isAppend {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(path, flags, 0666)
}

type httpWriter struct {
	buf      *bytes.Buffer
	url      string
	isAppend bool
}

func (w *httpWriter) Write(p []byte) (n int, err error) { return w.buf.Write(p) }

func (w *httpWriter) Close() error {
	method := "PUT"
	if w.isAppend {
		method = "APPEND"
	}
	req, err := http.NewRequest(method, w.url, w.buf)
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("%s: %s", w.url, msg)
	}

	return err
}

func create(path string, isAppend bool) (io.WriteCloser, error) {
	u, err := url.Parse(path)

	// It's not a URL its a file.
	if err != nil || (u.Scheme == "" && u.Host == "") {
		return createFile(path, isAppend)
	}

	// Handle file://... URLs
	if u.Scheme == "file" {
		return createFile(u.Path, isAppend)
	}

	return &httpWriter{buf: &bytes.Buffer{}, url: path, isAppend: isAppend}, err
}

func Create(path string) (io.WriteCloser, error) { return create(path, false) }
func Append(path string) (io.WriteCloser, error) { return create(path, true) }
