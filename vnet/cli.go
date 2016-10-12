package vnet

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"
)

func (v *Vnet) CliAdd(c *cli.Command)                     { v.loop.CliAdd(c) }
func (v *Vnet) Logf(format string, args ...interface{})   { v.loop.Logf(format, args...) }
func (v *Vnet) Fatalf(format string, args ...interface{}) { v.loop.Fatalf(format, args...) }

type cliListener struct {
	socketConfig  string
	disablePrompt bool
	server        *cli.Server
}

func (l *cliListener) Parse(in *parse.Input) {
	for !in.End() {
		switch {
		case in.Parse("no-prompt"):
			l.disablePrompt = true
		case in.Parse("socket %s", &l.socketConfig):
		default:
			panic(parse.ErrInput)
		}
	}
}

type cliMain struct {
	Package
	v           *Vnet
	enableStdin bool
	listeners   []cliListener
}

func (v *Vnet) CliInit() {
	m := &v.cliMain
	m.v = v
	v.AddPackage("cli", m)
}

func (m *cliMain) Configure(in *parse.Input) {
	for !in.End() {
		var (
			l  cliListener
			li parse.Input
		)
		switch {
		case in.Parse("listen %v", &li) && li.Parse("%v", &l):
			m.listeners = append(m.listeners, l)
		case in.Parse("stdin"):
			m.enableStdin = true
		default:
			panic(parse.ErrInput)
		}
	}
}

func (m *cliMain) Init() (err error) {
	m.v.loop.Cli.Prompt = "vnet# "
	if m.enableStdin {
		m.v.loop.Cli.AddStdin()
	}
	for i := range m.listeners {
		l := &m.listeners[i]
		l.server, err = m.v.loop.Cli.AddServer(l.socketConfig, l.disablePrompt)
		if err != nil {
			return
		}
	}
	return
}

func (m *cliMain) Exit() (err error) {
	for i := range m.listeners {
		l := &m.listeners[i]
		l.server.Close()
	}
	return
}
