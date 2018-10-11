// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
)

const gopkgdir = "main/go-package"

var gopkgpath string

const Tmpl = `// DO NOT EDIT! Instead, "go generate {{.Pkg.ImportPath}}".
package {{.Pkg.Name}}

//go:generate go run {{.Info.GoPkgDir}}/main.go {{.Info.GoRootDir}} {{.Info.RootUrl}}

var Package = map[string]string{
	"importpath": "{{.Info.RootUrl}}",
	"generated.by": "{{.Info.Generated.By}}",
	"generated.on": "{{.Info.Generated.On}}",
	"version": "{{.Info.Version}}",{{if .Info.Tag}}
	"tag": ` + "`" + `{{.Info.Tag}}` + "`" + `,{{end}}{{if .Info.Diff}}
	"diff": ` + "`" + `{{.Info.Diff}}` + "`" + `,{{end}}
	"license": ` + "`" + `{{.Info.License}}` + "`" + `,
	"patents": ` + "`" + `{{.Info.Patents}}` + "`" + `,
}
`

var Args = os.Args
var Exit = os.Exit
var Stderr io.Writer = os.Stderr

type Info struct {
	GoPkgDir, GoRootDir, RootUrl, Diff, Version, Tag, License, Patents string

	Generated struct {
		By, On string
	}
}

func main() {
	//	defer func() {
	//		if x := recover(); x != nil {
	//			fmt.Fprintln(Stderr, x)
	//			Exit(1)
	//		}
	//	}()
	useremail := "no.one@no.where"
	buf, err := exec.Command("git", "config", "--get",
		"user.email").Output()
	if err == nil && len(buf) > 0 {
		useremail = string(bytes.TrimSpace(buf))
	}
	generatedBy := useremail
	generatedOn := time.Now().UTC().String()
	gopkgpath = Args[1] + "/" + gopkgdir
	gopkg, err := build.ImportDir(gopkgpath, 0)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("gen").Parse(Tmpl)
	if err != nil {
		panic(err)
	}
	if len(Args) == 2 {
		Args = append(Args, ".")
	}
	path := Args[2]
	var assume_unchanged string
	info := new(Info)
	info.GoRootDir = Args[1]
	info.Generated.By = generatedBy
	info.Generated.On = generatedOn
	info.RootUrl = Args[3]
	pkg, err := build.ImportDir(path, 0)
	if err != nil {
		panic(err)
	}
	info.GoPkgDir, err = filepath.Rel(pkg.Dir, gopkg.Dir)
	if err != nil {
		panic(err)
	}
	if !build.IsLocalImport(info.GoPkgDir) {
		info.GoPkgDir = "./" + info.GoPkgDir
	}
	pkgfn := filepath.Join(pkg.Dir, "package.go")
	if _, err = os.Stat(pkgfn); err == nil {
		// overwrite existing package.go w/ repos version
		buf, err = exec.Command("git", "-C", pkg.Dir,
			"rev-parse", "HEAD").Output()
		if err != nil {
			panic(err)
		}
		info.Version = string(buf[:len(buf)-1])
		assume_unchanged = "--assume-unchanged"
	} else {
		info.Version = fmt.Sprint("FIXME with go generate ",
			pkg.ImportPath)
		assume_unchanged = "--no-assume-unchanged"
	}
	buf, err = exec.Command("git", "describe", "--tags",
		"--dirty='").Output()
	if err == nil && len(buf) > 0 {
		info.Tag = string(buf[:len(buf)-1])
	}
	buf, err = exec.Command("git", "diff", "--numstat").Output()
	if err == nil && len(buf) > 0 {
		info.Diff = string(buf[:len(buf)-1])
	}
	for _, x := range []struct {
		fn string
		p  *string
	}{
		{"LICENSE", &info.License},
		{"PATENTS", &info.Patents},
	} {
		r, e := os.Open(filepath.Join(pkg.Dir, x.fn))
		if e == nil {
			buf, err = ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
			*x.p = string(buf)
		}
	}
	pkgf, err := os.Create(pkgfn)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(pkgf, struct {
		Pkg  *build.Package
		Info *Info
	}{pkg, info})
	pkgf.Close()
	if err != nil {
		panic(err)
	}
	exec.Command("git", "-C", pkg.Dir, "update-index",
		assume_unchanged, pkgfn).Run()

}
