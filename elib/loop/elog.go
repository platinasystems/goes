package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
)

type poller_elog_event_type byte

const (
	poller_start poller_elog_event_type = iota + 1
	poller_done
	poller_suspend
	poller_resume
	poller_resumed
	poller_activate
)

var poller_elog_event_type_names = [...]string{
	poller_start:    "start",
	poller_done:     "done",
	poller_suspend:  "suspend",
	poller_resume:   "resume",
	poller_resumed:  "resumed",
	poller_activate: "activate",
}

func (t poller_elog_event_type) String() string {
	return elib.Stringer(poller_elog_event_type_names[:], int(t))
}

type pollerElogEvent struct {
	poller_index byte
	event_type   poller_elog_event_type
	flags        byte
	name         [elog.EventDataBytes - 2]byte
}

func (n *Node) pollerElog(t poller_elog_event_type, f node_flags) {
	if elog.Enabled() {
		le := pollerElogEvent{
			event_type:   t,
			poller_index: byte(n.activePollerIndex),
			flags:        byte(f),
		}
		copy(le.name[:], n.name)
		le.Log()
	}
}

func (e *pollerElogEvent) String() (s string) {
	s = "poller"
	if e.poller_index != 0xff {
		s += fmt.Sprintf(" %d:", e.poller_index)
	}
	s += fmt.Sprintf(" %s %s", elog.String(e.name[:]), e.event_type)
	if e.flags != 0 {
		s += fmt.Sprintf(", flags: %s", node_flags(e.flags))
	}
	return
}
func (e *pollerElogEvent) Encode(b []byte) int {
	b = elog.PutUvarint(b, int(e.poller_index))
	b = elog.PutUvarint(b, int(e.event_type))
	b = elog.PutUvarint(b, int(e.flags))
	return copy(b, e.name[:])
}
func (e *pollerElogEvent) Decode(b []byte) int {
	var i [3]int
	b, i[0] = elog.Uvarint(b)
	b, i[1] = elog.Uvarint(b)
	b, i[2] = elog.Uvarint(b)
	e.poller_index = byte(i[0])
	e.event_type = poller_elog_event_type(i[1])
	e.flags = byte(i[2])
	return copy(e.name[:], b)
}

//go:generate gentemplate -d Package=loop -id pollerElogEvent -d Type=pollerElogEvent github.com/platinasystems/go/elib/elog/event.tmpl

type eventElogEvent struct {
	s [elog.EventDataBytes]byte
}

func (e *eventElogEvent) String() string {
	return "loop event: " + elog.String(e.s[:])
}
func (e *eventElogEvent) Encode(b []byte) int { return copy(b, e.s[:]) }
func (e *eventElogEvent) Decode(b []byte) int { return copy(e.s[:], b) }

//go:generate gentemplate -d Package=loop -id eventElogEvent -d Type=eventElogEvent github.com/platinasystems/go/elib/elog/event.tmpl
