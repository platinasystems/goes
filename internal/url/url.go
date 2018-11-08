// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
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
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/platinasystems/go/internal/tftp"
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

	// Handle tftp://... URLs
	if u.Scheme == "tftp" {
		rc, err := tftp.GetFileRC(u.Host + ":" + u.Path)
		if err != nil {
			return nil, err
		}
		return rc, nil
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

func FetchReqs(parallel int, reqs []*grab.Request) (successes int, err error) {
	successes = 0
	firstErr := error(nil)

	// create a custom client
	client := grab.NewClient()
	client.UserAgent = "Platina Go-ES"

	// start file downloads at the requested batch size
	fmt.Printf("Downloading %d files...\n", len(reqs))
	respch := client.DoBatch(parallel, reqs...)

	// start a ticker to update progress every 200ms
	t := time.NewTicker(200 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// clear lines
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					//Link request to response
					resp.Request.Tag = resp
					// print final result
					if resp.Error != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Error)
						if firstErr == nil {
							firstErr = resp.Error
						}
					} else {
						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
						successes++
					}

					// mark completed
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}

	t.Stop()

	return successes, firstErr
}
