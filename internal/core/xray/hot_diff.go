package xray

import "strings"

// ConfigDiff holds the differences between two config snapshots.
type ConfigDiff struct {
	InboundsToAdd    []InboundConfig
	InboundsToRemove []InboundConfig
	InboundsToUpdate []InboundConfig
	OutboundsChanged bool
	RoutingChanged   bool
	DNChanged        bool
}

// DiffConfig compares old and new configs and returns the changes.
// This is used to decide whether a full restart is needed or hot-reload via gRPC will suffice.
func DiffConfig(old, new *XrayConfig) *ConfigDiff {
	diff := &ConfigDiff{}

	// Build maps for quick lookup
	oldInbounds := make(map[string]*InboundConfig)
	for i := range old.InboundConfigs {
		oldInbounds[old.InboundConfigs[i].Tag] = &old.InboundConfigs[i]
	}

	newInbounds := make(map[string]*InboundConfig)
	for i := range new.InboundConfigs {
		newInbounds[new.InboundConfigs[i].Tag] = &new.InboundConfigs[i]
	}

	// Find removed inbounds
	for tag, oldIb := range oldInbounds {
		if newIb, ok := newInbounds[tag]; !ok {
			diff.InboundsToRemove = append(diff.InboundsToRemove, *oldIb)
		} else if !oldIb.Equals(newIb) {
			diff.InboundsToRemove = append(diff.InboundsToRemove, *oldIb)
			diff.InboundsToAdd = append(diff.InboundsToAdd, *newIb)
		}
	}

	// Find added inbounds
	for tag, newIb := range newInbounds {
		if _, ok := oldInbounds[tag]; !ok {
			diff.InboundsToAdd = append(diff.InboundsToAdd, *newIb)
		}
	}

	// Check outbounds
	diff.OutboundsChanged = old.OutboundConfigs != new.OutboundConfigs

	// Check routing
	diff.RoutingChanged = old.RouterConfig != new.RouterConfig

	return diff
}

// NeedsFullRestart returns true if the config changes require a full process restart
// rather than hot-reload via gRPC.
func (d *ConfigDiff) NeedsFullRestart() bool {
	// Full restart needed if outbounds or transport changed
	return d.OutboundsChanged || len(d.InboundsToRemove) > 0
}

// HasChanges returns true if there are any changes to apply.
func (d *ConfigDiff) HasChanges() bool {
	return len(d.InboundsToAdd) > 0 ||
		len(d.InboundsToRemove) > 0 ||
		d.OutboundsChanged ||
		d.RoutingChanged
}

// Summary returns a human-readable summary of the diff.
func (d *ConfigDiff) Summary() string {
	var parts []string
	if len(d.InboundsToAdd) > 0 {
		parts = append(parts, formatCount(len(d.InboundsToAdd), "inbound(s) to add"))
	}
	if len(d.InboundsToRemove) > 0 {
		parts = append(parts, formatCount(len(d.InboundsToRemove), "inbound(s) to remove"))
	}
	if d.OutboundsChanged {
		parts = append(parts, "outbounds changed")
	}
	if d.RoutingChanged {
		parts = append(parts, "routing changed")
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, ", ")
}

func formatCount(n int, noun string) string {
	return strings.Replace(formatInt(n)+" "+noun, "1 ", "1 ", 1)
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []rune(itoa(n))
	out := make([]rune, 0, len(digits)+len(digits)/3)
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, d)
	}
	return string(out)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
