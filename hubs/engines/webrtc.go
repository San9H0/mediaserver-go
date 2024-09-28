package engines

import (
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/thirdparty/h264"
	"mediaserver-go/utils/pointer"
)

func GetWebRTCCapabilities() map[pion.RTPCodecType][]pion.RTPCodecParameters {
	r := make(map[pion.RTPCodecType][]pion.RTPCodecParameters)
	r[pion.RTPCodecTypeAudio] = append(r[pion.RTPCodecTypeAudio], opusRTPCodecCapabilities())
	//r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], av1RTPCodecCapabilities())
	//r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], vp8RTPCodecCapabilities())
	r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], h264RTPCodecCapabilities()...)
	return r
}

func opusRTPCodecCapabilities() pion.RTPCodecParameters {
	return pion.RTPCodecParameters{
		RTPCodecCapability: pion.RTPCodecCapability{
			MimeType:     pion.MimeTypeOpus,
			ClockRate:    48000,
			Channels:     2,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 96,
	}
}

func h264RTPCodecCapabilities() []pion.RTPCodecParameters {
	return []pion.RTPCodecParameters{
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(1),
					ProfileLevelId:        "42001f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 111,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(0),
					ProfileLevelId:        "42001f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 112,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(1),
					ProfileLevelId:        "42e01f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 113,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(0),
					ProfileLevelId:        "42e01f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 114,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(1),
					ProfileLevelId:        "4d001f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 115,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:  pion.MimeTypeH264,
				ClockRate: 90000,
				Channels:  0,
				SDPFmtpLine: h264.RtpParameter{
					LevelAsymmetryAllowed: 1,
					PacketizationMode:     pointer.Uint8(0),
					ProfileLevelId:        "4d001f",
				}.String(),
				RTCPFeedback: nil,
			},
			PayloadType: 116,
		},
	}
}

func vp8RTPCodecCapabilities() pion.RTPCodecParameters {
	return pion.RTPCodecParameters{
		RTPCodecCapability: pion.RTPCodecCapability{
			MimeType:     pion.MimeTypeVP8,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 117,
	}
}

func av1RTPCodecCapabilities() pion.RTPCodecParameters {
	return pion.RTPCodecParameters{
		RTPCodecCapability: pion.RTPCodecCapability{
			MimeType:     pion.MimeTypeAV1,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "level-idx=5;profile=0;tier=0",
			RTCPFeedback: nil,
		},
		PayloadType: 118,
	}
}
