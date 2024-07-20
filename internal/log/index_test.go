package log

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	idx, err := newIndex(f, c)
	require.NoError(t, err)
	_, _, err = idx.Read(-1)
	// We expect an EOF error because the index is empty.
	require.Error(t, err)
	require.Equal(t, f.Name(), idx.Name())
	entries := []struct {
		Off uint32
		Pos uint64
	}{
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}

	for _, e := range entries {
		err := idx.Write(e.Off, e.Pos)
		require.NoError(t, err)
		_, pos, err := idx.Read(int64(e.Off))
		require.NoError(t, err)
		require.Equal(t, e.Pos, pos)
	}

	// We expect an EOF error when we read past the existing entries.
	_, _, err = idx.Read(int64(len(entries)))
	require.Equal(t, io.EOF, err)
	_ = idx.Close()

	// If we create a new index with the same file, we should expect the same state as before
	f, err = os.OpenFile(f.Name(), os.O_RDWR, 0600)
	require.NoError(t, err)
	idx, err = newIndex(f, c)
	require.NoError(t, err)
	off, pos, err := idx.Read(-1)
	require.NoError(t, err)
	require.Equal(t, entries[1].Off, off)
	require.Equal(t, entries[1].Pos, pos)
}
