package vt100

import (
	"fmt"
	"html"
	"strings"
)

// Color represents RGB color values
type Color struct {
	R, G, B uint8
}

func (c Color) Clone() Color {
	return Color{R: c.R, G: c.G, B: c.B}
}

func (c Color) Equals(t Color) bool {
	return c.R == t.R && c.G == t.G && c.B == t.B
}

func (c Color) IsDefault() bool {
	return ColorDefault.Equals(c)
}

// Contrast 返回兩個顏色之間的對比度比率
// 基於 WCAG 2.0 標準的相對亮度計算公式
func (c Color) Contrast(other Color) float64 {
	l1 := c.Luminance()
	l2 := other.Luminance()
	// 確保 l1 >= l2
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// Luminance 返回顏色的相對亮度
// 基於 WCAG 2.0 標準的公式
// https://www.w3.org/TR/WCAG20/#relativeluminancedef
func (c Color) Luminance() float64 {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255

	// 應用 gamma 校正
	if r <= 0.03928 {
		r = r / 12.92
	} else {
		r = ((r + 0.055) / 1.055) * 2.4
	}
	if g <= 0.03928 {
		g = g / 12.92
	} else {
		g = ((g + 0.055) / 1.055) * 2.4
	}
	if b <= 0.03928 {
		b = b / 12.92
	} else {
		b = ((b + 0.055) / 1.055) * 2.4
	}

	// 計算相對亮度
	return 0.2126*r + 0.7152*g + 0.0722*b
}

var (
	ColorDefault = Color{0, 0, 0}
	ColorBlack   = Color{0, 0, 0}
	ColorRed     = Color{205, 0, 0}
	ColorGreen   = Color{0, 205, 0}
	ColorYellow  = Color{205, 205, 0}
	ColorBlue    = Color{0, 0, 238}
	ColorMagenta = Color{205, 0, 205}
	ColorCyan    = Color{0, 205, 205}
	ColorWhite   = Color{229, 229, 229}
)

// Style represents text styling attributes (SGR - Select Graphic Rendition)
type Style struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Inverse       bool
	Blink         bool
	Dim           bool
	Hidden        bool
	Foreground    Color
	Background    Color
}

func (s *Style) Clone() *Style {
	if s == nil {
		return &Style{
			Foreground: ColorWhite,
			Background: ColorDefault,
		}
	}

	return &Style{
		Bold:          s.Bold,
		Italic:        s.Italic,
		Underline:     s.Underline,
		Strikethrough: s.Strikethrough,
		Inverse:       s.Inverse,
		Blink:         s.Blink,
		Dim:           s.Dim,
		Hidden:        s.Hidden,
		Foreground:    s.Foreground.Clone(),
		Background:    s.Background.Clone(),
	}
}

func (s Style) Equals(t Style) bool {
	if s.Foreground.Equals(t.Foreground) &&
		s.Background.Equals(t.Background) &&
		s.Bold != t.Bold &&
		s.Italic != t.Italic &&
		s.Underline != t.Underline &&
		s.Strikethrough != t.Strikethrough &&
		s.Inverse != t.Inverse &&
		s.Blink != t.Blink &&
		s.Dim != t.Dim &&
		s.Hidden != t.Hidden {
		return false
	}
	return true
}

func (c Cell) HtmlChar() string {
	return html.EscapeString(string(c.Char))
}

func (s Style) HtmlStyle() string {
	styles := []string{"white-space: pre"}
	fg, bg := s.Foreground, s.Background
	if s.Inverse {
		fg, bg = bg, fg
	}

	// Add colors
	if !fg.IsDefault() {
		styles = append(styles, fmt.Sprintf("color: rgb(%d, %d, %d)", fg.R, fg.G, fg.B))
	}
	if !bg.IsDefault() {
		styles = append(styles, fmt.Sprintf("background-color: rgb(%d, %d, %d)", bg.R, bg.G, bg.B))
	}

	// Add text attributes
	if s.Bold {
		styles = append(styles, "font-weight: bold")
	}
	if s.Dim {
		styles = append(styles, "opacity: 0.7")
	}
	if s.Italic {
		styles = append(styles, "font-style: italic")
	}
	if s.Underline {
		styles = append(styles, "text-decoration: underline")
	}
	if s.Strikethrough {
		styles = append(styles, "text-decoration: line-through")
	}
	if s.Hidden {
		styles = append(styles, "visibility: hidden")
	}
	return strings.Join(styles, "; ")
}

// Contrast 返回前景色與背景色之間的對比度比率
// 對比度比率範圍為 1-21，越大表示對比度越高
// 根據 WCAG 2.0 標準，文字與背景的對比度應至少為 4.5:1（一般文字）或 3:1（大文字）
func (s Style) Contrast() float64 {
	return s.Foreground.Contrast(s.Background)
}

