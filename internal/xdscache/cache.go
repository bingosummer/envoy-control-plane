package xdscache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/weinong/envoy-control-plane/apis/v1alpha1"
	"github.com/weinong/envoy-control-plane/internal/resources"
)

type XDSCache struct {
	Listeners map[string]resources.Listener
	Routes    map[string]resources.Route
	Clusters  map[string]resources.Cluster
	Endpoints map[string]resources.Endpoint
	RouteKey  string
}

func (xds *XDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, c.MakeCluster())
	}

	return r
}

func (xds *XDSCache) RouteContents() []types.Resource {

	var routesArray []resources.Route
	for _, r := range xds.Routes {
		routesArray = append(routesArray, r)
	}

	return []types.Resource{resources.MakeRoute(xds.RouteKey, routesArray)}
}

func (xds *XDSCache) ListenerContents() []types.Resource {
	var r []types.Resource

	for _, l := range xds.Listeners {
		r = append(r, resources.MakeHTTPListener(l.Name, l.RouteNames[0], l.Address, l.Port, l.CertFile, l.KeyFile))
	}

	return r
}

func (xds *XDSCache) AddListener(name string, routeNames []string, address string, port uint32, certFile, keyFile string) {
	xds.Listeners[name] = resources.Listener{
		Name:       name,
		Address:    address,
		Port:       port,
		RouteNames: routeNames,
		CertFile:   certFile,
		KeyFile:    keyFile,
	}
}

func (xds *XDSCache) AddRoute(name, prefix string, header string, hostRewrite string) {
	xds.Routes[name] = resources.Route{
		Name:        name,
		Prefix:      prefix,
		Header:      header,
		HostRewrite: hostRewrite,
	}
}

func (xds *XDSCache) AddCluster(cluster v1alpha1.Cluster) {
	c := resources.Cluster{
		Name:          cluster.Name,
		IsHTTPS:       cluster.IsHTTPS,
		DiscoveryType: string(cluster.DiscoveryType),
	}

	for _, v := range cluster.Endpoints {
		c.Endpoints = append(c.Endpoints, resources.Endpoint{
			UpstreamHost: v.Address,
			UpstreamPort: v.Port,
		})
	}

	xds.Clusters[cluster.Name] = c
}
