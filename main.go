package main

import (
	"context"
	"fmt"
	"github.com/pion/ice/v2"
	pion "github.com/pion/webrtc/v3"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	egress "mediaserver-go/egress/servers"
	"mediaserver-go/endpoints"
	"mediaserver-go/hubs"
	ingress "mediaserver-go/ingress/servers"
	"mediaserver-go/utils/configs"
	"mediaserver-go/utils/log"
)

func main() {
	if err := configs.Init(); err != nil {
		panic(err)
	}

	if err := log.Init(); err != nil {
		panic(err)
	}

	log.Logger.Info("Starting server")

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

	whipServer, err := ingress.NewWHIP(hub, se)
	if err != nil {
		panic(err)
	}
	fileServer, err := ingress.NewFileServer(hub)
	if err != nil {
		panic(err)
	}
	ingressRTPServer, err := ingress.NewRTPServer(hub)
	if err != nil {
		panic(err)
	}

	whepServer, err := egress.NewWHEP(hub, se)
	if err != nil {
		panic(err)
	}
	efs, err := egress.NewFileServer(hub)
	if err != nil {
		panic(err)
	}
	egressRTPServer, err := egress.NewRTPServer(hub)
	if err != nil {
		panic(err)
	}
	go func() {
		rtmpServer, err := ingress.NewRTMPServer(hub)
		_ = rtmpServer
		if err != nil {
			panic(err)
		}

	}()
	e := endpoints.Initialize(&whipServer, &fileServer, &whepServer, &efs, &ingressRTPServer, &egressRTPServer)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return e.Start(fmt.Sprintf(":%d", viper.GetInt("general.port")))
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
