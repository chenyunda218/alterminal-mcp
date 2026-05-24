package vt100

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/creack/pty"
)

type Status int

const (
	StatusNormal Status = iota
	StatusEscape
	StatusCSI
	StatusOSC
	StatusEscapeIntermediate
	StatusCSIEntry
)

type Cursor struct {
	X int
	Y int
}

type VT100 struct {
	Status       Status
	Cols         int
	Rows         int
	Cursor       Cursor
	Screen       [][]Cell
	buf          []byte
	savedCursor  Cursor
	currentStyle Style
	ptmx         *os.File
	cmd          *exec.Cmd
	SignalRecord *BoundedList[byte]
	Log          *BoundedList[string]
}

func (vt *VT100) saveCursor() {
	vt.savedCursor = Cursor{X: vt.Cursor.X, Y: vt.Cursor.Y}
}

func (vt *VT100) restoreCursor() {
	vt.Cursor.X = vt.savedCursor.X
	vt.Cursor.Y = vt.savedCursor.Y
}

func (vt *VT100) Reset() {
	vt.Status = StatusNormal
	vt.Cursor = Cursor{X: 0, Y: 0}
	vt.clearBuf()
	vt.savedCursor = Cursor{X: 0, Y: 0}
	vt.Screen = newScreen(vt.Cols, vt.Rows)
	vt.currentStyle = DefaultStyle()
	// 重啟 ptmx
	if vt.cmd != nil && vt.cmd.Path != "" {
		if vt.ptmx != nil {
			vt.ptmx.Close()
		}
		// 終止舊的 cmd 進程
		if vt.cmd != nil && vt.cmd.Process != nil {
			vt.cmd.Process.Kill()
		}
		// 創建新的 cmd
		newCmd := exec.Command(vt.cmd.Path, vt.cmd.Args[1:]...)
		ptmx, err := pty.StartWithSize(newCmd, &pty.Winsize{Rows: uint16(vt.Rows), Cols: uint16(vt.Cols)})
		if err != nil {
			log.Printf("Failed to restart ptmx: %v", err)
			return
		}
		vt.ptmx = ptmx
		vt.cmd = newCmd
		buf := make([]byte, 4096)
		go func() {
			for {
				n, err := ptmx.Read(buf)
				if err != nil {
					return
				}
				vt.Write(buf[:n]...)
			}
		}()
	}
}

func (vt *VT100) Up() {
	n := 1
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	vt.Cursor.Y -= n
	if vt.Cursor.Y < 0 {
		vt.Cursor.Y = 0
	}
	vt.clearBuf()
}

func (vt *VT100) Down() {
	n := 1
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	vt.Cursor.Y += n
	if vt.Cursor.Y >= vt.Rows {
		vt.Cursor.Y = vt.Rows - 1
	}
	vt.clearBuf()
}

func (vt *VT100) Right() {
	n := 1
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	vt.Cursor.X += n
	if vt.Cursor.X >= vt.Cols {
		vt.Cursor.X = vt.Cols - 1
	}
	vt.clearBuf()
}

func (vt *VT100) Left() {
	n := 1
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	vt.Cursor.X -= n
	if vt.Cursor.X < 0 {
		vt.Cursor.X = 0
	}
	vt.clearBuf()
}

func (vt *VT100) JumpCursor() {
	y, x := 1, 1
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d;%d", &y, &x)
	}
	vt.Cursor.Y = y - 1
	if vt.Cursor.Y < 0 {
		vt.Cursor.Y = 0
	}
	if vt.Cursor.Y >= vt.Rows {
		vt.Cursor.Y = vt.Rows - 1
	}
	vt.Cursor.X = x - 1
	if vt.Cursor.X < 0 {
		vt.Cursor.X = 0
	}
	if vt.Cursor.X >= vt.Cols {
		vt.Cursor.X = vt.Cols - 1
	}
	vt.clearBuf()
}

func (vt *VT100) clearRow() {
	n := 0
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	switch n {
	case 0:
		for j := vt.Cursor.X; j < vt.Cols; j++ {
			vt.Screen[vt.Cursor.Y][j] = Cell{Char: ' ', Style: vt.currentStyle}
		}
	case 1:
		for j := 0; j <= vt.Cursor.X; j++ {
			vt.Screen[vt.Cursor.Y][j] = Cell{Char: ' ', Style: vt.currentStyle}
		}
	case 2:
		for j := 0; j < vt.Cols; j++ {
			vt.Screen[vt.Cursor.Y][j] = Cell{Char: ' ', Style: vt.currentStyle}
		}
	}
	vt.clearBuf()
}

