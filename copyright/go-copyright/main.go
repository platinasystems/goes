// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

const backquote = "`"
const copyright = `
// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DO NOT EDIT! Instead, "go generate github.com/platinasystems/go/copyright".
package copyright

//go:generate go run ./go-copyright/main.go

const License = {{.Backquote}}{{.License}}{{.Backquote}}

const Patents = {{.Backquote}}{{.Patents}}{{.Backquote}}
`

func main() {
	license, err := ioutil.ReadFile("../LICENSE")
	if err != nil {
		panic(err)
	}
	patents, err := ioutil.ReadFile("../PATENTS")
	if err != nil {
		panic(err)
	}
	f, err := os.Create("copyright.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	template.Must(template.New("gen").Parse(copyright[1:])).Execute(f,
		struct {
			Backquote, License, Patents string
		}{
			backquote, string(license), string(patents),
		})
	exec.Command("git", "update-index", "--assume-unchanged",
		"copyright.go").Run()
}
