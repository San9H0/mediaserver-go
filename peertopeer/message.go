package peertopeer

type Header struct {
	Type string `json:"type"`
}

type SDPMessage struct {
	Header
	SDP string `json:"sdp"`
}
