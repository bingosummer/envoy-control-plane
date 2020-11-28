package processor

import (
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/weinong/envoy-control-plane/internal/resources"
	"github.com/weinong/envoy-control-plane/internal/utils"
	"github.com/weinong/envoy-control-plane/internal/watcher"
	"github.com/weinong/envoy-control-plane/internal/xdscache"
)

type Processor struct {
	cache  cache.SnapshotCache
	nodeID string

	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64

	xdsCache xdscache.XDSCache
}

func NewProcessor(cache cache.SnapshotCache, nodeID string) *Processor {
	return &Processor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		xdsCache: xdscache.XDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
	}
}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (p *Processor) newSnapshotVersion() string {

	// Reset the snapshotVersion if it ever hits max size.
	if p.snapshotVersion == math.MaxInt64 {
		p.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	p.snapshotVersion++
	return strconv.FormatInt(p.snapshotVersion, 10)
}

// ProcessFile takes a file and generates an xDS snapshot
func (p *Processor) ProcessFile(file watcher.NotifyMessage) {

	// Parse file into object
	envoyConfig, err := utils.ParseEnvoyConfig(file.FilePath)
	if err != nil {
		log.Printf("error parsing yaml file: %s", err)
		return
	}

	// Parse Listeners
	for _, l := range envoyConfig.Listeners {
		var lRoutes []string
		for _, lr := range l.Routes {
			lRoutes = append(lRoutes, lr.Name)
		}

		p.xdsCache.AddListener(l.Name, lRoutes, l.Address, l.Port)

		for _, r := range l.Routes {
			p.xdsCache.AddRoute(r.Name, r.Prefix, r.Header, r.Cluster, r.HostRewrite)
		}
	}

	// Parse Clusters
	for _, c := range envoyConfig.Clusters {
		p.xdsCache.AddCluster(c.Name, c.IsHTTPS)

		// Parse endpoints
		for _, e := range c.Endpoints {
			p.xdsCache.AddEndpoint(c.Name, e.Address, e.Port)
		}
	}

	// Create the snapshot that we'll serve to Envoy
	snapshot := cache.NewSnapshot(
		p.newSnapshotVersion(),         // version
		p.xdsCache.EndpointsContents(), // endpoints
		p.xdsCache.ClusterContents(),   // clusters
		p.xdsCache.RouteContents(),     // routes
		p.xdsCache.ListenerContents(),  // listeners
		[]types.Resource{},             // runtimes
		[]types.Resource{},             // secrets
	)

	if err := snapshot.Consistent(); err != nil {
		log.Printf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		return
	}
	log.Printf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := p.cache.SetSnapshot(p.nodeID, snapshot); err != nil {
		log.Printf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}
