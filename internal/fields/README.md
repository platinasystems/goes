Slice a string into args while combining single, double, or backslash
escaped spaced arguments, e.g.:

```go
	fields.New(`echo hello\ beautiful\ world`) == []string{
		"echo",
		"hello beautiful world",
	}
	fields.New(`echo "hello 'beautiful world'"`) == []string{
		"echo",
		"hello 'beautiful world'",
	}
	fields.New(`echo 'hello \"beautiful world\"'`) == []string{
		"echo",
		`hello \"beautiful world\"`,
	}
```

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ../LICENSE
