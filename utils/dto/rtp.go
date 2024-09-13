package dto

import "mediaserver-go/utils/types"

type IngressRTPRequest struct {
	Token       string
	Addr        string `json:"addr"`
	Port        int    `json:"port"`
	PayloadType uint8  `json:"payloadType"`
}

type IngressRTPResponse struct {
}

type EgressRTPRequest struct {
	Addr       string            `json:"addr"`
	Port       int               `json:"port"`
	MediaTypes []types.MediaType `json:"mediaTypes"`
}

type EgressRTPResponse struct {
	PayloadType uint8  `json:"payloadType"`
	SDP         string `json:"sdp"`
}
