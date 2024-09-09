package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	egressserver "mediaserver-go/egress/servers"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/servers"

	"mediaserver-go/endpoints"
)

func main() {
	ctx := context.Background()

	hubManager := hubs.NewManager()

	webrtcServer, err := servers.NewWebRTC(hubManager)
	if err != nil {
		panic(err)
	}

	efs, err := egressserver.NewFileServer(hubManager)
	if err != nil {
		panic(err)
	}

	fileServer, err := servers.NewFileServer(hubManager)
	if err != nil {
		panic(err)
	}

	rtpServer, err := servers.NewRTPServer("0.0.0.0", 5000)
	if err != nil {
		panic(err)
	}

	go rtpServer.Run(ctx)
	e := endpoints.Initialize(&webrtcServer, &fileServer, &efs)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return e.Start(":8080")
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
