package main

import (
	"context"
	"github.com/pion/ice/v2"
	pion "github.com/pion/webrtc/v3"
	"golang.org/x/sync/errgroup"
	egressserver "mediaserver-go/egress/servers"
	"mediaserver-go/hubs"
	"mediaserver-go/ingress/servers"

	"mediaserver-go/endpoints"
)

func main() {
	ctx := context.Background()

	hub := hubs.NewHub()

	se := pion.SettingEngine{}
	if err := se.SetEphemeralUDPPortRange(10000, 20000); err != nil {
		panic(err)
	}
	se.SetIncludeLoopbackCandidate(true)
	se.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
	//se.SetNAT1To1IPs([]string{"127.0.0.1"}, webrtc.ICECandidateTypeHost)
	se.SetLite(true)

	whipServer, err := servers.NewWHIP(hub, se)
	if err != nil {
		panic(err)
	}

	whepServer, err := egressserver.NewWHEP(hub, se)
	if err != nil {
		panic(err)
	}

	efs, err := egressserver.NewFileServer(hub)
	if err != nil {
		panic(err)
	}

	fileServer, err := servers.NewFileServer(hub)
	if err != nil {
		panic(err)
	}

	egressRTPServer, err := egressserver.NewRTPServer(hub)
	if err != nil {
		panic(err)
	}

	ingressRTPServer, err := servers.NewRTPServer(hub)
	if err != nil {
		panic(err)
	}

	e := endpoints.Initialize(&whipServer, &fileServer, &whepServer, &efs, &ingressRTPServer, &egressRTPServer)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return e.Start(":8080")
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
