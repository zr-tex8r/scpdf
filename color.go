// Copyright (c) 2018 Takayuki YATO (aka. "ZR")
//   GitHub:   https://github.com/zr-tex8r
//   Twitter:  @zr_tex8r
// Distributed under the MIT License.

package scpdf

import (
	"image/color"
	"strings"
)

type colorInfo struct {
	op     string
	params []float64
}

func makeColorInfo(c color.Color) (ci *colorInfo) {
	switch c := c.(type) {
	case color.Gray:
		return &colorInfo{"g", makeParams(0xFF, uint(c.Y))}
	case color.Gray16:
		return &colorInfo{"g", makeParams(0xFFFF, uint(c.Y))}
	case color.CMYK:
		return &colorInfo{"k", makeParams(0xFF, uint(c.C), uint(c.M), uint(c.Y), uint(c.K))}
	default:
		r, g, b, _ := c.RGBA()
		return &colorInfo{"rg", makeParams(0xFFFF, uint(r), uint(g), uint(b))}
	}
}

func makeParams(mv uint, v ...uint) []float64 {
	t := make([]float64, len(v))
	for i := 0; i < len(v); i++ {
		t[i] = float64(v[i]) / float64(mv)
	}
	return t
}

func (ci *colorInfo) pdfCode(fill bool) string {
	buf := make([]byte, 0, 32)
	for _, p := range ci.params {
		s := realStr(p)
		buf = append(append(buf, s...), ' ')
	}
	if fill {
		buf = append(buf, ci.op...)
	} else {
		buf = append(buf, strings.ToUpper(ci.op)...)
	}
	return string(buf)
}
