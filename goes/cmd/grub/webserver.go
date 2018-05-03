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

func (c *Command) startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":8080"}
	m := Menuentry.Menus
	for i, v := range m {
		me := v
		http.HandleFunc(fmt.Sprintf("/boot/%d/", i), func(w http.ResponseWriter, r *http.Request) {
			ws := &webserver{w: w}
			io.WriteString(w, `<html>`)
			err := me.RunFun(ws, ws, ws, false, false)
			if err != nil {
				fmt.Fprintf(w, `Menu exit status: %s
<br>`, html.EscapeString(err.Error()))
			} else {
				kexec := c.KexecCommand()
				s := html.EscapeString(strings.Join(kexec, " "))
				fmt.Printf("kexec command: %s\n", s)
				fmt.Fprintf(w, `Execute <a href="kexec">%s</a><br>`, s)
			}
			io.WriteString(w, `</html>`)
		})
		http.HandleFunc(fmt.Sprintf("/boot/%d/kexec", i), func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<html><img src=http://www.platinasystems.com/wp-content/uploads/2016/10/PLA-Logo-Final-01-1-1-300x36.png><br>`)
		for i, v := range m {
			fmt.Fprintf(w, `<a href="boot/%d/">%s</a>
<br>
`, i, v.Name)
		}
		io.WriteString(w, `</html>`)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func (c *Command) ServeMenus() *http.Server {
	return c.startHttpServer()
}
