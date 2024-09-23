package servers

import (
	"context"
	"errors"
	"fmt"
	"github.com/bluenviron/gohlslib/pkg/playlist"
	_ "github.com/bluenviron/gohlslib/pkg/playlist"
	"go.uber.org/zap"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/hls"
	"mediaserver-go/hubs"
	"mediaserver-go/utils"
	"mediaserver-go/utils/buffers"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/types"
	"strconv"
	"sync"
	"time"
)

type HLSServer struct {
	mu sync.RWMutex

	hub        *hubs.Hub
	hlsStreams map[string]*HLSHandler
}

func NewHLSServer(hub *hubs.Hub) (HLSServer, error) {
	return HLSServer{
		hub:        hub,
		hlsStreams: make(map[string]*HLSHandler),
	}, nil
}

func (h *HLSServer) StartSession(streamID string, req dto.HLSRequest) (dto.HLSResponse, error) {
	stream, ok := h.hub.GetStream(streamID)
	fmt.Println("[TESTDEBUG] StartSession.. ok:", ok, " streamID:", streamID)
	if !ok {
		return dto.HLSResponse{}, errors.New("stream not found")
	}

	hlsStream := newHLSStream()

	handler := hls.NewHandler(buffers.NewMemory(), hlsStream)
	if err := handler.Init(context.Background(), stream.Sources()); err != nil {
		return dto.HLSResponse{}, err
	}

	video := handler.CodecString(types.MediaTypeVideo)
	audio := handler.CodecString(types.MediaTypeAudio)

	version := 10
	master := playlist.Multivariant{
		Version:             version,
		IndependentSegments: true,
		Variants: []*playlist.MultivariantVariant{
			{
				URI:        "video.m3u8",
				Bandwidth:  1_000_000,
				Codecs:     []string{video, audio},
				Resolution: "1280x720",
				FrameRate:  utils.GetPointer(29.970),
			},
		},
	}

	llhlsmedia := playlist.Media{
		Version:             version,
		IndependentSegments: true,
		TargetDuration:      2,
		ServerControl: &playlist.MediaServerControl{
			CanBlockReload: true,
			PartHoldBack:   utils.GetPointer(time.Duration(3051) * time.Millisecond),
		},
		MediaSequence: 0,
		//DiscontinuitySequence: utils.GetPointer(0),
		Map:     &playlist.MediaMap{URI: "init.mp4"},
		PartInf: &playlist.MediaPartInf{PartTarget: time.Second},
	}

	hlsmedia := playlist.Media{
		Version:             version,
		IndependentSegments: true,
		TargetDuration:      2,
		MediaSequence:       0,
		Map:                 &playlist.MediaMap{URI: "init.mp4"},
	}

	hlsStream.master = &master
	hlsStream.hlsmedia = &hlsmedia
	hlsStream.llhlsMedia = &llhlsmedia
	h.mu.Lock()
	h.hlsStreams[streamID] = hlsStream
	h.mu.Unlock()

	sess := sessions.NewSession[*hls.OnTrackContext](handler)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sess.Run(ctx)
	}()

	return dto.HLSResponse{}, nil
}

func (h *HLSServer) GetHLSStream(streamID string) (*HLSHandler, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	stream, ok := h.hlsStreams[streamID]
	if !ok {
		return nil, errors.New("stream not found")
	}
	return stream, nil
}

type HLSHandler struct {
	mu sync.RWMutex

	master *playlist.Multivariant

	hlsmedia        *playlist.Media
	hlsTempPayload  []byte
	hlsTempDuration float64

	llhlsMedia *playlist.Media

	lastSN       int
	mediaPayload map[string]*Media
}

func (h *HLSHandler) loadOrStoreMedia(key string) *Media {
	media, ok := h.mediaPayload[key]
	if !ok {
		media = NewMedia()
		h.mediaPayload[key] = media
	}
	return media
}

type Media struct {
	closeCh chan struct{}
	payload []byte
}

func NewMedia() *Media {
	return &Media{
		closeCh: make(chan struct{}),
	}
}

func newHLSStream() *HLSHandler {
	return &HLSHandler{
		mediaPayload: make(map[string]*Media),
	}
}

func (h *HLSHandler) GetMasterM3U8() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.master.Marshal()
}

func (h *HLSHandler) getMediaM3U8LLHLS(sn int) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.llhlsMedia.MediaSequence > sn {
		return nil, errors.New("no segments found")
	}
	for _, sg := range h.llhlsMedia.Segments {
		if sg.Title == strconv.Itoa(sn) {
			fmt.Println("[TESTDEBUG] sg.Title:", sg.Title, ", strconv.Itoa(sn):", strconv.Itoa(sn))
			return h.llhlsMedia.Marshal()
		}
	}
	return nil, nil
}

func (h *HLSHandler) GetMediaM3U8HLS() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.hlsmedia.Marshal()
}

