package resources

import (
	"time"

	"github.com/golang/protobuf/ptypes"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_file_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	extauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/weinong/envoy-control-plane/internal/utils"
)

const (
	UpstreamHost = "www.envoyproxy.io"
	UpstreamPort = 80
)

type Listener struct {
	Name       string
	Address    string
	Port       uint32
	RouteNames []string
	CertFile   string
	KeyFile    string
}

type Route struct {
	Name        string
	Prefix      string
	Header      string
	Cluster     string
	HostRewrite string
}

type Cluster struct {
	Name      string
	IsHTTPS   bool
	Endpoints []Endpoint
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}

func MakeCluster(clusterName string, isHTTPS bool) *cluster.Cluster {
	c := &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		//LoadAssignment:       makeEndpoint(clusterName, UpstreamHost),
		DnsLookupFamily:  cluster.Cluster_V4_ONLY,
		EdsClusterConfig: makeEDSCluster(),
	}
	if isHTTPS {
		c.TransportSocket = &core.TransportSocket{
			Name: wellknown.TransportSocketTls,
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: utils.MustMarshalAny(&envoy_tls_v3.UpstreamTlsContext{}),
			},
		}
	}
	return c
}

func makeEDSCluster() *cluster.Cluster_EdsClusterConfig {
	return &cluster.Cluster_EdsClusterConfig{
		EdsConfig: makeConfigSource(),
	}
}

func MakeEndpoint(clusterName string, eps []Endpoint) *endpoint.ClusterLoadAssignment {
	var endpoints []*endpoint.LbEndpoint

	for _, e := range eps {
		endpoints = append(endpoints, &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  e.UpstreamHost,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: e.UpstreamPort,
								},
							},
						},
					},
				},
			},
		})
	}

	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: endpoints,
		}},
	}
}

func MakeRoute(routes []Route) *route.RouteConfiguration {
	var rts []*route.Route

	for _, r := range routes {
		action := &route.Route_Route{}
		if r.Header != "" {
			action.Route = &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_ClusterHeader{
					ClusterHeader: r.Header,
				},
			}
		} else {
			action.Route = &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: r.Cluster,
				},
			}
		}
		if r.HostRewrite != "" {
			action.Route.HostRewriteSpecifier = &route.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: r.HostRewrite,
			}
		}
		rts = append(rts, &route.Route{
			//Name: r.Name,
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{Prefix: r.Prefix},
			},
			Action: action,
		})
	}

	return &route.RouteConfiguration{
		Name: "listener_0",
		VirtualHosts: []*route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
			Routes:  rts,
		}},
	}
}

func MakeHTTPListener(listenerName, route, address string, port uint32, certFile, keyFile string) *listener.Listener {
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: "listener_0",
			},
		},
		AccessLog: []*accesslog.AccessLog{
			{
				Name: wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: utils.MustMarshalAny(&envoy_file_v3.FileAccessLog{
						Path: "/dev/stdout",
					}),
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.HTTPExternalAuthorization,
				ConfigType: &hcm.HttpFilter_TypedConfig{
					TypedConfig: utils.MustMarshalAny(&extauth.ExtAuthz{
						Services: &extauth.ExtAuthz_GrpcService{
							GrpcService: &core.GrpcService{
								TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "ext-auth"},
								},
							},
						},
						TransportApiVersion: core.ApiVersion_V3}),
				},
			},
			{
				Name: wellknown.Router,
			},
		},
	}

	l := &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: utils.MustMarshalAny(manager),
				},
			}},
		}},
	}

	if certFile != "" && keyFile != "" {
		l.FilterChains[0].TransportSocket = &core.TransportSocket{
			Name: wellknown.TransportSocketTls,
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: utils.MustMarshalAny(&envoy_tls_v3.DownstreamTlsContext{
					CommonTlsContext: &envoy_tls_v3.CommonTlsContext{
						TlsCertificates: []*envoy_tls_v3.TlsCertificate{
							{
								PrivateKey: &core.DataSource{
									Specifier: &core.DataSource_Filename{Filename: keyFile},
								},
								CertificateChain: &core.DataSource{
									Specifier: &core.DataSource_Filename{Filename: certFile},
								},
							},
						},
					},
				}),
			},
		}
	}

	return l
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