// ContrastWith 返回當前 Style 與另一個 Style 的前景色之間的對比度比率
func (s Style) ContrastWith(other Style) float64 {
	return s.Foreground.Contrast(other.Foreground)
}

func (s Style) Set(params []int) Style {
	i := 0
	for i < len(params) {
		p := params[i]
		i++
		switch p {
		case 0: // Reset
			s = DefaultStyle()
		case 1: // Bold
			s.Bold = true
		case 2: // Dim
			s.Dim = true
		case 3: // Italic
			s.Italic = true
		case 4: // Underline
			s.Underline = true
		case 5, 6: // Blink
			s.Blink = true
		case 7: // Inverse
			s.Inverse = true
		case 8: // Hidden
			s.Hidden = true
		case 9: // Strikethrough
			s.Strikethrough = true
		case 22: // Normal intensity
			s.Bold = false
			s.Dim = false
		case 23: // Not italic
			s.Italic = false
		case 24: // Not underline
			s.Underline = false
		case 25: // Not blink
			s.Blink = false
		case 27: // Not inverse
			s.Inverse = false
		case 28: // Not hidden
			s.Hidden = false
		case 29: // Not strikethrough
			s.Strikethrough = false
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground standard colors
			s.Foreground = standardColor(p - 30)
		case 38: // Extended foreground color
			if i < len(params) {
				switch params[i] {
				case 5:
					i++
					if i < len(params) {
						s.Foreground = color256(params[i])
					}
				case 2:
					i++
					if i+2 < len(params) {
						s.Foreground = Color{uint8(params[i]), uint8(params[i+1]), uint8(params[i+2])}
						i += 3
					}
				}
			}
		case 39: // Default foreground
			s.Foreground = ColorDefault
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background standard colors
			s.Background = standardColor(p - 40)
		case 48: // Extended background color
			if i < len(params) {
				switch params[i] {
				case 5:
					i++
					if i < len(params) {
						s.Background = color256(params[i])
					}
				case 2:
					i++
					if i+2 < len(params) {
						s.Background = Color{uint8(params[i]), uint8(params[i+1]), uint8(params[i+2])}
						i += 3
					}
				}
			}
		case 49: // Default background
			s.Background = ColorDefault
		case 90, 91, 92, 93, 94, 95, 96, 97: // Foreground bright colors
			s.Foreground = brightColor(p - 90)
		case 100, 101, 102, 103, 104, 105, 106, 107: // Background bright colors
			s.Background = brightColor(p - 100)

		}
	}
	return s
}

// DefaultStyle returns a style with default attributes
func DefaultStyle() Style {
	return Style{
		Foreground: ColorWhite,
		Background: ColorDefault,
	}
}

// standardColor returns the standard ANSI color
func standardColor(n int) Color {
	colors := []Color{
		{0, 0, 0},       // Black
		{205, 0, 0},     // Red
		{0, 205, 0},     // Green
		{205, 205, 0},   // Yellow
		{0, 0, 238},     // Blue
		{205, 0, 205},   // Magenta
		{0, 205, 205},   // Cyan
		{229, 229, 229}, // White
	}
	if n >= 0 && n < 8 {
		return colors[n]
	}
	return ColorDefault
}

// brightColor returns the bright ANSI color
func brightColor(n int) Color {
	colors := []Color{
		{127, 127, 127}, // Bright Black (Gray)
		{255, 0, 0},     // Bright Red
		{0, 255, 0},     // Bright Green
		{255, 255, 0},   // Bright Yellow
		{92, 92, 255},   // Bright Blue
		{255, 0, 255},   // Bright Magenta
		{0, 255, 255},   // Bright Cyan
		{255, 255, 255}, // Bright White
	}
	if n >= 0 && n < 8 {
		return colors[n]
	}
	return ColorDefault
}

// color256 returns a color from the 256-color palette
func color256(n int) Color {
	if n < 0 || n > 255 {
		return ColorDefault
	}
	if n < 16 {
		return standardColor(n)
	}
	if n < 232 {
		n -= 16
		r := (n / 36) * 51
		g := ((n / 6) % 6) * 51
		b := (n % 6) * 51
		return Color{uint8(r), uint8(g), uint8(b)}
	}
	v := ((n - 232) * 10) + 8
	return Color{uint8(v), uint8(v), uint8(v)}
}

// Cell represents a single cell in the terminal screen
type Cell struct {
	Char  rune
	Style Style
}

func newScreen(cols, rows int) [][]Cell {
	cells := make([][]Cell, rows)
	for i := range cells {
		cells[i] = newLine(cols)
	}
	return cells
}

func newLine(cols int) []Cell {
	line := make([]Cell, cols)
	for i := range line {
		line[i] = newCell()
	}
	return line
}

func newCell() Cell {
	return Cell{Char: ' ', Style: DefaultStyle()}
}

// DefaultCell returns a cell with default attributes
func DefaultCell() Cell {
	return Cell{
		Char:  ' ',
		Style: DefaultStyle(),
	}
}
