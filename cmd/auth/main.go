package main

import (
	"context"
	"flag"

	server "github.com/weinong/envoy-control-plane/internal/server/auth"
)

var (
	port   uint
	nodeID string
)

func init() {
	// The port that this auth server listens on
	flag.UintVar(&port, "port", 9002, "auth server port")

}

func main() {
	flag.Parse()

	ctx := context.Background()
	server.RunServer(ctx, port)
}
