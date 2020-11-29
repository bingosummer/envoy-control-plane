package main

import (
	"context"
	"flag"
	"log"
	"path"

	server "github.com/weinong/envoy-control-plane/internal/server/auth"
	"github.com/weinong/envoy-control-plane/internal/watcher"
)

var (
	clusterName            string
	port                   uint
	watchDirectoryFileName string
	configFile             string
)

func init() {
	// The port that this auth server listens on
	flag.UintVar(&port, "port", 9002, "auth server port")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&watchDirectoryFileName, "watchDirectoryFileName", "/config", "full path to directory to watch for files")

	flag.StringVar(&configFile, "configFile", "config.yaml", "config file name to watch")

	flag.StringVar(&clusterName, "clusterName", "cluster1", "cluster name that configuration will apply to")
}

func main() {
	flag.Parse()

	// Notify channel for file system events
	notifyCh := make(chan watcher.NotifyMessage)

	go func() {
		// Watch for file changes
		watcher.Watch(watchDirectoryFileName, notifyCh)
	}()

	srv := server.NewServer(clusterName)
	srv.ParseConfig(path.Join(watchDirectoryFileName, configFile))

	go func() {
		ctx := context.Background()
		server.Run(ctx, srv, port)
	}()

	for {
		select {
		case msg := <-notifyCh:
			if msg.Operation == watcher.Remove {
				continue
			}
			log.Printf("process file %v", msg)
			srv.ParseConfig(msg.FilePath)
		}
	}
}
