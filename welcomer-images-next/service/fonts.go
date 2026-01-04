package service

type Font struct {
	name          string
	defaultWeight string
	weights       map[string]string
	websafe       bool
}

func (f Font) GetWeight(weight string) string {
	if weight == "" {
		weight = f.defaultWeight
	}

	if val, ok := f.weights[weight]; ok {
		return val
	}

	return "normal"
}

// Fonts map is defined in fonts_generated.go
// To regenerate with latest Google Fonts, run: python3 generate_fonts.py