func (vt *VT100) ClearScreen() {
	n := 0
	if len(vt.buf) > 0 {
		fmt.Sscanf(string(vt.buf), "%d", &n)
	}
	switch n {
	case 0:
		for i := vt.Cursor.Y; i < vt.Rows; i++ {
			for j := 0; j < vt.Cols; j++ {
				vt.Screen[i][j] = Cell{Char: ' ', Style: vt.currentStyle}
			}
		}
	case 1:
		for i := 0; i <= vt.Cursor.Y; i++ {
			for j := 0; j < vt.Cols; j++ {
				vt.Screen[i][j] = Cell{Char: ' ', Style: vt.currentStyle}
			}
		}
	case 2:
		for i := 0; i < vt.Rows; i++ {
			for j := 0; j < vt.Cols; j++ {
				vt.Screen[i][j] = Cell{Char: ' ', Style: vt.currentStyle}
			}
		}
	}
	vt.clearBuf()
}

func (vt *VT100) Write(p ...byte) (n int, err error) {
	for _, b := range p {
		vt.eat(b)
	}
	return len(p), nil
}

func (vt *VT100) eat(p byte) {
	vt.SignalRecord.Push(p)
	switch p {
	case 0x00, 0x05, 0x07, 0x11, 0x13:
		return
	case 0x18, 0x1A:
		vt.clearBuf()
		vt.Status = StatusNormal
		return
	}
	switch vt.Status {
	case StatusNormal:
		vt.normal(p)
	case StatusEscape:
		vt.escape(p)
	case StatusCSI:
		vt.csi(p)
	case StatusOSC:
		vt.osc(p)
	}
}

func (vt *VT100) scrollDown() {
	vt.Screen = vt.Screen[1:]
	newRow := make([]Cell, vt.Cols)
	for i := range newRow {
		newRow[i] = Cell{Char: ' ', Style: vt.currentStyle}
	}
	vt.Screen = append(vt.Screen, newRow)
	vt.Cursor.Y = vt.Rows - 1
}

func (vt *VT100) normal(p byte) {
	switch p {
	case '\a':
		// Bell, do nothing
	case '\b':
		if vt.Cursor.X > 0 {
			vt.Cursor.X--
		}
	case '\t':
		// Tab, move to next tab stop (every 8 columns)
		vt.Cursor.X = (vt.Cursor.X/8 + 1) * 8
		if vt.Cursor.X >= vt.Cols {
			vt.Cursor.X = vt.Cols - 1
		}
	case '\n':
		vt.index()
	case '\v':
		vt.Cursor.Y++
		if vt.Cursor.Y >= vt.Rows {
			vt.scrollDown()
		}
		vt.Cursor.X = 0
	case '\f':
		vt.Screen = newScreen(vt.Cols, vt.Rows)
		vt.Cursor.X = 0
		vt.Cursor.Y = 0
	case '\r':
		vt.Cursor.X = 0
	case 0x0E:
		// Shift Out, ignore for now
	case 0x0F:
		// Shift In, ignore for now
	case 0x18:
		// Cancel, ignore for now
	case 0x1A:
		// Substitute, ignore for now
	case 0x1B:
		vt.Status = StatusEscape
	case 0x00:
		// Null, ignore for now
	case 0x7F:
		// Delete, ignore for now
	default:
		vt.putChar(p)
	}
}

func (vt *VT100) putChar(c byte) {
	vt.buf = append(vt.buf, c)
	if utf8.FullRune(vt.buf) {
		r, i := utf8.DecodeRune(vt.buf)
		vt.buf = vt.buf[i:]
		vt.clearBuf()
		if vt.Cursor.X < vt.Cols && vt.Cursor.Y < vt.Rows {
			vt.Screen[vt.Cursor.Y][vt.Cursor.X] = Cell{Char: r, Style: vt.currentStyle}
		}
		vt.Cursor.X++
		if vt.Cursor.X >= vt.Cols {
			vt.Cursor.X = 0
			vt.Cursor.Y++
			if vt.Cursor.Y >= vt.Rows {
				vt.scrollDown()
			}
		}
	}
}

func (vt *VT100) index() {
	if vt.Cursor.Y < vt.Rows-1 {
		vt.Cursor.Y++
	} else {
		for i := range vt.Rows - 1 {
			vt.Screen[i] = vt.Screen[i+1]
		}
		vt.Screen[vt.Rows-1] = newLine(vt.Cols)
	}
}

