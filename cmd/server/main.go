package main

import "log"

func main() {
	app, err := SetupGRPCServer()
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer app.Cleanup()
	app.Start()
}
