// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !nowebserver,!bootrom

package grub

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/platinasystems/goes/cmd/grub/menu"
)

var pageExists = make(map[string]struct{})

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

func (c *Command) serveKexecMenu(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `<html>`)
	kexec := c.KexecCommand()
	if len(kexec) == 0 {
		fmt.Fprintf(w, "No kernel defined, did you bookmark this URL?<br>")
	} else {
		err := c.g.Main(kexec...)
		if err != nil {
			fmt.Fprintf(w, "Failed: %s<br>", html.EscapeString(err.Error()))
		} else {
			io.WriteString(w, "Success, so how do you see this?")
		}
	}
	io.WriteString(w, `</html>`)
}

func (c *Command) menuHtml(w http.ResponseWriter, m *menu.Menu, parent []int) {
	mp := ""
	sep := "?"
	for _, i := range parent {
		mp = mp + sep + "m=" + strconv.Itoa(i)
		sep = "&"
	}
	if m != nil {
		for i, v := range *(m.Entries) {
			fmt.Fprintf(w, `<a href="%s%sm=%d">%s</a>
<br>
`, mp, sep, i, v.Name)
		}
	}
	k := c.KexecCommand()
	if len(k) > 0 {
		fmt.Fprintf(w, `<br><a href="kexec">Kexec command: %v</a><br>`, k)
	}
}

//		io.WriteString(w, `<html><img src=http://www.platinasystems.com/wp-content/uploads/2016/10/PLA-Logo-Final-01-1-1-300x36.png><br>`)

func (c *Command) serveRootMenu(n string, w http.ResponseWriter, r *http.Request) {
	ws := &webserver{w: w}
	Cli.Stdin = ws
	Cli.Stdout = ws
	Cli.Stderr = ws

	io.WriteString(w, `<html>`)
	menuPath := r.URL.Query()["m"]
	if menuPath == nil {
		menuEntry.Reset()
		if err := c.runScript(n); err != nil {
			fmt.Fprintf(w, "Script returned error: %s<br>",
				html.EscapeString(err.Error()))
		} else {
			c.menuHtml(w, menuEntry.R.RootMenu, nil)
		}
	} else {
		mp := make([]int, len(menuPath))
		err := error(nil)
		for i := 0; i < len(menuPath); i++ {
			mp[i], err = strconv.Atoi(menuPath[i])
			if err != nil {
				break
			}
		}
		if err != nil {
			fmt.Fprintf(w, "Error parsing menu path: %s</br>",
				html.EscapeString(err.Error()))
		} else {
			e, err := menuEntry.FindEntry(mp...)
			if err != nil {
				fmt.Fprintf(w, "Error finding menu path %v: %s</br>",
					mp, err)
			} else {
				if err := e.RunFun(ws, ws, ws); err != nil {
					fmt.Fprintf(w, "Menu returned error: %s</br>",
						html.EscapeString(err.Error()))
				} else {
					c.menuHtml(w, e.Submenu, mp)
				}
			}
		}
	}
	io.WriteString(w, `</html>`)
}

func (c *Command) ServeMenus(n string) {
	srv := &http.Server{Addr: ":8080"}
	mutex := &sync.Mutex{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if r.URL.Path == "/" {
			c.serveRootMenu(n, w, r)
			return
		}
		if r.URL.Path == "/kexec" {
			c.serveKexecMenu(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	if err := srv.ListenAndServe(); err != nil {
		// cannot panic, because this probably is an intentional close
		log.Printf("Httpserver: ListenAndServe() error: %s", err)
	}
}
