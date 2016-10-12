package main

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/elib/socket"
	"github.com/platinasystems/go/elib/srpc"

	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

type rpcServer struct {
	config
	socket.Server
	lock       sync.Mutex
	clients    []*rpcClient
	bufRecycle chan elib.ByteVec
}

func newRpcServer(cfg *config) (s *rpcServer, err error) {
	s = &rpcServer{}
	s.bufRecycle = make(chan elib.ByteVec, 64)
	err = s.Config(cfg.socketConfig, socket.Listen)
	s.config = *cfg
	return
}

// Return a possibly recycled zero-length buffer ready to use.
func (s *rpcServer) getBuf() (b elib.ByteVec) {
	select {
	case b = <-s.bufRecycle:
		b = b[:0]
	default:
	}
	return
}

// Return a previously used buffer for recycling.
func (s *rpcServer) putBuf(b elib.ByteVec) {
	select {
	case s.bufRecycle <- b:
	default:
	}
}

func (s *rpcServer) ReadReady() (err error) {
	c := &rpcClient{}
	err = s.AcceptClient(&c.Client)
	if err != nil {
		return
	}
	err = newRpcClient(c, "")
	if err != nil {
		return
	}
	if s.verbose {
		fmt.Printf("server: new client %s <- %s\n", socket.SockaddrString(c.SelfAddr), socket.SockaddrString(c.PeerAddr))
	}
	return
}

type rpcClient struct {
	socket.Client
	srpc.Server
	index uint32
	iter  uint32
}

func newRpcClient(c *rpcClient, socketConfig string) (err error) {
	s := c.RPCServer()
	s.lock.Lock()
	c.index = uint32(len(s.clients))
	s.clients = append(s.clients, c)
	s.lock.Unlock()

	if len(socketConfig) > 0 {
		err = c.Config(socketConfig, socket.Flags(0))
		if err != nil {
			return
		}
	}
	iomux.Add(c)

	c.EventTag = fmt.Sprintf("%s", c)
	c.InitWriter(c, s.bufRecycle, new(T))

	go c.runClient()

	return
}

func (c *rpcClient) ReadReady() (err error) {
	err = c.Client.ReadReady()
	if len(c.RxBuffer) > 0 {
		n := c.Input(c.RxBuffer)
		b := []byte{} // c.RPCServer().getBuf()
		if n < len(c.RxBuffer) {
			b = append(b, c.RxBuffer[n:]...)
		}
		c.RxBuffer = b
	}
	return
}

func (c *rpcClient) WriteReady() (err error) {
	newConnection, err := c.ClientWriteReady()
	if newConnection && c.RPCServer().verbose {
		fmt.Printf("client: connected %s\n", c)
	}
	return
}

// Needed otherwise ambiguous since both Socket & Rpc have Close() method.
func (c *rpcClient) Close() error { return c.Client.Close() }

func (c *rpcClient) RPCServer() *rpcServer { return defaultServer }

var defaultServer *rpcServer

func elogDumpOnSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	w := os.Stderr
	fmt.Fprintf(w, "%d events in log:\n", elog.Len())
	elog.Print(w)
	os.Exit(0)
}

func main() {
	var err error

	cfg := &config{}

	cfg.nIter = 10
	cfg.nClient = 1

	flag.StringVar(&cfg.socketConfig, "config", "localhost:5555", "address:port for listen/connect")
	flag.Var(&cfg.nIter, "iter", "Number of calls per client")
	flag.Var(&cfg.nClient, "clients", "Number of clients")
	flag.Var(&cfg.printEvery, "print", "Number of calls between status outputs")
	flag.IntVar(&cfg.minMsgBytes, "min-size", 1, "Min message size in bytes")
	flag.IntVar(&cfg.maxMsgBytes, "max-size", 10, "Max message size in bytes")
	flag.BoolVar(&cfg.elog, "elog", false, "Enable event logging")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Verbose")
	flag.Float64Var(&cfg.minDelay, "min-delay", 0, "Min delay between RPC calls in seconds")
	flag.Float64Var(&cfg.maxDelay, "max-delay", 0, "Max delay between RPC calls in seconds")
	flag.Parse()

	elog.Enable(cfg.elog)

	if cfg.elog {
		go elogDumpOnSignal()
	}

	defaultServer, err = newRpcServer(cfg)
	s := defaultServer
	if err != nil {
		panic(err)
	}
	iomux.Add(s)

	go iomux.Wait(false)

	for i := 0; i < int(cfg.nClient); i++ {
		c := &rpcClient{}
		err = newRpcClient(c, cfg.socketConfig)
		if err != nil {
			panic(err)
		}
	}

	for {
		n := len(s.clients)
		max := uint32(0)
		min := ^max
		for i := 0; i < n; i++ {
			c := s.clients[i]
			if c.iter < min {
				min = c.iter
			}
			if c.iter > max {
				max = c.iter
			}
		}
		fmt.Printf("%d rpc clients, %d go routines, min %d, max %d\n", n, runtime.NumGoroutine(), min, max)
		if min == uint32(s.nIter) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

type reqEvent struct {
	index  uint32
	client uint32
	s      [elog.EventDataBytes - 2*4]byte
}

func (e *reqEvent) String() string {
	c := defaultServer.clients[e.client]
	return fmt.Sprintf("call %s #%d %s", c, 1+e.index, elog.String(e.s[:]))
}

func (e *reqEvent) Encode(b []byte) (i int) {
	i += elog.EncodeUint32(b[i:], e.index)
	i += elog.EncodeUint32(b[i:], e.client)
	i += copy(b[i:], e.s[:])
	return
}
func (e *reqEvent) Decode(b []byte) (i int) {
	e.index, i = elog.DecodeUint32(b, i)
	e.client, i = elog.DecodeUint32(b, i)
	i += copy(e.s[:], b[i:])
	return i
}

//go:generate gentemplate -d Package=main -id ReqEvent -d Type=reqEvent github.com/platinasystems/go/elib/elog/event.tmpl

type Req struct {
	A string
}

func randString(min, max int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	n := min
	if max > min {
		n += rand.Intn(max - min)
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type T struct{}

func (t *T) F(req *Req, rep *Req) error {
	rep.A = req.A
	return nil
}

type config struct {
	socketConfig       string
	nIter              elib.Count
	printEvery         elib.Count
	nClient            elib.Count
	maxMsgBytes        int
	minMsgBytes        int
	minDelay, maxDelay float64
	seed               int64
	verbose            bool
	elog               bool
}

var once sync.Once

func (c *rpcClient) runClient() {
	s := c.RPCServer()
	rand.Seed(s.seed + int64(c.index))
	for i := uint32(0); s.nIter == 0 || i < uint32(s.nIter); i++ {
		var req, rep Req

		req = Req{A: randString(s.minMsgBytes, s.maxMsgBytes)}

		event := reqEvent{index: i, client: c.index}
		elog.Printf(event.s[:], "`%s' %x", req.A, req.A)
		event.Log()

		err := c.Call("T.F", &req, &rep)
		if err != nil {
			panic(err)
		}

		event = reqEvent{index: i, client: c.index}
		elog.Printf(event.s[:], "done `%s' -> `%s'", req.A, rep.A)
		event.Log()

		if s.printEvery != 0 && (i%uint32(s.printEvery) == 0 || i+1 >= uint32(s.nIter)) {
			fmt.Printf("%s iter %d\n", c, i)
		}

		c.iter++

		if s.maxDelay > 0 {
			dt := s.minDelay + (s.maxDelay-s.minDelay)*rand.Float64()
			time.Sleep(time.Duration(1e9 * dt))
		}
	}
}
