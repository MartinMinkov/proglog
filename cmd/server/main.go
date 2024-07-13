package main

import (
	"log"

	"github.com/MartinMinkov/proglog/internal/server"
)

func main() {
	address := server.NewAddress("localhost", "8080")
	server := server.NewAPIServer(address)
	log.Fatal(server.Start())

}
