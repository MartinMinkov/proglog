package server

import (
	"context"
	"flag"
	"net"
	"os"
	"testing"
	"time"

	api "github.com/MartinMinkov/proglog/api/v1"
	"github.com/MartinMinkov/proglog/internal/auth"
	"github.com/MartinMinkov/proglog/internal/config"
	"github.com/MartinMinkov/proglog/internal/log"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var debug = flag.Bool("debug", false, "Enable observability for debugging")

func TestMain(m *testing.M) {
	flag.Parse()
	if *debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		zap.ReplaceGlobals(logger)
	}
	os.Exit(m.Run())
}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, rootClient, nobodyClient api.LogClient, config *Config){
		"produce/consume a message to/form the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundry fails":                     testConsumePastBoundry,
		"unauthorized access fails":                          testUnauthorized,
	} {
		t.Run(scenario, func(t *testing.T) {
			rootClient, nobodyClient, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, config)
		})
	}
}

func setupTLSClient(t *testing.T, crtPath, keyPath, serverAddress string) (*grpc.ClientConn, api.LogClient, []grpc.DialOption) {
	t.Helper()
	tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: crtPath,
		KeyFile:  keyPath,
		CAFile:   config.CAFile,
		Server:   false,
	})
	require.NoError(t, err)
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
	conn, err := grpc.NewClient(serverAddress, opts...)
	require.NoError(t, err)
	client := api.NewLogClient(conn)
	return conn, client, opts
}

/**
 * Define a helper function to setup a test server and client.
 */
func setupTest(t *testing.T, fn func(*Config)) (
	rootClient, nobodyClient api.LogClient, c *Config, teardown func()) {
	t.Helper()

	// Set up a network listener for the server on a random port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dir, err := os.MkdirTemp(os.TempDir(), "server_test")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	authorizer := auth.New(config.ACLModelFile, config.ACLPolicyFile)
	cfg := &Config{
		CommitLog:  clog,
		Authorizer: authorizer,
	}

	var telemetryExporter *exporter.LogExporter
	if *debug {
		metricsLogFile, err := os.CreateTemp(os.TempDir(), "metrics-*.log")
		require.NoError(t, err)
		t.Logf("Metrics will be logged to %s", metricsLogFile.Name())

		tracesLogFile, err := os.CreateTemp(os.TempDir(), "traces-*.log")
		require.NoError(t, err)
		t.Logf("Traces will be logged to %s", tracesLogFile.Name())

		telemetryExporter, err = exporter.NewLogExporter(
			exporter.Options{
				MetricsLogFile:    metricsLogFile.Name(),
				TracesLogFile:     tracesLogFile.Name(),
				ReportingInterval: time.Second,
			})
		require.NoError(t, err)
		err = telemetryExporter.Start()
		require.NoError(t, err)
	}

	// If a function is provided, call it with the configuration
	if fn != nil {
		fn(c)
	}

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		ServerAddress: l.Addr().String(),
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
	})
	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)

	// Create a new GRPC server instance with the provided configuration
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	// Run a goroutine to serve the server on the listener
	go func() {
		server.Serve(l)
	}()

	// Configure the client to use the provided CA configuration to verify the server's certificate
	var rootConn *grpc.ClientConn
	rootConn, rootClient, _ = setupTLSClient(t, config.RootClientCertFile, config.RootClientKeyFile, l.Addr().String())

	var nobodyConn *grpc.ClientConn
	nobodyConn, nobodyClient, _ = setupTLSClient(t, config.NobodyClientCertFile, config.NobodyClientKeyFile, l.Addr().String())

	// Return the client, configuration, and a function to tear down the test
	return rootClient, nobodyClient, cfg, func() {
		server.Stop()
		rootConn.Close()
		nobodyConn.Close()
		l.Close()
		clog.Remove()
		if telemetryExporter != nil {
			time.Sleep(1500 * time.Millisecond)
			telemetryExporter.Stop()
			telemetryExporter.Close()
		}
	}
}

func testProduceConsume(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()
	want := &api.Record{
		Value: []byte("hello world"),
	}

	// Create a produce request and send it to the server with the record value
	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: want})
	require.NoError(t, err)

	// Create a consume request and send it to the server with the offset of the produce request.
	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	// Assert that the consumed record matches the produce record
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, produce.Offset, consume.Record.Offset)
}

func testConsumePastBoundry(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()

	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: &api.Record{Value: []byte("hello world")}})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset + 1})
	if consume != nil {
		t.Fatal("consume not nil")
	}
	got := status.Code(err)
	want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())
	if got != want {
		t.Fatalf("got err: %v, want err: %v", got, want)
	}
}

func testProduceConsumeStream(t *testing.T, client, _ api.LogClient, config *Config) {
	ctx := context.Background()
	records := []*api.Record{
		{Value: []byte("hello world"), Offset: 0},
		{Value: []byte("hello again"), Offset: 1},
	}
	{
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)
		for offset, record := range records {
			err = stream.Send(&api.ProduceRequest{Record: record})
			require.NoError(t, err)
			res, err := stream.Recv()
			require.NoError(t, err)
			if res.Offset != uint64(offset) {
				t.Fatalf("got offset: %d, want offset: %d", res.Offset, offset)
			}
		}
	}
	{
		stream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
		require.NoError(t, err)
		for i, record := range records {
			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, res.Record, &api.Record{
				Value:  record.Value,
				Offset: uint64(i),
			})
		}
	}

}

func testUnauthorized(t *testing.T, _, client api.LogClient, config *Config) {
	ctx := context.Background()
	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: &api.Record{Value: []byte("hello world")}})
	if produce != nil {
		t.Fatalf("produce response should be nil")
	}
	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got err code: %d, want err code: %d", gotCode, wantCode)
	}
	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: 0})
	if consume != nil {
		t.Fatalf("consume response should be nil")
	}
	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got err code: %d, want err code: %d", gotCode, wantCode)
	}
}
