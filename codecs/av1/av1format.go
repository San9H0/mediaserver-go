package av1

import (
	"fmt"
	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
)

type OBUType int

const (
	OBUSequenceHeader       OBUType = 1
	OBUTemporalDelimiter    OBUType = 2
	OBUFrameHeader          OBUType = 3
	OBUTileGroup            OBUType = 4
	OBUMetadata             OBUType = 5
	OBUFrame                OBUType = 6
	OBURedundantFrameHeader OBUType = 7
	OBUTileList             OBUType = 8
	OBuPadding              OBUType = 15
)

var LevelIdxMap = map[uint8]string{
	0:  "20", // 2.0
	1:  "21", // 2.1
	2:  "22", // 2.2
	3:  "23", // 2.3
	4:  "30", // 3.0
	5:  "31", // 3.1
	6:  "32", // 3.2
	7:  "33", // 3.3
	8:  "40", // 4.0
	9:  "41", // 4.1
	10: "42", // 4.2
	11: "43", // 4.3
	12: "50", // 5.0
	13: "51", // 5.1
	14: "52", // 5.2
	15: "53", // 5.3
	16: "60", // 6.0
	17: "61", // 6.1
	18: "62", // 6.2
	19: "63", // 6.3
	20: "70", // 7.0
	21: "71", // 7.1
	22: "72", // 7.2
	23: "73", // 7.3
}

/*
+-+-+-+-+-+-+-+-+
|f|t y p e|f|s|-|
+-+-+-+-+-+-+-+-+
*/
// 5.3.2. OBU header syntax
func ParseOBUHeaderSyntax(data []byte) error {
	forbiddenBit := int((data[0] & 0x80) >> 7)
	if forbiddenBit != 0 {
		return fmt.Errorf("forbidden bit is set")
	}

	obuType := OBUType(data[0]&0x78) >> 3
	obuExtensionFlag := int((data[0] & 0x04) >> 2)
	obuHasSizeField := int((data[0] & 0x02) >> 1)
	_ = obuHasSizeField

	if obuExtensionFlag == 1 {
		// TODO: extension flag is not supported yet
	}

	if obuType == OBUSequenceHeader {
		//var header av1.SequenceHeader
		//if err := header.Unmarshal(data); err != nil {
		//	log.Logger.Error("failed to unmarshal sequence header", zap.Error(err))
		//}
		//fmt.Println("[TESTDEBUG] width:", header.Width(), "height:", header.Height())
	}

	return nil
}

// AV1 MIME 표현 생성 함수
func createAV1MIME(seqHeader av1.SequenceHeader) string {
	// 프로파일 결정 (0: Main, 1: High, 2: Professional)
	var profile string
	switch seqHeader.SeqProfile {
	case 0:
		profile = "0" // Main profile
	case 1:
		profile = "1" // High profile
	case 2:
		profile = "2" // Professional profile
	default:
		profile = "unknown"
	}

	// 첫 번째 레벨과 티어를 가져옴 (복수의 operating points가 있을 경우)
	level := seqHeader.SeqLevelIdx[0]
	tier := seqHeader.SeqTier[0]

	// 티어 결정 (Main: M, High: H)
	var tierStr string
	if tier {
		tierStr = "H"
	} else {
		tierStr = "M"
	}

	bitDepth := seqHeader.ColorConfig.BitDepth

	// MIME 문자열 생성
	mime := fmt.Sprintf("av01.%s.%02d%s.%02d", profile, level, tierStr, bitDepth)
	return mime
}

func ParseExtraData(extradata []byte) (seqHeaderData []byte) {
	return extradata[4:]
	//fmt.Println("[TESTDEBUG] extradata:", len(extradata))
	////marker := extradata[0] & 0x80
	////version := extradata[0] & 0x7f
	////profile := extradata[1] & 0xe0 >> 5
	////seqLevelIdx := extradata[1] & 0x1f
	//_ = extradata[2]
	//_ = extradata[3] // padding
	//var seqHeader av1.SequenceHeader
	//var oh av1.OBUHeader
	//buf := extradata[4:]
	//if err := oh.Unmarshal(buf); err != nil {
	//	fmt.Println("OBUHeader err:", err)
	//	return nil, nil
	//}
	//
	//offset := sn
	//if oh.HasSize {
	//	s, sn, err := LEB128Unmarshal(extradata[5:])
	//	if err != nil {
	//		fmt.Println("LEB128Unmarshal err:", err)
	//		return nil, nil
	//	}
	//	fmt.Println("s:", s, ", sn:", sn)
	//}
	//
	//fmt.Println("[TESTDEBUG] exp header extradata:", len(extradata[4:]))
	//if err := seqHeader.Unmarshal(extradata[4:]); err != nil {
	//	fmt.Println("err:", err)
	//}
	//fmt.Printf("[TESTDEBUG] seqHeader:%+v\n", seqHeader)

	//data := make([]byte, 0, len(c.seqData)+4)
	//version := 1
	//// marker and version
	//data = append(data, byte(0x80|(version&0x7F)))
	//// seq profile and level
	//data = append(data, (c.header.SeqProfile&0x07)<<5|c.header.SeqLevelIdx[0]&0x1F)
	//
	//ext := byte(0)
	//tier := utils.GetIntFromBool(c.header.SeqTier[0])
	//depth0 := utils.GetIntFromBool(c.header.ColorConfig.BitDepth > 8)
	//depth1 := utils.GetIntFromBool(c.header.ColorConfig.BitDepth == 12)
	//monochrome := utils.GetIntFromBool(c.header.ColorConfig.MonoChrome)
	//subSampleX := utils.GetIntFromBool(c.header.ColorConfig.SubsamplingX)
	//subSampleY := utils.GetIntFromBool(c.header.ColorConfig.SubsamplingY)
	//samplePosition := c.header.ColorConfig.ChromaSamplePosition
	//ext = byte(tier&0x01) << 7
	//ext = byte(depth0&0x01) << 6
	//ext = byte(depth1&0x01) << 5
	//ext = byte(monochrome&0x01) << 4
	//ext = byte(subSampleX&0x01) << 3
	//ext = byte(subSampleY&0x01) << 2
	//ext = byte(samplePosition & 0x03)
	//data = append(data, ext)
	//data = append(data, 0x00) // padding
	//data = append(data, c.seqData...)
	//return data, nil

	//return nil, nil
}
