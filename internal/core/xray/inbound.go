package xray

// InboundConfig represents an Xray inbound configuration.
type InboundConfig struct {
	Listen         string `json:"listen"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings,omitempty"`
	Tag            string `json:"tag"`
	Sniffing       string `json:"sniffing,omitempty"`
}

// Equals compares two InboundConfig instances for deep equality.
func (c *InboundConfig) Equals(other *InboundConfig) bool {
	if c.Listen != other.Listen {
		return false
	}
	if c.Port != other.Port {
		return false
	}
	if c.Protocol != other.Protocol {
		return false
	}
	if c.Settings != other.Settings {
		return false
	}
	if c.StreamSettings != other.StreamSettings {
		return false
	}
	if c.Tag != other.Tag {
		return false
	}
	if c.Sniffing != other.Sniffing {
		return false
	}
	return true
}

// OutboundConfig represents an Xray outbound configuration.
type OutboundConfig struct {
	Tag            string `json:"tag"`
	Protocol       string `json:"protocol"`
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings,omitempty"`
	SendThrough    string `json:"sendThrough,omitempty"`
	Mux            string `json:"mux,omitempty"`
	ProxySettings  string `json:"proxySettings,omitempty"`
}
