package core

// Category groups modules in the settings navigation.
type Category string

const (
	CategorySystem    Category = "system"
	CategoryWindowing Category = "windowing"
	CategoryGaming    Category = "gaming"
	CategoryInput     Category = "input"
	CategoryFiles     Category = "files"
	CategoryAdvanced  Category = "advanced"
)

// CategoryOrder is the display order in the navigation pane.
var CategoryOrder = []Category{
	CategorySystem,
	CategoryWindowing,
	CategoryGaming,
	CategoryInput,
	CategoryFiles,
	CategoryAdvanced,
}

var CategoryLabels = map[Category]string{
	CategorySystem:    "System Tools",
	CategoryWindowing: "Windowing & Layouts",
	CategoryGaming:    "Gaming",
	CategoryInput:     "Input / Output",
	CategoryFiles:     "File Management",
	CategoryAdvanced:  "Advanced",
}

// SettingType selects which control the settings UI renders for a field.
type SettingType string

const (
	SettingToggle   SettingType = "toggle"
	SettingShortcut SettingType = "shortcut"
	SettingSlider   SettingType = "slider"
	SettingSelect   SettingType = "select"
	SettingText     SettingType = "text"
)

type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// SettingField is a declarative description of one module setting.
type SettingField struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Description string         `json:"description,omitempty"`
	Type        SettingType    `json:"type"`
	Value       any            `json:"value"`
	Options     []SelectOption `json:"options,omitempty"`
	Min         *float64       `json:"min,omitempty"`
	Max         *float64       `json:"max,omitempty"`
	Step        *float64       `json:"step,omitempty"`
}

// Info is the module metadata consumed by the settings UI.
type Info struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	Category    Category       `json:"category"`
	Enabled     bool           `json:"enabled"`
	Available   bool           `json:"available"` // false renders a "coming soon" badge
	Settings    []SettingField `json:"settings,omitempty"`
}

// Module is the contract every PolyTools utility implements.
type Module interface {
	Info() Info
	SetEnabled(enabled bool) error
	UpdateSetting(key string, value any) error
}
