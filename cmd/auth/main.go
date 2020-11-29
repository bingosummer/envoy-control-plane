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
	flag.StringVar(&clusterName, "name", "cluster1", "cluster name")

	// The port that this auth server listens on
	flag.UintVar(&port, "port", 9002, "auth server port")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&watchDirectoryFileName, "watchDirectoryFileName", "/config", "full path to directory to watch for files")

	flag.StringVar(&configFile, "configFile", "config.yaml", "config file name to watch")
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
			file := path.Join(watchDirectoryFileName, configFile)
			if msg.FilePath != file {
				log.Println("skip", msg.FilePath)
				continue
			}
			log.Printf("process file %v", msg)
			srv.ParseConfig(msg.FilePath)
		}
	}
}
