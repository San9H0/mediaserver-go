package servers

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafov/m3u8"
	"mediaserver-go/egress/sessions"
	"mediaserver-go/egress/sessions/hls"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/buffers"
	"mediaserver-go/utils/dto"
	"mediaserver-go/utils/types"
	"sync"
	"time"
)

type HLSServer struct {
	mu sync.RWMutex

	hub        *hubs.Hub
	hlsStreams map[string]*HLSStreamHandle
}

func NewHLSServer(hub *hubs.Hub) (HLSServer, error) {
	return HLSServer{
		hub:        hub,
		hlsStreams: make(map[string]*HLSStreamHandle),
	}, nil
}

func (h *HLSServer) StartSession(streamID string, req dto.HLSRequest) (dto.HLSResponse, error) {
	stream, ok := h.hub.GetStream(streamID)
	if !ok {
		return dto.HLSResponse{}, errors.New("stream not found")
	}

	hlsStream := newHLSStream()

	handler := hls.NewHandler(buffers.NewMemory(), hlsStream)
	if err := handler.Init(context.Background(), stream.Tracks()); err != nil {
		return dto.HLSResponse{}, err
	}

	video := handler.CodecString(types.MediaTypeVideo)
	audio := handler.CodecString(types.MediaTypeAudio)

	playlist := m3u8.NewMasterPlaylist()
	mediaPlayList, err := m3u8.NewMediaPlaylist(3, 3)
	if err != nil {
		return dto.HLSResponse{}, err
	}
	mediaPlayList.SetDefaultMap("init.mp4", 0, 0)

	for _, track := range handler.NegotiatedTracks() {
		if track.MediaType() == types.MediaTypeAudio {
			continue
		}
		playlist.Append("video.m3u8", mediaPlayList, m3u8.VariantParams{
			Bandwidth:  1_000_000,
			Resolution: "1280x720",
			FrameRate:  29.970,
			Codecs:     fmt.Sprintf("%s,%s", video, audio),
		})
	}

	hlsStream.master = playlist
	hlsStream.media = mediaPlayList
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

func (h *HLSServer) GetHLSStream(streamID string) (*HLSStreamHandle, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	stream, ok := h.hlsStreams[streamID]
	if !ok {
		return nil, errors.New("stream not found")
	}
	return stream, nil
}

func (h *HLSServer) GetMasterM3U8(streamID string) (string, error) {
	h.mu.RLock()
	stream, ok := h.hlsStreams[streamID]
	h.mu.RUnlock()
	if !ok {
		return "", errors.New("stream not found")
	}
	return stream.GetMasterM3U8(), nil
}

func (h *HLSServer) GetMediaM3U8(streamID string) (string, error) {
	h.mu.RLock()
	stream, ok := h.hlsStreams[streamID]
	h.mu.RUnlock()
	if !ok {
		return "", errors.New("stream not found")
	}
	return stream.GetMediaM3U8(), nil
}

type HLSStreamHandle struct {
	mu sync.RWMutex

	master       *m3u8.MasterPlaylist
	media        *m3u8.MediaPlaylist
	mediaPayload map[string][]byte
}

func newHLSStream() *HLSStreamHandle {
	return &HLSStreamHandle{
		mediaPayload: make(map[string][]byte),
	}
}

func (h *HLSStreamHandle) GetMasterM3U8() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.master.String()
}

func (h *HLSStreamHandle) GetMediaM3U8() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.media.String()
}

func (h *HLSStreamHandle) GetInitFile() []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.mediaPayload["init.mp4"]
}

func (h *HLSStreamHandle) GetPayload(name string) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	b, ok := h.mediaPayload[name]
	if !ok {
		return nil, errors.New("media not found")
	}
	return b, nil
}

func (h *HLSStreamHandle) SetPayload(payload []byte, name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.mediaPayload[name] = payload
}

func (h *HLSStreamHandle) AppendMedia(payload []byte, name string, duration float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.media.Slide(name, duration, "")
	h.media.SetProgramDateTime(time.Now().UTC())
	h.mediaPayload[name] = payload
}
