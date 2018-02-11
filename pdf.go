// Copyright (c) 2018 Takayuki YATO (aka. "ZR")
//   GitHub:   https://github.com/zr-tex8r
//   Twitter:  @zr_tex8r
// Distributed under the MIT License.

package scpdf

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"fmt"
	"io"
	"time"
	"unicode/utf16"
)

const pdfUseDeflate = true

var pdfNix, pdfNL, pdfSP = []byte{}, []byte{'\n'}, []byte{' '}

func pdfChunk(format string, a ...interface{}) []byte {
	return []byte(fmt.Sprintf(format, a...))
}

func pdfNewLine(prev []byte) []byte {
	if len(prev) > 0 && prev[len(prev)-1] == '\n' {
		return pdfNix
	}
	return pdfNL
}

func pdfFormatDate(t time.Time) string {
	b := []byte(t.Format("::20060102150405Z07:00"))
	b[0] = 'D'
	if len(b) > 17 {
		b[19] = '\''
		b = append(b, '\'')
	}
	return string(b)
}

func pdfParseDate(s string) (t time.Time, err error) {
	errFormat := func() (time.Time, error) {
		return t, fmt.Errorf("invalid PDF date format: %q", s)
	}
	b, l := []byte(s), len(s)
	if string(b[0:2]) == "D:" {
		b, l = b[2:], l-2
	}
	switch l {
	case 15:
		break
	case 4, 6, 8, 10, 12, 14, 18, 21:
		d := "00000101000000+00'00'"
		b = append(b, d[l:]...)
		if b[17] != '\'' || b[20] != '\'' {
			return errFormat()
		}
		b = b[:20]
		b[17] = ':'
	default:
		return errFormat()
	}
	if t, err = time.Parse("20060102150405Z07:00", string(b)); err != nil {
		return errFormat()
	}
	return
}

func pdfSure(ok bool) {
	if !ok {
		panic(fmt.Errorf("%s: INTERNAL ERROR", packageName))
	}
}

type pdfId int

func (id pdfId) String() string {
	pdfSure(0 < id && id < 65536)
	return fmt.Sprint(int(id), " 0 R")
}

type pdfDoc struct {
	writer  *bufio.Writer
	pos     int
	xref    []int
	pages   []pdfId
	pagesId pdfId
	version string

	title, author, subject          string
	producer, creator, creationDate string
}

const pdfUseNow = "<now?\b>"

func newPdfDoc(wr io.Writer, info map[string]string) (d *pdfDoc, err error) {
	infoGet := func(key, dflt string) string {
		if value, ok := info[key]; ok {
			return value
		}
		return dflt
	}
	d = &pdfDoc{
		writer:       bufio.NewWriter(wr),
		xref:         []int{0},
		version:      infoGet("version", dfltPdfVersion),
		title:        infoGet("title", ""),
		author:       infoGet("author", ""),
		subject:      infoGet("subject", ""),
		producer:     infoGet("producer", packageName+"-"+version),
		creator:      infoGet("creator", packageName),
		creationDate: infoGet("creationDate", pdfUseNow),
	}
	d.pagesId = d.newId()
	err = d.write(pdfChunk("%%PDF-%s\n", d.version), []byte("%\xc5\xdd\xc4\xb6\n"))
	return
}

func (d *pdfDoc) write(chunk ...[]byte) (err error) {
	var n int
	for _, c := range chunk {
		n, err = d.writer.Write(c)
		d.pos += n
		if err != nil {
			return
		}
	}
	return
}

func (d *pdfDoc) newId() (id pdfId) {
	id = pdfId(len(d.xref))
	d.xref = append(d.xref, 0)
	return
}

func (d *pdfDoc) startObject(id pdfId) (err error) {
	pdfSure(id > 0 && d.xref[id] == 0)
	d.xref[id] = d.pos
	err = d.write(pdfChunk("%v 0 obj\n", int(id)))
	return
}

func (d *pdfDoc) addObject(id pdfId, value []byte) (err error) {
	if err = d.startObject(id); err != nil {
		return
	}
	err = d.write(value, pdfNewLine(value), []byte("endobj\n"))
	return
}

func (d *pdfDoc) addStream(id pdfId, data []byte) (err error) {
	if err = d.startObject(id); err != nil {
		return
	}
	filter := ""
	if pdfUseDeflate {
		defl := new(bytes.Buffer)
		wr := zlib.NewWriter(defl)
		wr.Write(data)
		wr.Close()
		if defl.Len() <= len(data)-32 {
			data, filter = defl.Bytes(), "/Filter/FlateDecode"
		}
	}
	err = d.write(pdfChunk("<</Length %v%s>>\nstream\n", len(data), filter),
		data, pdfNewLine(data), []byte("endstream\nendobj\n"))
	return
}

