package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/MartinMinkov/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	index      *index
	store      *store
	baseOffset uint64 // The base offset is used to calculate the relative offset of the index. Because we can have multiple segments, we need to keep track of the base offset for each segment.
	nextOffset uint64 // The next offset to be used when appending a record
	config     Config
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

func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	// The record offset is set to the next offset
	curr := s.nextOffset
	record.Offset = curr

	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}
	// Index offsets are relative to the base offset of the segment
	offset = s.nextOffset - uint64(s.baseOffset)
	if err = s.index.Write(uint32(offset), pos); err != nil {
		return 0, err
	}
	// Increment the next offset
	s.nextOffset++
	return curr, nil
}

func (s *segment) Read(off uint64) (*api.Record, error) {
	// We need to convert the absolute offset to a relative offset that we can use for the index.
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

func (s *segment) Remove() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	return nil
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size >= s.config.Segment.MaxIndexBytes
}

func nearestMultiple(j, k uint64) uint64 {
	return (j / k) * k

}
