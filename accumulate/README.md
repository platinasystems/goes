Package `accumulate` provides a wrapper to sum Reader and Writers.
Use it in a WriteTo like this,

```Go
func (t *TYPE) WriteTo(w) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, ...)
	...
	fmt.Fprint(acc, ...)
	return acc.N, acc.Err
}
```

An accumulator will skip subsequent read and writes on error.

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ../LICENSE
