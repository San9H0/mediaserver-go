package dto

import "mediaserver-go/utils/types"

type IngressFileRequest struct {
	Path       string            `json:"path"`
	MediaTypes []types.MediaType `json:"mediaTypes"`
	Live       bool              `json:"live"`
}

type IngressFileResponse struct {
}

type EgressFileRequest struct {
	Path       string            `json:"path"`
	MediaTypes []types.MediaType `json:"mediaTypes"`
	Interval   int               `json:"interval"`
}

type EgressFileResponse struct {
}
