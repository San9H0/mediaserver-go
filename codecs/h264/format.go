package h264

func RemoveAnnexB(data []byte) [][]byte {
	var result [][]byte
	startIndex := 0
	index := 0
	for index < len(data) {
		if data[index] != 0x00 {
			index++
			continue
		}
		if data[index+1] != 0x00 {
			index += 2
			continue
		}
		if data[index+2] == 0x01 {
			index += 3
			if startIndex != 0 {
				result = append(result, data[startIndex:index])
			}
			startIndex = index
			continue
		}
		if data[index+2] != 0x00 {
			index += 3
			continue
		}
		if data[index+3] == 0x01 {
			index += 4
			if startIndex != 0 {
				result = append(result, data[startIndex:index])
			}
			startIndex = index
			continue
		}
		index += 4
	}
	if startIndex != 0 {
		result = append(result, data[startIndex:])
	}
	return result
}
