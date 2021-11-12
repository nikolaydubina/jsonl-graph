package color

import (
	"encoding/json"
	"image/color"
	"strconv"
)

// ColorScale is sequence of colors and numerical anchors between them
type ColorScale struct {
	Colors []color.RGBA
	Points []float64
}

// ColorMapping of manual match of string value to color
type ColorMapping map[string]color.RGBA

// ColorConfigVal is configuration for single key on how to color its value
type ColorConfigVal struct {
	ColorMapping *ColorMapping `json:"ColorMapping"`
	ColorScale   *ColorScale   `json:"ColorScale"`
}

// Color for a given value
func (cv ColorConfigVal) Color(v interface{}) color.Color {
	// convert value to string
	var key string
	if vs, ok := v.(string); ok {
		key = vs
	} else {
		vs, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		key = string(vs)
	}

	// first check manual values
	if cv.ColorMapping != nil {
		if c, ok := (*cv.ColorMapping)[key]; ok {
			return c
		}
	}

	// then check scale if present
	if cv.ColorScale != nil && len(cv.ColorScale.Points) > 0 && len(cv.ColorScale.Points) == (len(cv.ColorScale.Colors)-1) {
		if vs, err := strconv.ParseFloat(key, 64); err == nil {
			idx := 0
			for idx < len(cv.ColorScale.Points) && cv.ColorScale.Points[idx] <= vs {
				idx++
			}
			if idx >= len(cv.ColorScale.Colors) {
				idx = len(cv.ColorScale.Colors) - 1
			}
			return cv.ColorScale.Colors[idx]
		}

	}

	return color.White
}

// ColorConfig defines how to translate arbitrary values to some color
type ColorConfig map[string]ColorConfigVal

// Color checks if config is found for that key, and computes color based on config
func (c ColorConfig) Color(k string, v interface{}) color.Color {
	vc, ok := c[k]
	if !ok {
		return color.White
	}
	return vc.Color(v)
}
