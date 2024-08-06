package main

import (
	"net"
	"os"

	"github.com/MartinMinkov/proglog/internal/log"
	"github.com/MartinMinkov/proglog/internal/server"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	dir, err := os.MkdirTemp(os.TempDir(), "proglog")
	if err != nil {
		panic(err)
	}

	clog, err := log.NewLog(dir, log.Config{})
	if err != nil {
		panic(err)
	}
	defer clog.Remove()

	config := &server.Config{
		CommitLog: clog,
	}
	server, err := server.NewGRPCServer(config)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	go func() {
		server.Serve(listener)
	}()
}
