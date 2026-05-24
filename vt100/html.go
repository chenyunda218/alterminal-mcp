package vt100

import (
	"fmt"
	"strings"
)

// ToHTML converts the terminal screen to an HTML representation
func (vt *VT100) ToHTML() string {
	var sb strings.Builder
	sb.WriteString("<div class=\"terminal\" style=\"background: #000000\">")
	for i, row := range vt.Screen {
		var rowStringBuilder rowStringBuilder
		sb.WriteString("<div>")
		for j, cell := range row {
			if i == vt.Cursor.Y && j == vt.Cursor.X {
				sb.WriteString("<span sylte=\"cursor\"></span>")
			}
			rowStringBuilder.Put(cell)
		}
		rowStringBuilder.EndRow()
		sb.WriteString(rowStringBuilder.String())
		sb.WriteString("</div>")
	}
	sb.WriteString("</div>")
	return sb.String()
}

type rowStringBuilder struct {
	sb           strings.Builder
	currentStyle *Style
}

func (r *rowStringBuilder) Put(cell Cell) {
	if r.currentStyle == nil {
		r.currentStyle = cell.Style.Clone()
		r.sb.WriteString(fmt.Sprintf("<span style=\"%s\">", cell.Style.HtmlStyle()))
		r.sb.WriteString(cell.HtmlChar())
		return
	}
	if !r.currentStyle.Equals(cell.Style) {
		r.sb.WriteString("</span>")
		r.currentStyle = cell.Style.Clone()
		r.sb.WriteString(fmt.Sprintf("<span style=\"%s\">", cell.Style.HtmlStyle()))
		r.sb.WriteString(cell.HtmlChar())
		return
	}
	r.sb.WriteString(cell.HtmlChar())
}

func (r *rowStringBuilder) String() string {
	return r.sb.String()
}

func (r *rowStringBuilder) EndRow() {
	r.sb.WriteString("</span>")
}
