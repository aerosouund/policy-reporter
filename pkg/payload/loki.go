package payload

type Value = []string

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values []Value           `json:"values"`
}
