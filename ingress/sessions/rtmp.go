package sessions

import (
	"bytes"
	"fmt"
	pion "github.com/pion/webrtc/v3"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"
	"go.uber.org/zap"
	"io"
	"mediaserver-go/hubs"
	"mediaserver-go/hubs/codecs/aac"
	"mediaserver-go/hubs/codecs/factory"
	"mediaserver-go/hubs/codecs/h264"
	"mediaserver-go/parsers/format"
	"mediaserver-go/thirdparty/ffmpeg/avutil"
	"mediaserver-go/utils/log"
	"mediaserver-go/utils/units"
	"sync"
)

type RTMPSession struct {
	rtmp.DefaultHandler

	once      sync.Once
	streamKey string
	hub       *hubs.Hub
	stream    *hubs.Stream
	extraData h264.H264Config

	videoSource *hubs.HubSource
	audioSource *hubs.HubSource
	prevVideoTS uint32
	prevAudioTS uint32
}

func NewRTMPSession(hub *hubs.Hub) *RTMPSession {
	return &RTMPSession{
		hub: hub,
	}
}

func (h *RTMPSession) OnServe(conn *rtmp.Conn) {
	fmt.Println("OnServe")
}

func (h *RTMPSession) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	fmt.Printf("OnConnect cmd:%+v\n", *cmd)
	return nil
}

func (h *RTMPSession) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	fmt.Printf("OnCreateStream cmd:%+v\n", *cmd)
	return nil
}

func (h *RTMPSession) OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error {
	h.once.Do(func() {
		h.streamKey = cmd.StreamName
		h.stream = hubs.NewStream()
		h.hub.AddStream(h.streamKey, h.stream)
	})

	log.Logger.Debug("OnReleaseStream", zap.Any("cmd", cmd))
	return nil
}

func (h *RTMPSession) OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error {
	fmt.Println("OnDeleteStream")
	return nil
}

func (h *RTMPSession) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	h.once.Do(func() {
		h.streamKey = cmd.PublishingName
		h.stream = hubs.NewStream()
		h.hub.AddStream(h.streamKey, h.stream)
	})

	log.Logger.Debug("OnPublish", zap.Any("cmd", cmd))
	return nil
}

func (h *RTMPSession) OnPlay(_ *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPlay) error {
	fmt.Printf("OnPlay cmd:%+v\n", *cmd)
	return nil
}

func (h *RTMPSession) OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error {
	h.once.Do(func() {
		h.streamKey = cmd.StreamName
		h.stream = hubs.NewStream()
		h.hub.AddStream(h.streamKey, h.stream)
	})

	log.Logger.Debug("OnFCPublish", zap.Any("cmd", cmd))
	return nil
}

func (h *RTMPSession) OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error {
	fmt.Println("OnFCUnpublish")
	return nil
}

func (h *RTMPSession) OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error {
	r := bytes.NewReader(data.Payload)

	var script flvtag.ScriptData
	if err := flvtag.DecodeScriptData(r, &script); err != nil {
		log.Logger.Error("Failed to decode script data",
			zap.Error(err))
		return nil // ignore
	}

	for _, amf := range script.Objects {
		for key, v := range amf {
			fmt.Printf("key:%v, value:%v, type:%T\n", key, v, v)
			switch key {
			case "videocodecid":
				fv := v.(float64)
				value := flvtag.CodecID(fv)
				if value != flvtag.CodecIDAVC {
					return fmt.Errorf("unsupported video codec: %v", v)
				}
				typ, err := factory.NewType(pion.MimeTypeH264)
				if err != nil {
					return err
				}
				h.videoSource = hubs.NewHubSource(typ)
				h.stream.AddSource(h.videoSource)
			case "audiocodecid":
				fv := v.(float64)
				value := flvtag.SoundFormat(fv)
				if value != flvtag.SoundFormatAAC {
					return fmt.Errorf("unsupported audio codec: %v", v)
				}
				typ, err := factory.NewType("audio/aac")
				if err != nil {
					return err
				}
				h.audioSource = hubs.NewHubSource(typ)
				h.stream.AddSource(h.audioSource)
			}
		}
	}

	log.Logger.Info("SetDataFrame", zap.Any("script", script))

	return nil
}

func (h *RTMPSession) OnAudio(timestamp uint32, payload io.Reader) error {
	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}
	data, err := io.ReadAll(audio.Data)
	if err != nil {
		return err
	}
	switch audio.AACPacketType {
	case flvtag.AACPacketTypeSequenceHeader:
		fmt.Printf("audio:%+v\n", audio)
		fmt.Printf("data:%X\n", data)

		if audio.SoundFormat != flvtag.SoundFormatAAC {
			return fmt.Errorf("unsupported audio codec: %v", audio.SoundFormat)
		}

		config := format.AACConfig{}
		if err := config.ParseAACAudioSpecificConfig(data); err != nil {
			return err
		}
		codec := aac.NewAAC(aac.Parameters{
			SampleRate: config.SamplingRate,
			Channels:   config.Channel,
			SampleFmt:  int(avutil.AV_SAMPLE_FMT_FLTP),
		})
		h.audioSource.SetCodec(codec)
	case flvtag.AACPacketTypeRaw:
		duration := timestamp - h.prevAudioTS
		h.prevAudioTS = timestamp
		h.audioSource.Write(units.Unit{
			Payload:  data,
			PTS:      int64(timestamp),
			DTS:      int64(timestamp),
			Duration: int64(duration),
			TimeBase: 1000,
		})

	}
	return nil
}

func (h *RTMPSession) OnVideo(timestamp uint32, payload io.Reader) error {
	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	body, err := io.ReadAll(video.Data)
	if err != nil {
		return err
	}
	switch video.AVCPacketType {
	case flvtag.AVCPacketTypeSequenceHeader:
		if video.CodecID != flvtag.CodecIDAVC {
			return fmt.Errorf("unsupported video codec: %v", video.CodecID)
		}
		if err := h.extraData.UnmarshalFromConfig(body); err != nil {
			return err
		}
		h264Codecs, err := h264.NewH264(h.extraData.SPS, h.extraData.PPS)
		if err != nil {
			return err
		}
		h.videoSource.SetCodec(h264Codecs)
	case flvtag.AVCPacketTypeNALU:
		duration := timestamp - h.prevVideoTS
		h.prevVideoTS = timestamp
		for _, au := range format.GetAUFromAVC(body) {
			h.videoSource.Write(units.Unit{
				Payload:  au,
				PTS:      int64(timestamp),
				DTS:      int64(timestamp),
				Duration: int64(duration),
				TimeBase: 1000,
			})
		}
	case flvtag.AVCPacketTypeEOS:

	}
	return nil
}

func (h *RTMPSession) OnUnknownMessage(timestamp uint32, msg message.Message) error {
	fmt.Println("OnUnknownMessage")
	return nil
}

func (h *RTMPSession) OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error {
	fmt.Println("OnUnknownCommandMessage")
	return nil
}

func (h *RTMPSession) OnUnknownDataMessage(timestamp uint32, data *message.DataMessage) error {
	fmt.Println("OnUnknownDataMessage")
	return nil
}

func (h *RTMPSession) OnClose() {
	if h.streamKey == "" {
		return
	}
	h.stream.Close()
	h.hub.RemoveStream(h.streamKey)
}
