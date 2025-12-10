package main

import (
	"fmt"
	"log"
	"os"
)

const version = "0.1.0"

func main() {
	log.Printf("goCoax Prometheus Exporter v%s", version)
	log.Println("Starting exporter...")

	// TODO: Load configuration
	// TODO: Initialize collectors
	// TODO: Start HTTP server

	fmt.Fprintln(os.Stderr, "Not yet implemented")
	os.Exit(1)
}
