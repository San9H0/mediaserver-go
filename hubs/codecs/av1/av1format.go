package av1

import (
	"fmt"
)

type OBUType int

const (
	OBUSequenceHeader       OBUType = 1
	OBUTemporalDelimiter    OBUType = 2
	OBUFrameHeader          OBUType = 3
	OBUTileGroup            OBUType = 4
	OBUFrame                OBUType = 6
	OBURedundantFrameHeader OBUType = 7
	OBUTileList             OBUType = 8
	OBuPadding              OBUType = 15
)

type SequenceHeader struct {
	FrameWidth  int
	FrameHeight int
}

// open_bitstream_unit
func ParseOpenBitStreamUnit(data []byte) {
	ParseOBUHeaderSyntax(data)
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
