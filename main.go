package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"mediaserver-go/ingress/servers"

	"mediaserver-go/endpoints"
)

/*
// #cgo pkg-config: libavformat
// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L. -lmsvcrt -lhello
#include "hello.h"

// void printHello();
*/
import "C"

func main() {
	ctx := context.Background()

	webrtcServer, err := servers.NewWebRTC()
	if err != nil {
		panic(err)
	}
	e := endpoints.Initialize(&webrtcServer)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return e.Start(":8080")
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
