package dto

type WebRTCRequest struct {
	Token string
	Offer string
}

type WebRTCResponse struct {
	Answer string
}
