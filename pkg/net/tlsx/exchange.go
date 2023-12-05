// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/rctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
)

const (
	ExchangeDrops = iota
	ExchangeIncompletes
	ExchangeInvalids
	ExchangeUnknowns
	ExchangeCounters
)

var exchange = struct {
	sync.RWMutex
	member      map[uint32]*Member
	pkt         net.PacketConn
	reservation map[Confirmation]*Member
	atomic      [ExchangeCounters]uint64
	halt        chan string
}{
	halt:        make(chan string),
	member:      make(map[uint32]*Member),
	reservation: make(map[Confirmation]*Member),
}

const ExchangeUsageTemplate = `
usage: {{.Path}} [-r] [-a <address>]
Start exchange at <address> (default :8003).
{{.Flag}}`

func ExchangeUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		pathctx.StringIn(ctx),
		flagctx.StringIn(ctx),
	}
}

func Exchange(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("exchange")
	ctx = flagctx.Parameter.With(ctx, fs)
	tcp := fs.String("tcp", ":8003", "service address")
	udp := fs.String("udp", ":8003", "packet service (disable if empty)")
	fs.BoolVar(&Restricted, "r", false, "restrict clients to self")
	if flag.Search[bool]("complete") {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(ExchangeUsageTemplate[1:],
			ExchangeUsageData(ctx))
	}
	args = fs.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	ln, err := net.Listen("tcp", *tcp)
	if err != nil {
		return err
	}

	var pc net.PacketConn
	if len(*udp) > 0 {
		var lc net.ListenConfig
		pc, err = lc.ListenPacket(ctx, "udp", *udp)
		if err != nil {
			ln.Close()
			return err
		}
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go Routine(ctx, &wg, ln, pc)

	return nil
}

const JoinUsageTemplate = `
usage: {{.}} <confirmation>
Join exchange with reserve confirmation number.`

func JoinUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

func Join(ctx context.Context, conn net.Conn, args ...string) error {
	if flag.Search[bool]("complete") {
		return nil
	}
	if flag.Search[bool]("help") {
		return usage.Error(JoinUsageTemplate[1:], JoinUsageData(ctx))
	}
	if len(args) < 1 {
		return ErrIncomplete
	}

	var cno Confirmation
	_, err := fmt.Sscan(args[0], &cno)
	if err != nil {
		return err
	}

	m, err := xjoin(conn, cno)
	if err != nil {
		return err
	}

	// BREAK to ack join before starting PDU exchange protocol.
	wctx.Parameter.In(ctx).(lv.Encoding).Encode(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go m.service(ctx, &wg)
	wg.Wait()

	return nil
}

const ReserveUsageTemplate = `
usage: {{.}} <name-or-subject-key-id>
Reserve exchange membership.`

func ReserveUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

func Reserve(ctx context.Context, args ...string) error {
	if flag.Search[bool]("complete") {
		return nil
	}
	if flag.Search[bool]("help") {
		return usage.Error(ReserveUsageTemplate[1:],
			ReserveUsageData(ctx))
	}
	if len(args) < 1 {
		return ErrIncomplete
	}

	i := Subscribers().Index(args[0])
	if i < 0 {
		return ErrNotFound
	}
	req, err := io.ReadAll(rctx.Parameter.In(ctx))
	if err != nil {
		return egress.Mark(err)
	}
	rsp, err := xreserve(uint32(i), req)
	if err != nil {
		return err
	}
	_, err = wctx.Parameter.In(ctx).Write(rsp)
	return egress.Mark(err)
}

const WhoIsUsageTemplate = `
usage: {{.}} <id>
Returns the PEM encoded data containing the public key and nonce of member.`

func WhoIsUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

func WhoIs(ctx context.Context, args ...string) error {
	if flag.Search[bool]("complete") {
		return nil
	}
	if flag.Search[bool]("help") {
		return usage.Error(WhoIsUsageTemplate[1:],
			WhoIsUsageData(ctx))
	}
	if len(args) < 1 {
		return ErrIncomplete
	}
	var id uint32
	_, err := fmt.Sscan(args[0], &id)
	if err != nil {
		return err
	}
	w := wctx.Parameter.In(ctx)
	if m := xwhois(id); m == nil {
		return egress.Mark(ErrNotFound)
	} else if _, err = w.Write(m.PublicKeyData); err != nil {
		return egress.Mark(err)
	}
	return nil
}

func Xcounter(counter int) uint64 {
	if counter < ExchangeCounters {
		return atomic.LoadUint64(&exchange.atomic[counter])
	}
	return 0
}

func xinc(counter int) {
	if counter < ExchangeCounters {
		atomic.AddUint64(&exchange.atomic[counter], 1)
	}
}

func xforward(bx box.Box) {
	exchange.RLock()
	defer exchange.RUnlock()
	to := bx.ToWhom()
	from := bx.FromWhom()
	if bx.IsToAll() {
		for id, m := range exchange.member {
			if id != from {
				clone := bx.Clone().Close(m.cipher).
					Seal(m.cipher)
				if m.addr != nil {
					exchange.pkt.WriteTo(clone, m.addr)
				} else {
					m.conn.Write(clone)
				}
				clone.Recycle()
			}
		}
	} else if m, ok := exchange.member[to]; ok {
		bx.Seal(m.cipher)
		if m.addr != nil {
			exchange.pkt.WriteTo(bx, m.addr)
		} else if m.conn != nil {
			m.conn.Write(bx)
		} else {
			log.Printf("%d: neither packet nor stream", to)
		}
	} else {
		log.Print(to, ": not found")
		xinc(ExchangeDrops)
	}
}

func xjoin(conn net.Conn, cno Confirmation) (
	*Member, error,
) {
	exchange.Lock()
	defer exchange.Unlock()
	m, ok := exchange.reservation[cno]
	if !ok {
		err := fmt.Errorf("%d: %w", cno, ErrNotFound)
		return nil, egress.Mark(err)
	}
	delete(exchange.reservation, cno)
	m.conn = conn
	exchange.member[m.id] = m
	return m, nil
}

// if successful, return PEM encoded public key of exchange and headers
// containing the assigned confirmation number and id.
func xreserve(id uint32, pubkeydata []byte) ([]byte, error) {
	var cno Confirmation
	_, err := cno.ReadFrom(rand.Reader)
	if err != nil {
		return nil, egress.Mark(err)
	}
	pubkey, _ := pem.Decode(pubkeydata)
	if pubkey == nil {
		return nil, egress.Mark(ErrInvalid)
	}
	bxc, err := box.NewCipher(pubkey)
	if err != nil {
		return nil, egress.Mark(err)
	}
	bxc.PublicKey.Local.Headers["confirmation"] = cno.String()
	bxc.PublicKey.Local.Headers["id"] = fmt.Sprintf("%d", id)
	m := &Member{
		id:     id,
		cipher: bxc,

		PublicKey: pubkey,

		PublicKeyData: pubkeydata,
	}
	if err = func() error {
		exchange.Lock()
		defer exchange.Unlock()
		if old, exists := exchange.member[id]; exists {
			if old.conn != nil {
				return egress.Mark(ErrExists)
			} else if old.addr != nil {
				exchange.halt <- old.addr.String()
			}
		}
		exchange.reservation[cno] = m
		return nil
	}(); err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(m.cipher.PublicKey.Local), nil
}

func xpacket(
	ctx context.Context,
	wg *sync.WaitGroup,
	pkt net.PacketConn,
) {
	defer wg.Done()

	exchange.pkt = pkt
	pktctx := poll.WithReadFromer(ctx, exchange.pkt)
	mat := make(map[string]*Member)
	for bx := box.New(); true; bx = bx.Expand() {
		n, addr, err := pktctx.ReadFrom(bx)
		if err != nil {
			log.Print(err)
			break
		}
		select {
		case s := <-exchange.halt:
			if m, ok := mat[s]; ok {
				m.resign()
				delete(mat, s)
			}
		default:
		}
		if n == 8 {
			var cno Confirmation
			if err = cno.UnmarshalBinary(bx); err != nil {
				log.Print(err)
				continue
			}
			func() {
				exchange.Lock()
				defer exchange.Unlock()
				m, ok := exchange.reservation[cno]
				if ok {
					delete(exchange.reservation, cno)
					m.addr = addr
					exchange.member[m.id] = m
					mat[addr.String()] = m
				} else {
					log.Print("unmatched: ", cno)
				}
			}()
		} else if n < box.BeginContent {
			log.Print("incomplete")
			xinc(ExchangeIncompletes)
		} else if m, ok := mat[addr.String()]; !ok {
			log.Print("unknown")
			xinc(ExchangeUnknowns)
		} else if bx, err = bx.Unseal(m.cipher); err != nil {
			log.Print("invalid")
			xinc(ExchangeInvalids)
		} else {
			if bx.IsToAll() {
				if bx, err = bx.Open(m.cipher); err != nil {
					log.Print(err)
					continue
				}
			}
			xforward(bx)
		}
	}
}

func xwhois(id uint32) *Member {
	exchange.RLock()
	defer exchange.RUnlock()
	return exchange.member[id]
}

type Member struct {
	sync.RWMutex
	id     uint32
	cipher *box.Cipher
	conn   net.Conn
	addr   net.Addr

	PublicKey *pem.Block

	PublicKeyData []byte
}

func (m *Member) service(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer m.resign()

	bx := box.New()
	defer bx.Recycle()
	var err error
	for {
		if bx, err = bx.Receive(ctx, m.conn, m.cipher); err != nil {
			log.Print(err)
			break
		}
		if bx.IsToAll() {
			if bx, err = bx.Open(m.cipher); err != nil {
				log.Print(err)
				break
			}
		}
		xforward(bx)
	}
}

func (m *Member) resign() {
	exchange.Lock()
	defer exchange.Unlock()
	delete(exchange.member, m.id)
	m.conn = nil
	m.addr = nil
	m.PublicKey = nil
}
