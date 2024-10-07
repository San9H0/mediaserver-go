package whep

import (
	"context"
	pioncodec "github.com/pion/rtp/codecs"
	"go.uber.org/zap"
	"mediaserver-go/codecs/av1"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"mediaserver-go/utils/units"
	"sync/atomic"
	"time"
)

type ABSHandler struct {
	stats *Stats

	maxSpatialLayer     atomic.Int32
	targetSpatialLayer  atomic.Int32
	currentSpatialLayer atomic.Int32

	maxTemporalLayer    atomic.Int32
	targetTemporalLayer atomic.Int32
}

func NewABSHandler(stats *Stats) *ABSHandler {
	return &ABSHandler{
		stats: stats,
	}
}

func (a *ABSHandler) Run(ctx context.Context) {
	prevSendCount := uint32(0)
	prevNackCount := uint32(0)
	noNackDuration := 0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendCount := a.stats.sendCount.Load()
			nackCount := a.stats.nackCount.Load()
			sendCountPerInterval := sendCount - prevSendCount
			nackCountPerInterval := nackCount - prevNackCount
			prevSendCount = sendCount
			prevNackCount = nackCount

			if nackCountPerInterval == 0 && sendCountPerInterval > 0 {
				noNackDuration++
			} else {
				noNackDuration = 0
			}

			if noNackDuration > 5 {
				a.upgradeLayer()
				noNackDuration = 0
			}
		}
	}
}

func (a *ABSHandler) SetMaxSpatialLayer(rid string) {
	maxSpatialLayer := int32(0)
	if rid == "1" {
		maxSpatialLayer = 1
	} else if rid == "2" {
		maxSpatialLayer = 2
	}

	for {
		currentMaxSpatialLayer := a.maxSpatialLayer.Load()
		if currentMaxSpatialLayer < maxSpatialLayer {
			if !a.maxSpatialLayer.CompareAndSwap(currentMaxSpatialLayer, maxSpatialLayer) {
				continue
			}
		}
		break
	}
}

func (a *ABSHandler) SetMaxTemporalLayer(tid uint8) {
	maxTemporalLayer := int32(tid)

	for {
		currentMaxTemporalLayer := a.maxTemporalLayer.Load()
		if currentMaxTemporalLayer < maxTemporalLayer {
			if !a.maxTemporalLayer.CompareAndSwap(currentMaxTemporalLayer, maxTemporalLayer) {
				continue
			}
		}
		break
	}
}

func (a *ABSHandler) upgradeLayer() {
	maxSpatialLayer := a.maxSpatialLayer.Load()
	targetSpatialLayer := a.targetSpatialLayer.Load()
	maxTemporalLayer := a.maxTemporalLayer.Load()
	targetTemporalLayer := a.targetTemporalLayer.Load()
	if targetTemporalLayer < maxTemporalLayer {
		targetTemporalLayer = a.targetTemporalLayer.Add(1)

		log.Logger.Info("upgrade layer",
			zap.String("type", "temporal"),
			zap.Int("targetTemporalLayer", int(targetTemporalLayer)),
			zap.Int("maxTemporalLayer", int(maxTemporalLayer)),
			zap.Int("targetSpatialLayer", int(targetSpatialLayer)),
			zap.Int("maxSpatialLayer", int(maxSpatialLayer)),
		)
		return
	}

	if targetSpatialLayer >= maxSpatialLayer {
		return
	}

	targetSpatialLayer = a.targetSpatialLayer.Add(1)
	a.maxTemporalLayer.Store(0)
	a.targetTemporalLayer.Store(0)

	log.Logger.Info("upgrade layer",
		zap.String("type", "spatial"),
		zap.Int("targetTemporalLayer", int(a.targetTemporalLayer.Load())),
		zap.Int("maxTemporalLayer", int(a.maxTemporalLayer.Load())),
		zap.Int("targetSpatialLayer", int(targetSpatialLayer)),
		zap.Int("maxSpatialLayer", int(maxSpatialLayer)),
	)
}

func (a *ABSHandler) CanSendSpatialLayer(rid string, unit units.Unit) bool {
	if a.isCurrentSpatialLayer(rid) {
		return true
	}
	if a.isTargetSpatialLayer(rid) && unit.FrameInfo.Flag == 1 {
		a.SetMaxTemporalLayer(0)
		a.setCurrentSpatialLayer(rid)
		return true
	}
	return false
}

func (a *ABSHandler) CanSendTemporalLayer(track hubs.Track, unit units.Unit) bool {
	codec := track.GetCodec()
	switch codec.CodecType() {
	case types.CodecTypeVP8:
		vp8Packet, ok := unit.FrameInfo.PayloadHeader.(*pioncodec.VP8Packet)
		if !ok {
			return true
		}
		a.SetMaxTemporalLayer(vp8Packet.TID)
		return a.isTragetTemporalLayer(vp8Packet.TID)
	case types.CodecTypeAV1:
		var header av1.OBUHeader
		if err := header.Unmarshal(unit.Payload); err != nil {
			return false
		}
		if !header.HasExtensionFlag {
			return true
		}
		a.SetMaxTemporalLayer(header.TemporalID)
		return a.isTragetTemporalLayer(header.TemporalID)
	default:
		return true
	}
}

func isLayerMatch(layer int32, rid string) bool {
	switch layer {
	case 0:
		if rid != "" && rid != "0" {
			return false
		}
	case 1:
		if rid != "1" {
			return false
		}
	case 2:
		if rid != "2" {
			return false
		}
	}
	return true
}

func (a *ABSHandler) isCurrentSpatialLayer(rid string) bool {
	return isLayerMatch(a.currentSpatialLayer.Load(), rid)
}

func (a *ABSHandler) isTargetSpatialLayer(rid string) bool {
	return isLayerMatch(a.targetSpatialLayer.Load(), rid)
}

func (a *ABSHandler) isTragetTemporalLayer(tid uint8) bool {
	return a.targetTemporalLayer.Load() >= int32(tid)
}

func (a *ABSHandler) setCurrentSpatialLayer(rid string) {
	currentSpatialLayer := int32(0)
	if rid == "1" {
		currentSpatialLayer = 1
	} else {
		currentSpatialLayer = 2
	}
	a.currentSpatialLayer.Store(currentSpatialLayer)
}
