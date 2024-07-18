package server

import (
	"errors"
	"sync"
)

var ErrOffsetNotFound = errors.New("offset not found")

/**
 * Log is a log that stores records. It is thread safe. Each record is stored in a store.
 */
type Log struct {
	mu      sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (c *Log) Append(record Record) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)
	return record.Offset, nil
}

func (c *Log) Read(offset uint64) (Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if offset >= uint64(len(c.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}

/**
 * Record is type that stores a data payload and an offset of where it's located in a log. Any arbitary data can be stored in a record.
 */
type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}
