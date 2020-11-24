package main

import (
	"context"
	"flag"
	"path"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"github.com/weinong/envoy-control-plane/internal/processor"
	server "github.com/weinong/envoy-control-plane/internal/server/xds"
	"github.com/weinong/envoy-control-plane/internal/watcher"
)

var (
	l                      log.FieldLogger
	watchDirectoryFileName string
	configFile             string
	port                   uint
	nodeID                 string
)

func init() {
	l = log.New()
	log.SetLevel(log.DebugLevel)

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 9002, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&watchDirectoryFileName, "watchDirectoryFileName", "/config", "full path to directory to watch for files")

	flag.StringVar(&configFile, "configFile", "config.yaml", "config file name to watch")
}

func main() {
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create a processor
	proc := processor.NewProcessor(
		cache, nodeID)

	// Create initial snapshot from file
	proc.ProcessFile(watcher.NotifyMessage{
		Operation: watcher.Create,
		FilePath:  path.Join(watchDirectoryFileName, configFile),
	})

	// Notify channel for file system events
	notifyCh := make(chan watcher.NotifyMessage)

	go func() {
		// Watch for file changes
		watcher.Watch(watchDirectoryFileName, notifyCh)
	}()

	go func() {
		// Run the xDS server
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, port)
	}()

	for {
		select {
		case msg := <-notifyCh:
			file := path.Join(watchDirectoryFileName, configFile)
			if msg.FilePath != file {
				log.Info("skip", msg.FilePath)
			}
			log.Infof("process file %v", msg)
			proc.ProcessFile(msg)
		}
	}
}
