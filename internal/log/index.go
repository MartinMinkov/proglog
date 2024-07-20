package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	// Get the file info from the provided file.
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	// Store the size of the file in our index
	idx.size = uint64(fi.Size())
	// Truncate the file to the max index size (given in the config). The reason we do this now is because we cannot change the size of the file after it's been mapped into memory.
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}
	// Map the file to memory, with read and write permissions.
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Close() error {
	// Before closing the file, we need to flush the mmap to disk.
	if err := i.mmap.Sync(gommap.MS_ASYNC); err != nil {
		return err
	}
	// Before closing the file, we need to also flush any pending writes to disk.
	if err := i.file.Sync(); err != nil {
		return err
	}
	// Truncate the file size to the current size of the index
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	// If the index is empty, we return EOF.
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		// If we are given -1, we return the last index entry.
		out = uint32((i.size / entWidth) - 1)
	} else {
		// Otherwise, we store the index entry at the given position.
		out = uint32(in)
	}
	// Calculate the position of where to read the index entry from.
	pos = uint64(out) * entWidth
	// If the position is past the end of the file, we return EOF.
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	// Read the index entry from the file.
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	// Read the position of the record from the file.
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	// If the memory mapped file is not large enough to hold the index entry, we return an error.
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	// Write the offset to the index, placed at the end of the file plus 4 bytes for the offset.
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	// Write the position of the record to the index, placed at the end of the file plus 8 bytes for the position.
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	// Update the size of the index
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
