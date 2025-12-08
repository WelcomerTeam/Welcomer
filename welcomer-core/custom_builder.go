package welcomer

type CustomWelcomerImage struct {
	Fill   string                     `json:"fill"`
	Stroke *CustomWelcomerImageStroke `json:"stroke,omitempty"`

	Dimensions [2]int `json:"dimensions"`

	Layers []CustomWelcomerImageLayer `json:"layers"`
}

type CustomWelcomerImageLayerType int

const (
	CustomWelcomerImageLayerTypeText CustomWelcomerImageLayerType = iota
	CustomWelcomerImageLayerTypeImage
	CustomWelcomerImageLayerTypeShapeRectangle
	CustomWelcomerImageLayerTypeShapeCircle
)

type CustomWelcomerImageLayer struct {
	Type  CustomWelcomerImageLayerType `json:"type,omitempty"`
	Value string                       `json:"value"`

	Dimensions [2]int `json:"dimensions"`
	Position   [2]int `json:"position"`

	Rotation  int  `json:"rotation"`
	InvertedX bool `json:"inverted_x"`
	InvertedY bool `json:"inverted_y"`

	// BorderRadius will either be an integer or a percentage string.
	BorderRadius [4]string `json:"border_radius"`

	Fill   string                     `json:"fill,omitempty"`
	Stroke *CustomWelcomerImageStroke `json:"stroke,omitempty"`

	Typography *CustomWelcomerImageLayerTypography `json:"typography,omitempty"`
}

type CustomWelcomerImageStroke struct {
	Color string `json:"color"`
	Width int    `json:"width"`
}

type HorizontalAlignment string

const (
	HorizontalAlignmentLeft   HorizontalAlignment = "left"
	HorizontalAlignmentCenter HorizontalAlignment = "center"
	HorizontalAlignmentRight  HorizontalAlignment = "right"
)

type VerticalAlignment string

const (
	VerticalAlignmentTop    VerticalAlignment = "top"
	VerticalAlignmentCenter VerticalAlignment = "center"
	VerticalAlignmentBottom VerticalAlignment = "bottom"
)

type CustomWelcomerImageLayerTypography struct {
	FontFamily          string              `json:"font_family"`
	FontWeight          string              `json:"font_weight"`
	FontSize            int                 `json:"font_size"`
	LineHeight          float64             `json:"line_height"`
	LetterSpacing       float64             `json:"letter_spacing"`
	HorizontalAlignment HorizontalAlignment `json:"horizontal_alignment"`
	VerticalAlignment   VerticalAlignment   `json:"vertical_alignment"`
}
