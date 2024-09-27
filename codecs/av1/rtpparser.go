package av1

import (
	"github.com/pion/rtp"
	"go.uber.org/zap"
	"mediaserver-go/codecs"
	"mediaserver-go/utils/log"
)

type RTTPParser struct {
	fragments [][]byte

	cb func(codec codecs.Codec)
}

func NewRTPParser(cb func(codec codecs.Codec)) *RTTPParser {
	return &RTTPParser{
		cb: cb,
	}
}

/*
OBU 는 AV1에서 가장 작은 단위의 유닛. RTP Payload 에 가장 앞은, Aggregation Header (1Byte) 가 붙고 다음 길이가 옵셔널하게 오고, OBU 가 온다.

 0 1 2 3 4 5 6 7
+-+-+-+-+-+-+-+-+
|Z|Y| W |N|-|-|-|
+-+-+-+-+-+-+-+-+

Z (비트 0) OBU가 여러개의 RTP패킷에 분할될 때 사용된다. 1이라면 이전 RTP 패킷에서 분할된 연속된 OBU 이다. 0이라면 새로운 OBU 라는것을 의미.
Y (비트 1) 1 로 설정되면, 마지막 OBU 요소가 현재 RTP 패킷에서 끝나지않고 다음 패킷으로 이어짐. 0 이면 이 패킷에서 OBU가 완전히 끝남.
W (비트 2, 3) 해당 RTP 패킷에 OBU 요소가 몇개 포함되어있는지 나타냄. 0이면 각 OBU 는 길이필드로 표시함. 1,2,3 이면 RTP 페이로드에 OBU 길이 필드는 없고, 1,2,3개의 OBU 가 포함되어있음.
N (비트 4) 1이면 비디오 시퀀스의 첫번째 패킷임을 나타냄 0이면 새로운 비디오 시퀀스의 시작이 아님 아님. 새로운 비디오 시퀀스의 시작을 나타내는 신호.

RTP Payload는 집합(Aggregation) 또는 분할(fragmentation) 이 될 수있으며,
첫번째 또는 마지막 OBU 요소가 OBU의 분할된 조각일 수 있다.
*/
// Parse RTP Packet.
func (a *RTTPParser) Parse(rtpPacket *rtp.Packet) [][]byte {
	rtpPayload := rtpPacket.Payload

	Z := int(rtpPayload[0] & 0x80 >> 7)
	Y := int(rtpPayload[0] & 0x40 >> 6)
	W := int(rtpPayload[0] & 0x30 >> 4)
	N := int(rtpPayload[0] & 0x08 >> 3)

	if Z == 0 {
		a.fragments = nil
	}

	index := 0
	offset := 1

	for {
		if offset >= len(rtpPayload) {
			break
		}
		length := uint(0)
		read := 0
		var err error
		if W == 0 || index < W-1 {
			length, read, err = LEB128Unmarshal(rtpPayload[offset:])
			if err != nil {
				log.Logger.Error("av1 unmarshal err", zap.Error(err))
				return nil
			}
			offset += read
		} else {
			length = uint(len(rtpPayload[offset:]))
		}

		data := rtpPayload[offset : offset+int(length)]
		if index == 0 && Z == 1 { // 이전요소가 이어짐.
			a.fragments[len(a.fragments)-1] = append(a.fragments[len(a.fragments)-1], data...)
		} else {
			if N == 1 && index == 0 {
				config := &Config{}
				if err := config.UnmarshalSequenceHeader(data); err == nil {
					av1Codec := NewAV1(config)
					a.cb(av1Codec)
				} else {
					log.Logger.Error("av1 unmarshal err", zap.Error(err))
				}
			}
			a.fragments = append(a.fragments, data)
		}

		offset += int(length)
		index += 1
	}

	if Y == 1 {
		return nil
	}
	return a.fragments
}

/*
OBU (Open Bitstream Unit)
OBU 는 AV1 비트스트림의 최소 단위이다. 각 OBU 데이터는 헤더와 페이로드로 구성된다.
OBU 헤더는 OBU의 종류와 길이를 나타내며, OBU 페이로드는 실제 비디오 데이터를 포함한다.

av1의 참조프레임
최대 8개의 참조프레임을 관리할 수 있으며, 최대 7개 새로운 프레임을 참조할 수 있다.

1 바이트가 헤더임 (av1 aggregation header).
+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
| AV1 aggr hdr  |                                               |
+-+-+-+-+-+-+-+-+                                               |
|                                                               |
|                   Bytes 2..N of AV1 payload                   |
|                                                               |
|                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                               :    OPTIONAL RTP padding       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+



*/
