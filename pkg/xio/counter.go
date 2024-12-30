// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xio

// Use WriteCounter as an [io.Writer] to [io.TeeReader] or [io.MultiWriter].
//
// For example, to total bytes read by [bufio.Scanner] within an
// [io.Reader] interface.
//
//	func (t *T) ReadFrom(r io.Reader) (int64, error) {
//		var count xio.WriteCounter
//		scanner := bufio.NewScanner(io.TeeReader(r, &count))
//		for scanner.Scan() {
//			...
//		}
//		return count.Total(), nil
//	}
//
// Or to total bytes writen by a [io.WriterTo] interface.
//
//	func (t T) WriteTo(w io.Writer) (int64, error) {
//		var count xio.WriteCounter
//		mw := io.MultiWriter(w, &count)
//		fmt.Fprint(mw, ...)
//		...
//		return count.Total(), nil
//	}
type WriteCounter int64

func (wc WriteCounter) Total() int64 {
	return int64(wc)
}

func (wc *WriteCounter) Write(data []byte) (int, error) {
	n := len(data)
	*wc += WriteCounter(n)
	return n, nil
}
