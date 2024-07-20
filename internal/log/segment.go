package log

import (
	"fmt"
	"os"
	"path"
)

type segment struct {
	index                  *index
	store                  *store
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error

	storeFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	if off, _, err := s.index.Read(-1); err != nil {
		// If the index is empty, we set the next offset to the initial offset.
		s.nextOffset = baseOffset
	} else {
		// Otherwise, we set the next offset to the offset of the last index entry.
		s.nextOffset = baseOffset + uint64(off) + 1
	}
	return s, nil
}
