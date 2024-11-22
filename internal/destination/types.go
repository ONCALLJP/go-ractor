package destination

type Destination struct {
	Type    string      `yaml:"type"` // slack, lineworks, custom
	Token   TokenConfig `yaml:"token,omitempty"`
	Channel string      `yaml:"channel,omitempty"`
	URL     string      `yaml:"url,omitempty"`
}

type TokenConfig struct {
	Type  string `yaml:"type,omitempty"` // bearer, basic, api_key
	Value string `yaml:"value,omitempty"`
}
