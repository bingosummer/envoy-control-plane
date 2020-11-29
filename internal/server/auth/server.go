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
	ClusterName string
	*v1alpha1.EnvoyConfig
}

func (s *Server) ParseConfig(file string) {
	config, err := utils.ParseEnvoyConfig(file)
	if err != nil {
		log.Printf("unable to read ext auth config: %s", err)
		return
	}

	if config.Name != s.ClusterName {
		log.Printf("skip config from %s", file)
		return
	}
	log.Printf("incoming external auth config is %+v", config)
	s.EnvoyConfig = config
}

func (s *Server) Check(ctx context.Context, req *authservice.CheckRequest) (*authservice.CheckResponse, error) {

	if s.EnvoyConfig == nil {
		log.Printf("ext-authz is not configured. access is denied")
		return &authservice.CheckResponse{
			Status: &status.Status{Code: int32(rpc.PERMISSION_DENIED)},
		}, nil
	}

	token := strings.TrimPrefix(req.Attributes.Request.Http.Headers["authorization"], "Bearer ")
	targetCluster := req.Attributes.Request.Http.Headers["x-route"]
	log.Printf("x-route: %s, bearer token: %s", targetCluster, token)

	var desiredRoute *v1alpha1.ExtAuthzRoute

	for _, r := range s.ExtAuthz.Routes {
		if r.Cluster == targetCluster {
			desiredRoute = &r
			break
		}
	}
	if desiredRoute == nil || desiredRoute.RequiredToken != token {
		log.Printf("no route is configured for %s or token is mismatched, desired route: %+v", targetCluster, desiredRoute)
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
				Value: targetCluster,
			},
		},
	}
	if desiredRoute.RewriteHost != "" {
		headers = append(headers, &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   ":authority",
				Value: desiredRoute.RewriteHost,
			},
		})
	}
	if desiredRoute.OutgoingToken != "" {
		headers = append(headers, &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   "Authorization",
				Value: fmt.Sprintf("Bearer %s", desiredRoute.OutgoingToken),
			},
		})
	}

	for k, v := range desiredRoute.AdditionalHeaders {
		headers = append(headers, &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   k,
				Value: v,
			},
		})
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

func NewServer(name string) *Server {
	return &Server{ClusterName: name}
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
