package sessions

import (
	"fmt"
	"github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"
	"io"
)

type RTMPSession struct {
	rtmp.DefaultHandler
}

func (h *RTMPSession) OnServe(conn *rtmp.Conn) {
	fmt.Println("OnServe")
}

func (h *RTMPSession) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	fmt.Println("OnConnect")
	return nil
}

func (h *RTMPSession) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	fmt.Println("OnCreateStream")
	return nil
}

func (h *RTMPSession) OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error {
	fmt.Println("OnCreateStream")
	return nil
}

func (h *RTMPSession) OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error {
	fmt.Println("OnDeleteStream")
	return nil
}

func (h *RTMPSession) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	fmt.Println("OnPublish")
	return nil
}

func (h *RTMPSession) OnPlay(_ *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPlay) error {
	fmt.Println("OnPlay")
	return nil
}

func (h *RTMPSession) OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error {
	fmt.Println("OnFCPublish")
	return nil
}

func (h *RTMPSession) OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error {
	fmt.Println("OnFCUnpublish")
	return nil
}

func (h *RTMPSession) OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error {
	fmt.Println("OnSetDataFrame")
	return nil
}

func (h *RTMPSession) OnAudio(timestamp uint32, payload io.Reader) error {
	fmt.Println("OnAudio")
	return nil
}

func (h *RTMPSession) OnVideo(timestamp uint32, payload io.Reader) error {
	fmt.Println("OnVideo")
	return nil
}

func (h *RTMPSession) OnUnknownMessage(timestamp uint32, msg message.Message) error {
	fmt.Println("OnUnknownMessage")
	return nil
}

func (h *RTMPSession) OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error {
	fmt.Println("OnUnknownCommandMessage")
	return nil
}

func (h *RTMPSession) OnUnknownDataMessage(timestamp uint32, data *message.DataMessage) error {
	fmt.Println("OnUnknownDataMessage")
	return nil
}

func (h *RTMPSession) OnClose() {
	fmt.Println("OnClose")
}
