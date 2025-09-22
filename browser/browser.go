package browser

const (
	// Element Properties
	CHECKED      = "checked"
	VALUE        = "value"
	TEXT_CONTENT = "textContent"

	// Style
	DISPLAY = "display"
)

type Update struct {
	Selector  string            `json:"selector"`
	Text      map[string]string `json:"text,omitempty"`
	Integer   map[string]int    `json:"integer,omitempty"`
	Boolean   map[string]bool   `json:"boolean,omitempty"`
	Attribute map[string]string `json:"attribute,omitempty"`
	Style     map[string]string `json:"style,omitempty"`
	Remove    *bool             `json:"remove,omitempty"`
}
