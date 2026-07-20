package specwizardmgr

import (
	"encoding/json"
	"strconv"
)

// Option is a single selectable item returned by the backend catalogs.
type Option struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Recommendation is a generated architecture-health insight card.
type Recommendation struct {
	Level       string `json:"level"` // "success" | "warning" | "critical"
	Title       string `json:"title"`
	Description string `json:"description"`
}

// optionsFrom extracts a slice of Option from a plugin /options response,
// trying each candidate key in order and tolerating missing keys. Each entry
// may be a map with id/name/description or a bare string.
func optionsFrom(resp map[string]any, keys ...string) []Option {
	for _, key := range keys {
		raw, ok := resp[key]
		if !ok || raw == nil {
			continue
		}
		items, ok := raw.([]interface{})
		if !ok {
			continue
		}
		out := make([]Option, 0, len(items))
		for _, it := range items {
			switch v := it.(type) {
			case map[string]interface{}:
				out = append(out, optionFromMap(v))
			case string:
				out = append(out, Option{ID: v, Name: v})
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return []Option{}
}

func optionFromMap(m map[string]interface{}) Option {
	opt := Option{}
	if v, ok := m["id"]; ok {
		opt.ID = toString(v)
	}
	if v, ok := m["name"]; ok {
		opt.Name = toString(v)
	} else if opt.ID != "" {
		opt.Name = opt.ID
	}
	if v, ok := m["description"]; ok {
		opt.Description = toString(v)
	}
	return opt
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
