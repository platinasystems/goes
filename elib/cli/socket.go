package cli

import (
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/elib/socket"

	"fmt"
	"sync"
)

type Server struct {
	main *Main
	socket.Server
	socketConfig  string
	verbose       bool
	disablePrompt bool
	// Locks client pool.
	lock sync.Mutex
	clientPool
}

func (s *Server) ReadReady() (err error) {
	template := client{}
	err = s.AcceptClient(&template.Client)
	if err != nil {
		return
	}
	template.server = s
	ci, err := s.newClient(&template, "")
	if err != nil {
		return
	}
	if s.verbose {
		c := &s.clients[ci]
		fmt.Printf("server: new client #%d %s <- %s\n", c.index, socket.SockaddrString(c.SelfAddr), socket.SockaddrString(c.PeerAddr))
	}
	return
}

type client struct {
	server *Server
	socket.Client
	index uint
}

//go:generate gentemplate -d Package=cli -id client -d Data=clients -d PoolType=clientPool -d Type=client github.com/platinasystems/go/elib/pool.tmpl

func (s *Server) newClient(template *client, socketConfig string) (i uint, err error) {
	s.lock.Lock()
	i = s.clientPool.GetIndex()
	c := &s.clients[i]
	*c = *template
	c.index = i
	s.lock.Unlock()
	if len(socketConfig) > 0 {
		err = c.Config(socketConfig, socket.Flags(0))
		if err != nil {
			return
		}
	}
	c.server.main.AddFile(c, s.disablePrompt)
	return
}

func (c *client) ReadReady() (err error) {
	err = c.Client.ReadReady()
	if err != nil {
		return
	}
	if c.Client.IsClosed() {
		c.done("eof")
	}
	return
}

func (c *client) done(reason string) {
	s := c.server
	if s.verbose {
		fmt.Printf("server: %s client #%d %s <- %s\n", reason, c.index, socket.SockaddrString(c.SelfAddr), socket.SockaddrString(c.PeerAddr))
	}
	s.lock.Lock()
	s.clientPool.PutIndex(c.index)
	s.lock.Unlock()
}

func (c *client) Close() (err error) {
	err = c.Client.Close()
	if err != nil {
		return
	}
	c.done("requested-close")
	return
}

func (c *Main) AddServer(config string, disablePrompt ...bool) (s *Server, err error) {
	s = &Server{main: c}
	if len(disablePrompt) > 0 {
		s.disablePrompt = disablePrompt[0]
	}
	err = s.Config(config, socket.Listen)
	if err != nil {
		return
	}
	iomux.Add(s)
	c.servers = append(c.servers, s)
	return
}
