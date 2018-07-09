// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grub

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/platinasystems/go/goes/cmd/grub/menuentry"
)

type webserver struct {
	w http.ResponseWriter
}

func (ws *webserver) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (ws *webserver) Write(p []byte) (n int, err error) {
	s := strings.Replace(html.EscapeString(string(p)), "\n", "<br>", -1)
	ret, err := ws.w.Write([]byte(s))
	return ret, err
}

func (c *Command) addHandler(parent string, i int, me menuentry.Entry) {
	path := fmt.Sprintf("%s/%d", parent, i)
	http.HandleFunc(fmt.Sprintf("/%s/", path), func(w http.ResponseWriter, r *http.Request) {
		Menuentry.Menus = Menuentry.Menus[:0]
		ws := &webserver{w: w}
		io.WriteString(w, `<html>`)
		err := me.RunFun(ws, ws, ws, false, false)
		if err != nil {
			fmt.Fprintf(w, `Menu exit status: %s
<br>`, html.EscapeString(err.Error()))
		} else {
			kexec := c.KexecCommand()
			if len(kexec) > 0 {
				s := html.EscapeString(strings.Join(kexec, " "))
				fmt.Printf("kexec command: %s\n", s)
				fmt.Fprintf(w, `Execute <a href="kexec">%s</a><br>`, s)
			}
		}
		for j, v := range Menuentry.Menus {
			fmt.Fprintf(w, `<a href="%d/">%s</a>
<br>
`, j, v.Name)
			c.addHandler(path, j, v)
		}
		io.WriteString(w, `</html>`)
	})
	http.HandleFunc(fmt.Sprintf("/%s/kexec", path), func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<html>`)
		kexec := c.KexecCommand()
		err := Goes.Main(kexec...)
		if err != nil {
			fmt.Fprintf(w, "Failed: %s<br>", html.EscapeString(err.Error()))
		} else {
			io.WriteString(w, "Success, so how do you see this?")
		}
		io.WriteString(w, `</html>`)
	})
}

func (c *Command) startHttpServer(path string) {
	m := Menuentry.Menus
	Menuentry.Menus = Menuentry.Menus[:0]
	for i, v := range m {
		c.addHandler(path, i, v)

	}

	http.HandleFunc(fmt.Sprintf("/%s/", path), func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<html><img src=http://www.platinasystems.com/wp-content/uploads/2016/10/PLA-Logo-Final-01-1-1-300x36.png><br>`)
		for i, v := range m {
			fmt.Fprintf(w, `<a href="%d/">%s</a>
<br>
`, i, v.Name)
		}
		io.WriteString(w, `</html>`)
	})
}

func (c *Command) ServeMenus() {
	srv := &http.Server{Addr: ":8080"}
	c.startHttpServer("boot")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
}
