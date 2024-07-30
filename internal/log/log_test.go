package log

import (
	"io"
	"os"
	"testing"

	api "github.com/MartinMinkov/proglog/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T, log *Log,
	){
		"append and read a record succeeds": testAppendRead,
		"offset out of range errors":        testOutOfRange,
		"init with existing segments":       testInitExisting,
		"reader":                            testReader,
		"truncate":                          testTruncate,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := os.MkdirTemp(os.TempDir(), "log_test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			c := Config{}
			c.Segment.MaxStoreBytes = 32
			log, err := NewLog(dir, c)
			require.NoError(t, err)
			fn(t, log)
		})
	}

}

func testAppendRead(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, append.Value, read.Value)
}

func testOutOfRange(t *testing.T, log *Log) {
	read, err := log.Read(1)
	require.Nil(t, read)
	apiError := err.(api.ErrOffsetOutOfRange)
	require.Equal(t, uint64(1), apiError.Offset)
}

func testInitExisting(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}
	require.NoError(t, log.Close())

	offset, err := log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	offset, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)

	n, err := NewLog(log.Dir, log.Config)
	require.NoError(t, err)

	offset, err = n.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	offset, err = n.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), offset)
}

func testReader(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}

	offset, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	reader := log.Reader()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, read.Value, append.Value)
}

func testTruncate(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err)
}
