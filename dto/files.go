package dto

import "mediaserver-go/utils/types"

type IngressFileRequest struct {
	Token      string
	Path       string            `json:"path"`
	MediaTypes []types.MediaType `json:"mediaTypes"`
	Live       bool              `json:"live"`
}

type IngressFileResponse struct {
}

type EgressFileRequest struct {
	Token string
	Path  string `json:"path"`
}

type EgressFileResponse struct {
}