func (d *pdfDoc) addPage(id, contents, resources pdfId, chunk []byte) (err error) {
	if err = d.startObject(id); err != nil {
		return
	}
	d.pages = append(d.pages, id)
	err = d.write(pdfChunk("<</Type/Page/Contents %v/Resources %v/Parent %v\n",
		contents, resources, d.pagesId),
		chunk, []byte(">>\nendobj\n"))
	return
}

func (d *pdfDoc) finish() (err error) {
	infoChunk := func(key, value string) []byte {
		if value == "" {
			return pdfNix
		}
		return pdfChunk("/%s%s\n", key, pdfStr(value))
	}

	kids, plt := new(bytes.Buffer), 0
	kids.WriteString("[")
	for _, id := range d.pages {
		fmt.Fprint(kids, id)
		if kids.Len() > plt+70 {
			kids.Write(pdfNL)
			plt = kids.Len()
		} else {
			kids.Write(pdfSP)
		}
	}
	kids.Truncate(kids.Len() - 1)
	kids.WriteString("]")

	// page tree
	if err = d.startObject(d.pagesId); err != nil {
		return
	}
	err = d.write(pdfChunk("<</Type/Pages/Count %d/Kids\n", len(d.pages)),
		kids.Bytes(), []byte(">>\nendobj\n"))
	if err != nil {
		return
	}

	// catalog object
	catalogId := d.newId()
	err = d.addObject(catalogId, pdfChunk("<</Type/Catalog/Pages %v>>", d.pagesId))
	if err != nil {
		return
	}

	// document information object
	if d.creationDate == pdfUseNow {
		d.creationDate = pdfFormatDate(time.Now())
	}
	docInfoId := d.newId()
	if err = d.startObject(docInfoId); err != nil {
		return
	}
	err = d.write([]byte("<<"),
		infoChunk("Title", d.title),
		infoChunk("Author", d.author),
		infoChunk("Subject", d.subject),
		infoChunk("Producer", d.producer),
		infoChunk("Creator", d.creator),
		infoChunk("CreationDate", d.creationDate),
		infoChunk("ModDate", d.creationDate),
		[]byte("/Trapped/False>>\nendobj\n"))
	if err != nil {
		return
	}

	// xref table
	xrefPos, lmt := d.pos, len(d.xref)
	err = d.write(pdfChunk("xref\n0 %d\n", lmt),
		pdfChunk("%010d %05d f \n", 0, 65535))
	if err != nil {
		return
	}
	for i := 1; i < lmt; i++ {
		pdfSure(d.xref[i] > 0)
		err = d.write(pdfChunk("%010d %05d n \n", d.xref[i], 0))
	}

	// trailer
	ss := fmt.Sprintf("%s/%s/%v", d.creationDate, d.title, xrefPos)
	hash := md5.Sum([]byte(ss))
	err = d.write([]byte("trailer\n"),
		pdfChunk("<</Size %v/Root %v/Info %v\n", lmt, catalogId, docInfoId),
		pdfChunk("/ID[<%X><%X>]>>\n", hash, hash),
		[]byte("startxref\n"),
		pdfChunk("%v\n", xrefPos),
		[]byte("%%EOF\n"))
	if err != nil {
		return
	}

	// done
	err = d.writer.Flush()
	return
}

func pdfStr(str string) string {
	buf, ok := new(bytes.Buffer), true
	buf.WriteString("(")
LOOP:
	for _, u := range str {
		switch u {
		case '\n', '\r', '\t', '\b', '\f':
			t := fmt.Sprintf("%q", u)
			buf.WriteString(t[1:3])
		case '(', ')':
			fmt.Fprintf(buf, "\\%c", u)
		default:
			if u < 32 || u == 127 {
				fmt.Fprintf(buf, "\\%03o", u)
			} else if u > 127 {
				ok = false
				break LOOP
			} else {
				buf.Write([]byte{byte(u)})
			}
		}
	}
	if ok {
		buf.WriteString(")")
		return buf.String()
	}
	// use hexstring for non-ASCII strings
	buf.Truncate(0)
	buf.WriteString("<FEFF")
	for _, u := range utf16.Encode([]rune(str)) {
		fmt.Fprintf(buf, "%04X", u)
	}
	buf.WriteString(">")
	return buf.String()
}
