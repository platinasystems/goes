// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"fmt"
)

type EventTrack struct {
	Name  string
	index uint32
}

type eventTrackShared struct {
	trackByName map[string]*EventTrack
	tracks      []*EventTrack
}

func (s *eventTrackShared) AddTrack(format string, args ...interface{}) uint32 {
	t := &EventTrack{Name: fmt.Sprintf(format, args...)}
	if s.trackByName == nil {
		s.trackByName = make(map[string]*EventTrack)
	}
	s.trackByName[t.Name] = t
	t.index = uint32(len(s.tracks))
	s.tracks = append(s.tracks, t)
	return t.index
}
