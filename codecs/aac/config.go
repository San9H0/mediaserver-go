package aac

type Parameters struct {
	SampleRate   int
	Channels     int
	SampleFormat int
}

type Config struct {
	SampleRate   int
	Channels     int
	SampleFormat int
}

func NewConfig(parameters Parameters) *Config {
	return &Config{
		SampleRate:   parameters.SampleRate,
		Channels:     parameters.Channels,
		SampleFormat: parameters.SampleFormat,
	}
}
