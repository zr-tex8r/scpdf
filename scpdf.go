// Copyright (c) 2018 Takayuki YATO (aka. "ZR")
//   GitHub:   https://github.com/zr-tex8r
//   Twitter:  @zr_tex8r
// Distributed under the MIT License.

// Package scpdf is an ultra-simple PDF generation library for all
// Snowman Comedians.
//
// To achieve the simplicity, the content of the PDF document is limited
// to essential stuffs, that is, snowman pictures. You can choose the
// color of the mufflers that the snowmen wear.
//
//   doc := scpdf.Doc{}
//   file, err := os.Create("essential.pdf")
//   if err != nil { /* blah */ }
//   n, err := doc.WriteTo(file)
//   if err != nil { /* blah */ }
//   file.Close()
package scpdf

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"math"
	"strings"
	"time"
)

const (
	packageName = "scpdf"
	version     = "0.18.0"
)

const dfltPdfVersion = "1.4"
const dfltWidth = 210 * 72 / 25.4
const dfltHeight = 294 * 72 / 25.4

var dfltMuffler = &color.NRGBA{255, 0, 0, 255}

const stdScale = 0.6

// Doc represents an SC-oriented PDF document.
type Doc struct {
	width, height float64
	pages         []page
	frozen        bool
	info          map[string]string
}

type page struct {
	muffler color.Color
	scale   float64
}

// Version returns the version of this package.
func Version() string {
	return version
}

// NewWithSize makes a new document with the given width and height
// (measured in PDF points).
func NewWithSize(width, height float64) (*Doc, error) {
	d := &Doc{}
	err := d.SetPageSize(width, height)
	return d, err
}

// SetPageSize sets the width and the height (measured in PDF points) of
// this document. The length values must be positive.
func (d *Doc) SetPageSize(width, height float64) error {
	if d.frozen {
		return errFrozen()
	} else if width <= 0 || height <= 0 {
		return fmt.Errorf("illegal page size (%.3gx%.3g)", width, height)
	}
	d.width, d.height = width, height
	return nil
}

// PageSize returns the width and height of this document.
func (d *Doc) PageSize() (float64, float64) {
	d.autoPageSize()
	return d.width, d.height
}

func (d *Doc) autoPageSize() {
	if d.width == 0 && d.height == 0 {
		d.width, d.height = dfltWidth, dfltHeight
	}
}

// AddPageScaled adds to this document a new page with the given
// muffler color and the scale value to the standard snowman size.
func (d *Doc) AddPageScaled(muffler color.Color, scale float64) error {
	if d.frozen {
		return errFrozen()
	} else if scale <= 0 {
		return fmt.Errorf("illegal scale value (%.3g)", scale)
	} else if muffler == nil {
		return fmt.Errorf("illegal muffler value (nil)")
	}
	page := page{muffler: muffler, scale: scale * stdScale}
	d.pages = append(d.pages, page)
	return nil
}

// AddPageScaled adds to this document a new page with the given
// muffler color (without scale).
func (d *Doc) AddPage(muffler color.Color) error {
	return d.AddPageScaled(muffler, 1)
}

// SetDocInfo specifies several kinds of information of this document.
// The input is given as a map of strings.
//
// The following keys are effective:
//
//   version: PDF version, such as "1.5" (default: "1.4")
//   title / author / subject: with obvious meanings (default: empty)
//   creator: name of the PDF generation software (default: "scpdf")
//   creationDate: in the form "D:20180808120000+09'00'" (default: now)
func (d *Doc) SetDocInfo(info map[string]string) (err error) {
	if d.frozen {
		return errFrozen()
	}
	t := make(map[string]string, len(info))
	for k, v := range info { // copy
		t[k] = v
	}
	d.info = t
	return
}

// WriteTo generates the PDF file content of this document and writes
// the resulted bytes to an io.Writer. Since a PDF file with no pages is
// disallowed, the "default page" (a snowman with red muffler) will be
// added if the document has no pages yet.
//
// After generation, the document is frozen; no further modification
// to the document is allowed.
func (d *Doc) WriteTo(wr io.Writer) (n int64, err error) {
	d.autoPageSize()
	if len(d.pages) == 0 {
		if err = d.AddPage(dfltMuffler); err != nil {
			return
		}
	}

	d.frozen = true
	buf := new(bytes.Buffer)

	// header
	pd, err := newPdfDoc(wr, d.info)
	if err != nil {
		return int64(pd.pos), err
	}
	resources := pd.newId()
	err = pd.addObject(resources, []byte("<</ProcSet[/PDF]>>\n"))

	// pages
	for _, p := range d.pages {
		buf.Reset()
		cod, len := transformCode(d.width, d.height, p.scale)
		fmt.Fprintln(buf, "q", cod)
		fmt.Fprintf(buf, essentialCode(p, len))
		fmt.Fprintln(buf, "Q")
		contents := pd.newId()
		err = pd.addStream(contents, buf.Bytes())
		if err != nil {
			return int64(pd.pos), err
		}
		err = pd.addPage(pd.newId(), contents, resources,
			pdfChunk("/MediaBox[0 0 %s %s]", realStr(d.width), realStr(d.height)))
		if err != nil {
			return int64(pd.pos), err
		}
	}

	// done
	err = pd.finish()
	return int64(pd.pos), err
}

func transformCode(width, height, scale float64) (string, float64) {
	len := math.Max(width, height) * scale
	ox, oy := (width-len)/2, (height-len)/2
	s := fmt.Sprintf("%s 0 0 %s %s %s cm", realStr(len), realStr(len), realStr(ox), realStr(oy))
	return s, len
}

// PdfBytes generates the PDF file content and returns the resulted bytes.
// The same notice on WriteTo also applies.
func (d *Doc) PdfBytes() (bs []byte, err error) {
	var buf bytes.Buffer
	if _, err = d.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}

// Bytes is an alias to PdfBytes, provided to satisfy scdoc.Doc.
func (d *Doc) Bytes() (bs []byte, err error) {
	return d.PdfBytes()
}

// String dumps the internal information of the Doc value.
func (d *Doc) String() string {
	buf := make([]byte, 0, 32)
	buf = append(buf, "Doc"...)
	if d.frozen {
		buf = append(buf, '*')
	}
	s := fmt.Sprintf("(%.3gx%.3g)", d.width, d.height)
	buf = append(buf, s...)
	buf = append(buf, '[')
	for _, p := range d.pages {
		s := fmt.Sprintf("%+v*%.3g;", p.muffler, p.scale)
		buf = append(buf, s...)
	}
	buf[len(buf)-1] = ']'
	return string(buf)
}

func realStr(v float64) string {
	s := fmt.Sprintf("%.3f", v)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func errFrozen() error {
	return fmt.Errorf("document is frozen")
}

// FormatDate stringifies a time.Time value in the form suitable for
// PDF creation date information.
func FormatDate(t time.Time) string {
	return pdfFormatDate(t)
}
