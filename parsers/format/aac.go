package format

import (
	"errors"
	"fmt"
)

var (
	errInvalidData = errors.New("invalid data")
)

var samplingRates = []int{
	96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350,
}

// AAC 프로파일 인덱스 테이블 (AAC-LC 등 다양한 프로파일)
var profiles = []string{
	"AAC Main", "AAC LC (Low Complexity)", "AAC SSR (Scalable Sample Rate)", "AAC LTP (Long Term Prediction)",
}

type AACConfig struct {
	Profile      string
	SamplingRate int
	Channel      int
}

func (a *AACConfig) ParseAACAudioSpecificConfig(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("short data. %w", errInvalidData)
	}

	// 첫 번째 바이트에서 상위 5비트를 사용하여 AAC 프로파일 추출
	profileIndex := int(data[0]&0xF8) >> 3 // 11111000
	if profileIndex >= len(profiles) {
		return fmt.Errorf("unknown AAC profile index: %d, %w", profileIndex, errInvalidData)
	}
	profile := profiles[profileIndex]

	// 첫 번째 바이트 하위 3비트와 두 번째 바이트 상위 1비트를 사용하여 샘플링 레이트 인덱스 추출
	samplingIndex := int(((data[0] & 0x07) << 1) | ((data[1] & 0x80) >> 7)) // 00000111 | 10000000
	if samplingIndex >= len(samplingRates) {
		return fmt.Errorf("unknown sampling rate index: %d", samplingIndex)
	}
	samplingRate := samplingRates[samplingIndex]

	// 두 번째 바이트에서 상위 3비트를 사용하여 채널 구성 추출
	channelConfig := (data[1] & 0x78) >> 3 // 01111000

	a.Profile = profile
	a.SamplingRate = samplingRate
	a.Channel = int(channelConfig)
	return nil
}
