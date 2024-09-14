package dto

import "mediaserver-go/utils/types"

type RTMPRequest struct {
	Addr        string          `json:"addr"`
	Port        int             `json:"port"`
	PayloadType uint8           `json:"payloadType"`
	CodecType   types.CodecType `json:"codecType"`
}

type RTMPResponse struct {
}
