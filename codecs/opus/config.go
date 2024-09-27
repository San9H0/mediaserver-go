package opus

type Parameters struct {
	Channels     int
	SampleRate   int
	SampleFormat int
}

type Config struct {
	parameters Parameters
}

func NewConfig(parameters Parameters) *Config {
	return &Config{
		parameters: parameters,
	}
}

func (c *Config) Channels() int {
	return c.parameters.Channels
}

func (c *Config) SampleRate() int {
	return c.parameters.SampleRate
}

func (c *Config) SampleFormat() int {
	return c.parameters.SampleFormat
}
