scpdf
======

Go: An ultra-simple PDF generation library for Snowman Comedians

```go
doc := scpdf.Doc{}
file, err := os.Create("essential.pdf")
if err != nil { /* blah */ }
n, err := doc.WriteTo(file)
if err != nil { /* blah */ }
file.Close()
```

API
---

See [GoDoc](https://godoc.org/github.com/zr-tex8r/scpdf).

LICENSE
-------

MIT License

CHANGELOG
---------

  * Version 0.18.0 <2018/11/30>
      - Added `(*Doc).Bytes()`.
      - Added `Version()`.
  * Version 0.8.2  <2018/11/02>
  * Version 0.8.1  <2018/07/15>
  * Version 0.8.0  <2018/02/27>
      - Beta release.

--------------------
Takayuki YATO (aka. "ZR")  
https://github.com/zr-tex8r
