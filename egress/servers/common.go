package servers

import (
	"errors"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/types"
)

func filterMediaTypesInStream(stream *hubs.Stream, mediaTypes []types.MediaType) ([]*hubs.Track, error) {
	var sourceTracks []*hubs.Track
	tracksMap := stream.TracksMap()
	for _, requestMediaType := range mediaTypes {
		track, ok := tracksMap[requestMediaType]
		if !ok {
			continue
		}
		sourceTracks = append(sourceTracks, track)
	}

	if len(sourceTracks) == 0 {
		return nil, errors.New("no source tracks found")
	}

	return sourceTracks, nil
}
