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

var Fonts = map[string]Font{
	"Balsamiq Sans": {
		name:          "Balsamiq Sans",
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Fredoka": {
		name:          "Fredoka",
		defaultWeight: "regular",
		weights: map[string]string{
			"300":     "300",
			"regular": "400",
			"500":     "500",
			"600":     "600",
			"bold":    "700",
		},
	},

	"Inter": {
		name:          "Inter",
		defaultWeight: "regular",
		weights: map[string]string{
			"100":     "100",
			"200":     "200",
			"300":     "300",
			"regular": "400",
			"500":     "500",
			"600":     "600",
			"bold":    "700",
			"800":     "800",
			"900":     "900",
		},
	},

	"Luckiest Guy": {
		name:          "Luckiest Guy",
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
		},
	},

	"Mada": {
		name:          "Mada",
		defaultWeight: "regular",
		weights: map[string]string{
			"200":     "200",
			"300":     "300",
			"regular": "400",
			"500":     "500",
			"600":     "600",
			"bold":    "700",
			"800":     "800",
			"900":     "900",
		},
	},

	"Nunito": {
		name:          "Nunito",
		defaultWeight: "regular",
		weights: map[string]string{
			"200":     "200",
			"300":     "300",
			"regular": "400",
			"600":     "600",
			"bold":    "700",
			"800":     "800",
			"900":     "900",
			"1000":    "1000",
		},
	},

	"Poppins": {
		name:          "Poppins",
		defaultWeight: "regular",
		weights: map[string]string{
			"100":     "100",
			"200":     "200",
			"300":     "300",
			"regular": "400",
			"500":     "500",
			"600":     "600",
			"bold":    "700",
			"800":     "800",
			"900":     "900",
		},
	},

	"Raleway": {
		name:          "Raleway",
		defaultWeight: "regular",
		weights: map[string]string{
			"100":     "100",
			"200":     "200",
			"300":     "300",
			"regular": "400",
			"500":     "500",
			"600":     "600",
			"bold":    "700",
			"800":     "800",
			"900":     "900",
		},
	},

	// web safe fonts
	"Arial": {
		name:          "Arial",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Verdana": {
		name:          "Verdana",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Tahoma": {
		name:          "Tahoma",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Trebuchet MS": {
		name:          "Trebuchet MS",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Times New Roman": {
		name:          "Times New Roman",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Georgia": {
		name:          "Georgia",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Garamond": {
		name:          "Garamond",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},

	"Courier New": {
		name:          "Courier New",
		websafe:       true,
		defaultWeight: "regular",
		weights: map[string]string{
			"regular": "400",
			"bold":    "700",
		},
	},
}
