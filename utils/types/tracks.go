package types

type Track struct {
	mediaType MediaType
	codecType CodecType
}

func NewTrack(mediaType MediaType, codecType CodecType) Track {
	return Track{
		mediaType: mediaType,
		codecType: codecType,
	}
}
