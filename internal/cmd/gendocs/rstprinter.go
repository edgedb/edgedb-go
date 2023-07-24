package main

import (
	"bytes"
	"fmt"
	"go/doc/comment"
	"strings"
)

// Based on go/doc/comment htmlPrinter

type rstPrinter struct {
	*comment.Printer
}

func printRST(p *comment.Printer, d *comment.Doc) []byte {
	hp := &rstPrinter{Printer: p}
	var out bytes.Buffer
	for _, x := range d.Content {
		hp.block(&out, x)
	}
	return out.Bytes()
}

func (p *rstPrinter) block(out *bytes.Buffer, x comment.Block) {
	switch x := x.(type) {
	default:
		fmt.Fprintf(out, "?%T", x)

	case *comment.Paragraph:
		p.text(out, x.Text)
		out.WriteString("\n\n")

	case *comment.Heading:
		out.WriteString("\n")
		var headerBytes bytes.Buffer
		p.text(&headerBytes, x.Text)
		out.Write(headerBytes.Bytes())
		out.WriteString("\n")
		out.WriteString(strings.Repeat("-", headerBytes.Len()))
		out.WriteString("\n\n")

	case *comment.Code:
		out.WriteString(".. code-block:: go\n\n    ")
		out.WriteString(strings.ReplaceAll(x.Text, "\n", "\n    "))
		out.WriteString("\n")

	}
}

func (p *rstPrinter) text(out *bytes.Buffer, x []comment.Text) {
	for _, t := range x {
		switch t := t.(type) {
		case comment.Plain:
			p.escape(out, string(t))
		case comment.Italic:
			out.WriteString("*")
			p.escape(out, string(t))
			out.WriteString("*")
		case *comment.Link:
			out.WriteString("`")
			if len(t.Text) == 1 {
				if s, ok := t.Text[0].(comment.Plain); ok &&
					string(s) == t.URL &&
					strings.HasPrefix(string(s), "https://www.edgedb.com/") {
					out.WriteString(string(s)[23:])
				} else {
					p.text(out, t.Text)
				}
			} else {
				p.text(out, t.Text)
			}
			out.WriteString(" <")
			p.escape(out, t.URL)
			out.WriteString(">`_")
		case *comment.DocLink:
			out.WriteString("`")
			if len(t.Text) == 1 {
				if s, ok := t.Text[0].(comment.Plain); ok &&
					strings.HasPrefix(string(s), "github.com/edgedb/edgedb-go") {
					urlParts := strings.Split(string(s), "/")
					out.WriteString(urlParts[len(urlParts)-1])
				} else {
					p.text(out, t.Text)
				}
			} else {
				p.text(out, t.Text)
			}
			out.WriteString(" <")
			p.escape(out, "https://pkg.go.dev/"+t.ImportPath)
			out.WriteString(">`_")
		}
	}
}

func (p *rstPrinter) escape(out *bytes.Buffer, s string) {
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "\\\\*", "\\*")
	out.WriteString(s)
}
