package servers

import (
	"context"
	"fmt"
	"github.com/pion/rtp"
	"mediaserver-go/parser/codecparser"
	"net"
)

type RTPServer struct {
	conn       *net.UDPConn
	h264parser codecparser.H264
}

func NewRTPServer(ip string, port int) (RTPServer, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return RTPServer{}, err
	}

	return RTPServer{
		conn: conn,
	}, nil
}

func (r *RTPServer) Run(ctx context.Context) error {
	buffer := make([]byte, 1500)
	for {
		n, src, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("error reading from udp:", err)
			return err
		}
		_ = src

		var pkt rtp.Packet
		if err := pkt.Unmarshal(buffer[:n]); err != nil {
			fmt.Println("error unmarshalling rtp packet:", err)
			continue
		}

		au := r.h264parser.GetAU(pkt.Payload)
		if len(au) == 0 {
			continue
		}

		for _, accessUnit := range au {
			nalUnit := accessUnit[0] & 0x1F

			//fmt.Println("[TESDTEBUG] rtp recv ts:", pkt.Timestamp, ", sn:", pkt.SequenceNumber, ", pt:", pkt.PayloadType)
			fmt.Printf("[TESDTEBUG] nalUnit:%d, payload:%x\n", nalUnit, accessUnit[1:20])
		}
	}
	return nil
}
