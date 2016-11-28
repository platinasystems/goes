// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"os"
	"os/exec"
	"text/template"
)

func main() {
	tmpl, err := template.New("gen").Parse(`
// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DO NOT EDIT! Instead, "go generate github.com/platinasystems/go/version".
package {{.Package}}

//go:generate go run ./go-version/main.go

const Version = "{{.Version}}"
`[1:])
	if err != nil {
		panic(err)
	}

	buf, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		panic(err)
	}
	generate(tmpl, "version.go", "version", string(buf[:len(buf)-1]))
}

func generate(tmpl *template.Template, fn, pkg, version string) {
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		Package string
		Version string
	}{pkg, version})
	if err != nil {
		panic(err)
	}

	exec.Command("git", "update-index", "--assume-unchanged", fn).Run()
}
