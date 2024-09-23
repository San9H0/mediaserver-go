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
	egressFileServer, err := egress.NewFileServer(hub)
	if err != nil {
		panic(err)
	}
	egressRTPServer, err := egress.NewRTPServer(hub)
	if err != nil {
		panic(err)
	}
	hlsServer, err := egress.NewHLSServer(hub)
	if err != nil {
		panic(err)
	}
	egressImageServer, err := egress.NewImageServer(hub)
	if err != nil {
		panic(err)
	}

	e := endpoints.Initialize(&whipServer, &fileServer, &whepServer, &egressFileServer, &ingressRTPServer, &egressRTPServer, &hlsServer, &egressImageServer)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		port := viper.GetInt("general.port")
		return e.Start(fmt.Sprintf("0.0.0.0:%d", port))
	})
	g.Go(func() error {
		rtmpServer, err := ingress.NewRTMPServer(hub)
		if err != nil {
			return err
		}
		port := viper.GetUint16("rtmp.port")
		return rtmpServer.Start(fmt.Sprintf("0.0.0.0:%d", port))
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
