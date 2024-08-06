package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/MartinMinkov/proglog/internal/auth"
	"github.com/MartinMinkov/proglog/internal/config"
	"github.com/MartinMinkov/proglog/internal/log"
	"github.com/MartinMinkov/proglog/internal/server"
	"go.opencensus.io/examples/exporter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCServerResult struct {
	Server  *grpc.Server
	Log     *log.Log
	Cleanup func()
	Start   func()
}

func setupTelemetryServer() (*exporter.LogExporter, error) {
	var telemetryExporter *exporter.LogExporter
	if os.Getenv("DEBUG") == "true" {
		metricsLogFile, err := os.CreateTemp(os.TempDir(), "metrics-*.log")
		if err != nil {
			return nil, err
		}
		fmt.Printf("Metrics will be logged to %s", metricsLogFile.Name())

		tracesLogFile, err := os.CreateTemp(os.TempDir(), "traces-*.log")
		if err != nil {
			return nil, err
		}
		fmt.Printf("Traces will be logged to %s", tracesLogFile.Name())

		telemetryExporter, err = exporter.NewLogExporter(
			exporter.Options{
				MetricsLogFile:    metricsLogFile.Name(),
				TracesLogFile:     tracesLogFile.Name(),
				ReportingInterval: time.Second,
			})
		if err != nil {
			return nil, err
		}
		err = telemetryExporter.Start()
		if err != nil {
			return nil, err
		}
	}
	return telemetryExporter, nil
}

func setupTLSServerConfig(serverAddress string) (credentials.TransportCredentials, error) {
	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		ServerAddress: serverAddress,
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
	})
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(serverTLSConfig), nil
}

func SetupGRPCServer() (*GRPCServerResult, error) {
	// TODO: Make listener configurable
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp(os.TempDir(), "proglog")
	if err != nil {
		return nil, err
	}

	clog, err := log.NewLog(dir, log.Config{})
	if err != nil {
		return nil, err
	}

	authorizer := auth.New(config.ACLModelFile, config.ACLPolicyFile)
	config := &server.Config{
		CommitLog:  clog,
		Authorizer: authorizer,
	}

	serverTLS, err := setupTLSServerConfig(listener.Addr().String())
	if err != nil {
		return nil, err
	}
	server, err := server.NewGRPCServer(config, grpc.Creds(serverTLS))
	if err != nil {
		return nil, err
	}

	telemetryExporter, err := setupTelemetryServer()
	if err != nil {
		return nil, err
	}

	start := func() {
		server.Serve(listener)
	}

	cleanup := func() {
		server.Stop()
		listener.Close()
		clog.Remove()
		if telemetryExporter != nil {
			time.Sleep(1500 * time.Millisecond)
			telemetryExporter.Stop()
			telemetryExporter.Close()
		}
	}

	return &GRPCServerResult{Server: server,
		Log:     clog,
		Start:   start,
		Cleanup: cleanup}, nil
}