func (vt *VT100) reverseIndex() {
	if vt.Cursor.Y > 0 {
		vt.Cursor.Y--
	} else {
		for i := range vt.Rows - 1 {
			vt.Screen[i+1] = vt.Screen[i]
		}
		vt.Screen[0] = newLine(vt.Cols)
	}
}

func (vt *VT100) escape(p byte) {
	switch p {
	case 0x1B:
		vt.clearBuf()
	case '[':
		vt.Status = StatusCSI
	case ']':
		vt.Status = StatusOSC
	case 'D':
		vt.index()
	case 'E':
		panic("Next Line")
	case 'H':
		panic("Tab Set")
	case 'M':
		vt.reverseIndex()
	case '7':
		vt.saveCursor()
		vt.Status = StatusNormal
	case '8':
		vt.restoreCursor()
		vt.Status = StatusNormal
	}
}

func (vt *VT100) osc(p byte) {
	if p == '\a' {
		vt.clearBuf()
		vt.Status = StatusNormal
		return
	}
	vt.buf = append(vt.buf, p)
}

func (vt *VT100) csi(p byte) {
	switch p {
	case 'A':
		vt.Status = StatusNormal
		vt.Up()
	case 'B':
		vt.Status = StatusNormal
		vt.Down()
	case 'C':
		vt.Status = StatusNormal
		vt.Right()
	case 'D':
		vt.Status = StatusNormal
		vt.Left()
	case 'm':
		// 修改 currentStyle
		params := parseParams(string(vt.buf))
		vt.currentStyle = vt.currentStyle.Set(params)
		vt.clearBuf()
		vt.Status = StatusNormal
	case 's':
		vt.saveCursor()
		vt.Status = StatusNormal
	case 'u':
		vt.restoreCursor()
		vt.Status = StatusNormal
	case 'J':
		vt.ClearScreen()
		vt.Status = StatusNormal
	case 'K':
		vt.clearRow()
		vt.Status = StatusNormal
	case 'H', 'f':
		vt.JumpCursor()
		vt.Status = StatusNormal
	case 'l', 'h':
		// Hide cursor, ignore for now
		vt.Status = StatusNormal
		vt.clearBuf()
	default:
		vt.buf = append(vt.buf, p)
	}
}

func (vt *VT100) clearBuf() {
	vt.buf = nil
}

// parseParams 解析 CSI 序列中的參數字串，返回整數切片
func parseParams(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ";")
	params := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			params = append(params, 0)
		} else if n, err := strconv.Atoi(part); err == nil {
			params = append(params, n)
		}
	}
	return params
}

func (vt *VT100) CursorString() string {
	return fmt.Sprintf("cursor=%d,%d", vt.Cursor.X+1, vt.Cursor.Y+1)
}

func (vt *VT100) SizeString() string {
	return fmt.Sprintf("size=%dx%d", vt.Cols, vt.Rows)
}

func (vt *VT100) header() string {
	s := fmt.Sprintf("%s %s\n", vt.SizeString(), vt.CursorString())
	for range vt.Cols {
		s += "-"
	}
	return s + "\n"
}

func (vt *VT100) String() string {
	s := vt.header()
	for _, row := range vt.Screen {
		for _, cell := range row {
			s += string(cell.Char)
		}
		s += "\n"
	}
	return s
}

func (vt *VT100) WritePty(b []byte) {
	vt.ptmx.Write(b)
}

func (vt *VT100) Start(cmd *exec.Cmd) {
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(vt.Rows), Cols: uint16(vt.Cols)})
	vt.ptmx = ptmx
	vt.cmd = cmd
	if err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}
	buf := make([]byte, 4096)
	go func() {
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				return
			}
			vt.Write(buf[:n]...)
		}
	}()
}

func (vt *VT100) StringWithIndex() string {
	s := ""
	for i, row := range vt.Screen {
		s += fmt.Sprint(i+1) + ": "
		for _, cell := range row {
			s += string(cell.Char)
		}
		s += "\n"
	}
	return s
}

func New(cols, rows int) *VT100 {
	record := NewBoundedList[byte](1000)
	logger := NewBoundedList[string](1000)
	cells := newScreen(cols, rows)
	return &VT100{
		Cols:         cols,
		Rows:         rows,
		Cursor:       Cursor{X: 0, Y: 0},
		Screen:       cells,
		currentStyle: DefaultStyle(),
		SignalRecord: record,
		Log:          logger,
	}
}
