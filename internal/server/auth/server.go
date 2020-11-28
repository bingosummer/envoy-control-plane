package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/weinong/envoy-control-plane/apis/v1alpha1"
	"github.com/weinong/envoy-control-plane/internal/utils"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

const (
	grpcMaxConcurrentStreams = 1000000
)

type Server struct {
	*v1alpha1.ExtAuthConfig
}

func (s *Server) ParseConfig(file string) {
	config, err := utils.ParseExtAuthConfig(file)
	if err != nil {
		log.Printf("unable to read ext auth config: %s", err)
		return
	}

	log.Printf("incoming external auth config is %+v", config)
	s.ExtAuthConfig = config
}

func (s *Server) Check(ctx context.Context, req *authservice.CheckRequest) (*authservice.CheckResponse, error) {

	token := strings.TrimPrefix(req.Attributes.Request.Http.Headers["authorization"], "Bearer ")
	log.Printf("incoming bearer token is %s", token)

	route := ""
	if s.ExtAuthConfig != nil {
		// find the route associated with this bearer token
		for k, v := range s.RequiredTokenByRoutes {
			if token == v {
				route = k
			}
		}
	}
	if route == "" {
		log.Printf("cannot find matched route with bearer token %s", token)
		return &authservice.CheckResponse{
			Status: &status.Status{Code: int32(rpc.PERMISSION_DENIED)},
		}, nil
	}
	resp := &authservice.CheckResponse{
		Status: &status.Status{Code: int32(rpc.OK)},
	}
	headers := []*core.HeaderValueOption{
		{
			Header: &core.HeaderValue{
				Key:   "x-route",
				Value: route,
			},
		},
	}
	if s.ExtAuthConfig != nil {
		if s.AuthorizationToken != "" {
			headers = append(headers, &core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   "Authorization",
					Value: fmt.Sprintf("Bearer %s", s.AuthorizationToken),
				},
			})
		}

		for k, v := range s.ExtAuthConfigSpec.AdditionalHeaders {
			headers = append(headers, &core.HeaderValueOption{
				Header: &core.HeaderValue{
					Key:   k,
					Value: v,
				},
			})
		}
	}
	if len(headers) > 0 {
		resp.HttpResponse = &authservice.CheckResponse_OkResponse{
			OkResponse: &authservice.OkHttpResponse{
				Headers: headers,
			},
		}
	}
	log.Println("SGTM!!")
	return resp, nil
}

func NewServer() *Server {
	return &Server{}
}

func Run(ctx context.Context, server *Server, port uint) {
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

	authservice.RegisterAuthorizationServer(grpcServer, server)

	log.Printf("auth server listening on %d\n", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}
