package servers

import (
	"errors"
	"mediaserver-go/hubs"
	"mediaserver-go/utils/types"
)

func filterMediaTypesInStream(stream *hubs.Stream, mediaTypes []types.MediaType) ([]*hubs.HubSource, error) {
	var hubSources []*hubs.HubSource
	sourcesMap := stream.SourcesMap()
	for _, requestMediaType := range mediaTypes {
		source, ok := sourcesMap[requestMediaType]
		if !ok {
			continue
		}
		hubSources = append(hubSources, source)
	}

	if len(hubSources) == 0 {
		return nil, errors.New("no source tracks found")
	}

	return hubSources, nil
}
