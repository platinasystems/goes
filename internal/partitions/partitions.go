// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package partitions

import (
	"errors"
	"os"

	"github.com/platinasystems/go/internal/magic"
	"github.com/satori/go.uuid"
)

var ErrNotFilesystem = errors.New("not a filesystem")
var ErrNotSupported = errors.New("filesystem feature not supported")

type superBlock interface {
	UUID() (uuid.UUID, error)
}

type unknownSB struct {
}

func (sb *unknownSB) UUID() (uuid.UUID, error) {
	return uuid.Nil, ErrNotSupported
}

type ext234 struct {
	sUUID uuid.UUID
}

const (
	ext234SUUIDOff = 0x468
	ext234SUUIDLen = 16
)

func (sb *ext234) UUID() (uuid.UUID, error) {
	return sb.sUUID, nil
}

func ReadSuperBlock(dev string) (superBlock, error) {
	f, err := os.Open(dev)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fsHeader := make([]byte, 0x10000)
	_, err = f.Read(fsHeader)
	if err != nil {
		return nil, err
	}

	partitionMapType := magic.IdentifyPartitionMap(fsHeader)
	partitionType := magic.IdentifyPartition(fsHeader)

	if partitionMapType != "" {
		return nil, ErrNotFilesystem
	}

	if partitionType == "ext234" {
		sb := &ext234{}
		sb.sUUID = uuid.FromBytesOrNil(fsHeader[ext234SUUIDOff : ext234SUUIDOff+ext234SUUIDLen])
		return sb, nil
	}

	return &unknownSB{}, nil
}