func (h *HLSHandler) GetMediaM3U8LLHLS(sn, part string) ([]byte, error) {
	if sn == "" {
		h.mu.RLock()
		defer h.mu.RUnlock()
		return h.llhlsMedia.Marshal()
	}

	sequenceNumber, err := strconv.Atoi(sn)
	if err != nil {
		return nil, err
	}

	b, err := h.getMediaM3U8LLHLS(sequenceNumber)
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		return b, nil
	}

	output := fmt.Sprintf("output_%s.m4s", sn)
	if part != "" {
		output = fmt.Sprintf("output_%s_%s.m4s", sn, part)
	}

	h.mu.Lock()
	media := h.loadOrStoreMedia(output)
	h.mu.Unlock()

	select {
	case <-time.After(4 * time.Second):
		return nil, errors.New("timeout")
	case <-media.closeCh:
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.llhlsMedia.Marshal()
}

func (h *HLSHandler) GetPayload(name string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	media, ok := h.mediaPayload[name]
	if !ok {
		return nil, errors.New("media not found")
	}
	return media.payload, nil
}

func (h *HLSHandler) SetPayload(payload []byte, name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	media := h.loadOrStoreMedia(name)
	close(media.closeCh)
	media.payload = payload
}

func (h *HLSHandler) appendMediaHLS(payload []byte, index, segIndex, partIndex int, duration float64) *playlist.MediaSegment {
	h.mu.Lock()
	defer h.mu.Unlock()

	var deleted *playlist.MediaSegment

	segOutput := fmt.Sprintf("output_%d.m4s", segIndex)

	if partIndex == 0 {
		h.hlsTempPayload = payload
		h.hlsTempDuration = duration
		return nil
	}
	if len(h.hlsmedia.Segments) >= 3 {
		h.hlsmedia.MediaSequence += 1
		deleted = h.hlsmedia.Segments[0]
		h.hlsmedia.Segments = h.hlsmedia.Segments[1:]
	}
	h.hlsmedia.Segments = append(h.hlsmedia.Segments, &playlist.MediaSegment{
		URI:      segOutput,
		Duration: time.Duration(duration*1_000_000)*time.Microsecond + time.Duration(h.hlsTempDuration*1_000_000)*time.Microsecond,
		Title:    strconv.Itoa(segIndex),
		DateTime: utils.GetPointer(time.Now().UTC()),
	})

	media := h.loadOrStoreMedia(segOutput)
	media.payload = append(h.hlsTempPayload, payload...)
	close(media.closeCh)

	log.Logger.Info("hls Handler write file end",
		zap.Int("index", index),
		zap.Int("segIndex", segIndex),
		zap.Int("partIndex", partIndex),
		zap.Int("size", len(payload)))

	return deleted
}

func (h *HLSHandler) appendMediallhls(payload []byte, index, segIndex, partIndex int, duration float64) *playlist.MediaSegment {
	h.mu.Lock()
	defer h.mu.Unlock()

	var deleted *playlist.MediaSegment

	if partIndex == 0 {
		part0Output := fmt.Sprintf("output_%d_%d.m4s", segIndex, partIndex)
		nextOutput := fmt.Sprintf("output_%d_%d.m4s", segIndex, 1)
		h.llhlsMedia.Parts = []*playlist.MediaPart{
			{
				Duration:    time.Duration(duration*1_000_000) * time.Microsecond,
				URI:         part0Output,
				Independent: true,
			},
		}
		h.llhlsMedia.PreloadHint = &playlist.MediaPreloadHint{
			URI: nextOutput,
		}

		media := h.loadOrStoreMedia(part0Output)
		media.payload = payload
		close(media.closeCh)
	} else {
		part1Output := fmt.Sprintf("output_%d_%d.m4s", segIndex, partIndex)
		nextOutput := fmt.Sprintf("output_%d_%d.m4s", segIndex+1, 0)

		if len(h.hlsmedia.Segments) > 0 {
			if len(h.llhlsMedia.Segments) >= 3 {
				h.llhlsMedia.MediaSequence += 1
				deleted = h.llhlsMedia.Segments[0]
				h.llhlsMedia.Segments = h.llhlsMedia.Segments[1:]
			}
			lastSeg := h.hlsmedia.Segments[len(h.hlsmedia.Segments)-1]
			h.llhlsMedia.Segments = append(h.llhlsMedia.Segments, &playlist.MediaSegment{
				URI:      lastSeg.URI,
				Duration: lastSeg.Duration,
				Title:    lastSeg.Title,
				DateTime: lastSeg.DateTime,
				Parts: append(h.llhlsMedia.Parts, &playlist.MediaPart{
					Duration:    time.Duration(duration*1_000_000) * time.Microsecond,
					URI:         part1Output,
					Independent: true,
				}),
			})
		}
		h.llhlsMedia.Parts = nil

		h.llhlsMedia.PreloadHint = &playlist.MediaPreloadHint{
			URI: nextOutput,
		}

		media1 := h.loadOrStoreMedia(part1Output)
		media1.payload = payload
		close(media1.closeCh)
	}

	//log.Logger.Info("hls Handler write file end",
	//	zap.Int("index", index),
	//	zap.Int("segIndex", segIndex),
	//	zap.Int("partIndex", partIndex),
	//	zap.Int("size", len(payload)))

	return deleted
}

func (h *HLSHandler) AppendMedia(payload []byte, index int, duration float64) {
	segIndex := index / 2
	partIndex := index % 2

	deleted := h.appendMediaHLS(payload, index, segIndex, partIndex, duration)

	h.appendMediallhls(payload, index, segIndex, partIndex, duration)

	if deleted == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.mediaPayload, deleted.Title)
}
