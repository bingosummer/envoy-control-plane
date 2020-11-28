package xdscache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/weinong/envoy-control-plane/internal/resources"
)

type XDSCache struct {
	Listeners map[string]resources.Listener
	Routes    map[string]resources.Route
	Clusters  map[string]resources.Cluster
	Endpoints map[string]resources.Endpoint
}

func (xds *XDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeCluster(c.Name, c.IsHTTPS))
	}

	return r
}

func (xds *XDSCache) RouteContents() []types.Resource {

	var routesArray []resources.Route
	for _, r := range xds.Routes {
		routesArray = append(routesArray, r)
	}

	return []types.Resource{resources.MakeRoute(routesArray)}
}

func (xds *XDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.RouteNames[0], l.Address, l.Port))
	}

	return r
}

func (xds *XDSCache) EndpointsContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeEndpoint(c.Name, c.Endpoints))
	}

	return r
}

func (xds *XDSCache) AddListener(name string, routeNames []string, address string, port uint32) {
	xds.Listeners[name] = resources.Listener{
		Name:       name,
		Address:    address,
		Port:       port,
		RouteNames: routeNames,
	}
}

func (xds *XDSCache) AddRoute(name, prefix string, header string, cluster string, hostRewrite string) {
	xds.Routes[name] = resources.Route{
		Name:        name,
		Prefix:      prefix,
		Header:      header,
		Cluster:     cluster,
		HostRewrite: hostRewrite,
	}
}

func (xds *XDSCache) AddCluster(name string, isHTTPS bool) {
	xds.Clusters[name] = resources.Cluster{
		Name:    name,
		IsHTTPS: isHTTPS,
	}
}

func (xds *XDSCache) AddEndpoint(clusterName, upstreamHost string, upstreamPort uint32) {
	cluster := xds.Clusters[clusterName]

	cluster.Endpoints = append(cluster.Endpoints, resources.Endpoint{
		UpstreamHost: upstreamHost,
		UpstreamPort: upstreamPort,
	})

	xds.Clusters[clusterName] = cluster
}
