// Copyright (c) 2018 Takayuki YATO (aka. "ZR")
//   GitHub:   https://github.com/zr-tex8r
//   Twitter:  @zr_tex8r
// Distributed under the MIT License.

package scpdf

import (
	"bytes"
	"fmt"
	"strings"
)

//-------------------------------------- the essential one

const eoBodyCode = // body=false
`0.5 0.72 m 0.64 0.72 0.76 0.65 0.76 0.55 c
0.76 0.51 0.72 0.47 0.67 0.44 c 0.79 0.41 0.84 0.32 0.84 0.25 c
0.84 0.13 0.75 0.08 0.68 0.08 c 0.32 0.08 l
0.25 0.08 0.16 0.13 0.16 0.25 c 0.16 0.32 0.21 0.41 0.33 0.44 c
0.28 0.47 0.24 0.51 0.24 0.55 c 0.24 0.65 0.36 0.72 0.5 0.72 c s`

var eoEyesCode = // eyes=true
eoCircle(0.40, 0.56, 0.02, 0.03, "f") + eoCircle(0.60, 0.56, 0.02, 0.03, "f")

const eoMouthCode = // mouth=true, mouthshape=smile
`0.40 0.48 m 0.45 0.45 0.55 0.45 0.60 0.48 c S`

const eoHatCode = // hat=true
`0.58 0.90 m 0.77 0.81 l 0.74 0.61 l 0.66 0.60 0.50 0.66 0.46 0.72 c
0.58 0.90 l b`

const eoArmsCode = // arms=true
`0.20 0.31 m 0.19 0.33 0.14 0.41 0.13 0.42 c
0.12 0.43 0.10 0.43 0.07 0.44 c 0.04 0.46 0.06 0.46 0.08 0.46 c
0.09 0.46 0.11 0.44 0.12 0.44 c 0.14 0.46 0.14 0.47 0.15 0.49 c
0.16 0.51 0.16 0.49 0.16 0.48 c 0.16 0.46 0.14 0.44 0.15 0.43 c
0.16 0.42 0.21 0.35 0.22 0.33 c 0.23 0.31 0.21 0.30 0.20 0.31 c b
0.80 0.31 m 0.81 0.33 0.86 0.41 0.87 0.42 c
0.88 0.43 0.90 0.43 0.93 0.44 c 0.96 0.46 0.94 0.46 0.92 0.46 c
0.91 0.46 0.89 0.44 0.88 0.44 c 0.86 0.46 0.86 0.47 0.85 0.49 c
0.84 0.51 0.84 0.49 0.84 0.48 c 0.84 0.46 0.86 0.44 0.85 0.43 c
0.84 0.42 0.79 0.35 0.78 0.33 c 0.77 0.31 0.79 0.30 0.80 0.31 c b`

const eoMufflerCode = // muffler=<color>
`0.27 0.48 m 0.42 0.38 0.58 0.38 0.73 0.48 c
0.75 0.46 0.76 0.44 0.77 0.41 c 0.77 0.39 0.75 0.37 0.73 0.36 c
0.74 0.33 0.74 0.31 0.76 0.26 c 0.75 0.25 0.72 0.24 0.66 0.23 c
0.66 0.27 0.65 0.30 0.63 0.34 c 0.42 0.30 0.32 0.35 0.24 0.41 c
0.25 0.45 0.26 0.47 0.27 0.48 c b`

var eoButtonsCode = // buttons=true
eoCircle(0.50, 0.16, 0.03, 0.03, "b") + eoCircle(0.50, 0.26, 0.03, 0.03, "b")

var eoSnowCode = // snow=true
eoCircle(0.07, 0.28, 0.04, 0.04, "s") + eoCircle(0.08, 0.68, 0.04, 0.04, "s") +
	eoCircle(0.13, 0.55, 0.04, 0.04, "s") + eoCircle(0.23, 0.76, 0.04, 0.04, "s") +
	eoCircle(0.42, 0.89, 0.04, 0.04, "s") + eoCircle(0.74, 0.89, 0.04, 0.04, "s") +
	eoCircle(0.88, 0.73, 0.04, 0.04, "s") + eoCircle(0.92, 0.53, 0.04, 0.04, "s") +
	eoCircle(0.94, 0.23, 0.04, 0.04, "s")

//--------------------------------------

func essentialCode(p page, scale float64) string {
	buf := new(bytes.Buffer)
	ci := makeColorInfo(p.muffler)
	fmt.Fprintln(buf, "0 G 0 g 1 j 1 J 0.01389 w")
	fmt.Fprintln(buf, eoBodyCode)
	fmt.Fprint(buf, eoEyesCode)
	fmt.Fprintln(buf, eoMouthCode)
	fmt.Fprintln(buf, eoHatCode)
	fmt.Fprintln(buf, eoArmsCode)
	fmt.Fprint(buf, eoButtonsCode)
	fmt.Fprint(buf, eoSnowCode)
	fmt.Fprintln(buf, ci.pdfCode(false), ci.pdfCode(true))
	fmt.Fprintln(buf, eoMufflerCode)
	return buf.String()
}

func eoCircle(cx, cy, rx, ry float64, op string) string {
	str := func(v float64) string {
		s := fmt.Sprintf("%.4f", v)
		return strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	const a = 0.55228475
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, str(cx+rx), str(cy), "m")
	fmt.Fprintln(buf, str(cx+rx), str(cy+a*ry), str(cx+a*rx), str(cy+ry), str(cx), str(cy+ry), "c")
	fmt.Fprintln(buf, str(cx-a*rx), str(cy+ry), str(cx-rx), str(cy+a*ry), str(cx-rx), str(cy), "c")
	fmt.Fprintln(buf, str(cx-rx), str(cy-a*ry), str(cx-a*rx), str(cy-ry), str(cx), str(cy-ry), "c")
	fmt.Fprintln(buf, str(cx+a*rx), str(cy-ry), str(cx+rx), str(cy-a*ry), str(cx+rx), str(cy), "c", op)
	return buf.String()
}
