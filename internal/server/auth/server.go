package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/gogo/googleapis/google/rpc"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
)

const (
	grpcMaxConcurrentStreams = 1000000
)

type server struct {
}

func (s *server) Check(ctx context.Context, req *authservice.CheckRequest) (*authservice.CheckResponse, error) {
	log.Println("SGTM!!")
	resp := &authservice.CheckResponse{
		Status: &rpcstatus.Status{Code: int32(rpc.OK)},
	}
	return resp, nil
}

func RunServer(ctx context.Context, port uint) {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems.
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	authservice.RegisterAuthorizationServer(grpcServer, &server{})

	log.Printf("auth server listening on %d\n", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}
