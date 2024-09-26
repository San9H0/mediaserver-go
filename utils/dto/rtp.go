package dto

import "mediaserver-go/utils/types"

type IngressRTPRequest struct {
	Addr        string `json:"addr"`
	Port        int    `json:"port"`
	PayloadType uint8  `json:"payloadType"`
	MimeType    string `json:"mimeType"`
}

type IngressRTPResponse struct {
}

type EgressRTPRequest struct {
	Addr       string            `json:"addr"`
	Port       int               `json:"port"`
	MediaTypes []types.MediaType `json:"mediaTypes"`
}

type EgressRTPResponse struct {
	SDP string `json:"sdp"`
}
