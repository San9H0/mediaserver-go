package dto

type IngressRTPRequest struct {
	Token       string
	Addr        string `json:"addr"`
	Port        int    `json:"port"`
	PayloadType uint8  `json:"payloadType"`
	SSRC        uint32 `json:"ssrc"`
}

type IngressRTPResponse struct {
}

type EgressRTPRequest struct {
	Token string
	Addr  string `json:"addr"`
	Port  int    `json:"port"`
}

type EgressRTPResponse struct {
	PayloadType uint8  `json:"payloadType"`
	SDP         string `json:"sdp"`
}
