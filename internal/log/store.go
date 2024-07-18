package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

/**
 * Store is a log store that uses a buffered writer to write to a file. This is where we actually store our record data.
 * We store the file and the size of the file, as well as a mutex to ensure thread safety.
 */
type store struct {
	File *os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	// We write the length of the payload to the buffer in big endian format. This lets us efficiently read the payload from the file when we need to later.
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	// Write the payload to the buffer.
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	// Add the length of the payload to the size of the file.
	w += lenWidth
	// Update the size of the file.
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// We need to flush the buffer before reading from the file, to ensure that the data is written to disk.
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	// Construct a buffer to read the length of the payload from the file.
	size := make([]byte, lenWidth)
	// We read the length of the payload at the given position in the file
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	// We construct a buffer to read the payload data from the file.
	b := make([]byte, enc.Uint64(size))
	// We read the payload of the data where at the position plus the length of the payload length header.
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
