package engines

import (
	pion "github.com/pion/webrtc/v3"
	"mediaserver-go/thirdparty/h264"
	"mediaserver-go/utils/pointer"
	"slices"
)

func GetWebRTCCapabilities(useRTX bool) map[pion.RTPCodecType][]pion.RTPCodecParameters {
	r := make(map[pion.RTPCodecType][]pion.RTPCodecParameters)
	r[pion.RTPCodecTypeAudio] = append(r[pion.RTPCodecTypeAudio], opusRTPCodecCapabilities())
	r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], h264RTPCodecCapabilities(useRTX)...)
	r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], vp8RTPCodecCapabilities(useRTX)...)
	r[pion.RTPCodecTypeVideo] = append(r[pion.RTPCodecTypeVideo], av1RTPCodecCapabilities(useRTX)...)
	return r
}

func GetWHIPRTPHeaderExtensionCapabilities() map[pion.RTPCodecType][]pion.RTPHeaderExtensionCapability {
	result := make(map[pion.RTPCodecType][]pion.RTPHeaderExtensionCapability)
	result[pion.RTPCodecTypeVideo] = getRTPHeaderExtensionCapabilitiesVideo()
	result[pion.RTPCodecTypeAudio] = getRTPHeaderExtensionCapabilitiesAudio()
	return result
}

func GetWHEPRTPHeaderExtensionCapabilities() map[pion.RTPCodecType][]pion.RTPHeaderExtensionCapability {
	result := make(map[pion.RTPCodecType][]pion.RTPHeaderExtensionCapability)
	result[pion.RTPCodecTypeVideo] = []pion.RTPHeaderExtensionCapability{
		{URI: "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01"},
		{URI: "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time"},
		//{URI: "http://www.webrtc.org/experiments/rtp-hdrext/playout-delay"},
	}
	result[pion.RTPCodecTypeAudio] = nil
	return result
}

func getRTPHeaderExtensionCapabilitiesVideo() []pion.RTPHeaderExtensionCapability {
	return []pion.RTPHeaderExtensionCapability{
		//{URI: "urn:ietf:params:rtp-hdrext:sdes:mid"},
		//{URI: "urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id"},
		//{URI: "urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id"},
	}
}

func getRTPHeaderExtensionCapabilitiesAudio() []pion.RTPHeaderExtensionCapability {
	return nil
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

func h264RTPCodecCapabilities(useRTX bool) []pion.RTPCodecParameters {
	params := []pion.RTPCodecParameters{
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
				MimeType:    "video/rtx",
				ClockRate:   90000,
				SDPFmtpLine: "apt=111",
			},
			PayloadType: 112,
		},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:  pion.MimeTypeH264,
		//		ClockRate: 90000,
		//		Channels:  0,
		//		SDPFmtpLine: h264.RtpParameter{
		//			LevelAsymmetryAllowed: 1,
		//			PacketizationMode:     pointer.Uint8(0),
		//			ProfileLevelId:        "42001f",
		//		}.String(),
		//		RTCPFeedback: nil,
		//	},
		//	PayloadType: 113,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:    "video/rtx",
		//		ClockRate:   90000,
		//		SDPFmtpLine: "apt=113",
		//	},
		//	PayloadType: 114,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:  pion.MimeTypeH264,
		//		ClockRate: 90000,
		//		Channels:  0,
		//		SDPFmtpLine: h264.RtpParameter{
		//			LevelAsymmetryAllowed: 1,
		//			PacketizationMode:     pointer.Uint8(1),
		//			ProfileLevelId:        "42e01f",
		//		}.String(),
		//		RTCPFeedback: nil,
		//	},
		//	PayloadType: 115,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:    "video/rtx",
		//		ClockRate:   90000,
		//		SDPFmtpLine: "apt=115",
		//	},
		//	PayloadType: 116,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:  pion.MimeTypeH264,
		//		ClockRate: 90000,
		//		Channels:  0,
		//		SDPFmtpLine: h264.RtpParameter{
		//			LevelAsymmetryAllowed: 1,
		//			PacketizationMode:     pointer.Uint8(0),
		//			ProfileLevelId:        "42e01f",
		//		}.String(),
		//		RTCPFeedback: nil,
		//	},
		//	PayloadType: 117,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:    "video/rtx",
		//		ClockRate:   90000,
		//		SDPFmtpLine: "apt=117",
		//	},
		//	PayloadType: 118,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:  pion.MimeTypeH264,
		//		ClockRate: 90000,
		//		Channels:  0,
		//		SDPFmtpLine: h264.RtpParameter{
		//			LevelAsymmetryAllowed: 1,
		//			PacketizationMode:     pointer.Uint8(1),
		//			ProfileLevelId:        "4d001f",
		//		}.String(),
		//		RTCPFeedback: nil,
		//	},
		//	PayloadType: 119,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:    "video/rtx",
		//		ClockRate:   90000,
		//		SDPFmtpLine: "apt=119",
		//	},
		//	PayloadType: 120,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:  pion.MimeTypeH264,
		//		ClockRate: 90000,
		//		Channels:  0,
		//		SDPFmtpLine: h264.RtpParameter{
		//			LevelAsymmetryAllowed: 1,
		//			PacketizationMode:     pointer.Uint8(0),
		//			ProfileLevelId:        "4d001f",
		//		}.String(),
		//		RTCPFeedback: nil,
		//	},
		//	PayloadType: 121,
		//},
		//{
		//	RTPCodecCapability: pion.RTPCodecCapability{
		//		MimeType:    "video/rtx",
		//		ClockRate:   90000,
		//		SDPFmtpLine: "apt=121",
		//	},
		//	PayloadType: 122,
		//},
	}
	return slices.DeleteFunc(params, func(param pion.RTPCodecParameters) bool {
		return !useRTX && param.MimeType == "video/rtx"
	})
}

func vp8RTPCodecCapabilities(useRTX bool) []pion.RTPCodecParameters {
	params := []pion.RTPCodecParameters{
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:     pion.MimeTypeVP8,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: nil,
			},
			PayloadType: 123,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:    "video/rtx",
				ClockRate:   90000,
				SDPFmtpLine: "apt=123",
			},
			PayloadType: 124,
		},
	}
	return slices.DeleteFunc(params, func(param pion.RTPCodecParameters) bool {
		return useRTX && param.MimeType == "video/rtx"
	})
}

func av1RTPCodecCapabilities(useRTX bool) []pion.RTPCodecParameters {
	params := []pion.RTPCodecParameters{
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:     pion.MimeTypeAV1,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-idx=5;profile=0;tier=0",
				RTCPFeedback: nil,
			},
			PayloadType: 125,
		},
		{
			RTPCodecCapability: pion.RTPCodecCapability{
				MimeType:    "video/rtx",
				ClockRate:   90000,
				SDPFmtpLine: "apt=125",
			},
			PayloadType: 124,
		},
	}
	return slices.DeleteFunc(params, func(param pion.RTPCodecParameters) bool {
		return useRTX && param.MimeType == "video/rtx"
	})
}
