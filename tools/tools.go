package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/chenyunda218/alterminal-mcp/vt100"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type toolsOption struct {
	prefix string
}

func WithPrefix(prefix string) func(option toolsOption) toolsOption {
	return func(option toolsOption) toolsOption {
		option.prefix = prefix
		return option
	}
}

func ServeStdio(vt *vt100.VT100, options func(option toolsOption) toolsOption) {
	option := toolsOption{}
	s := server.NewMCPServer(
		"Alterminal",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithDescription(description()),
	)
	screenshotTool := mcp.NewTool(option.prefix+"screenshot",
		mcp.WithDescription("[Alterminal] Show current terminal status. return screenshot."),
		mcp.WithString("format",
			mcp.DefaultString("html"),
			mcp.Enum("text", "html")),
	)

	s.AddTool(screenshotTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, _ := request.RequireString("format")
		switch format {
		case "text":
			return mcp.NewToolResultText(vt.String()), nil
		}
		return mcp.NewToolResultText(vt.ToHTML()), nil
	})

	getSizeTool := mcp.NewTool(option.prefix+"get_window_size",
		mcp.WithDescription("[Alterminal] Get the current terminal window size. Format is `size=<cols>x<rows>`"),
	)
	s.AddTool(getSizeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(vt.SizeString()), nil
	})

	writeTool := mcp.NewTool(option.prefix+"write",
		mcp.WithDescription("[Alterminal] Write text to terminal."),
		mcp.WithString("text", mcp.Required()),
		mcp.WithBoolean("with_enter", mcp.DefaultBool(false)),
	)
	s.AddTool(writeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, _ := request.RequireString("text")
		vt.WritePty([]byte(text))
		withEnter, _ := request.RequireBool("with_enter")
		if withEnter {
			vt.WritePty([]byte{byte(13)})
		}
		time.Sleep(1 * time.Second)
		return mcp.NewToolResultText("success"), nil
	})

	getCursor := mcp.NewTool(option.prefix+"get_cursor_position",
		mcp.WithDescription("[Alterminal] Get cursor position. Format is `cursor=<col>,<row>. col and row is start from 1.`"))
	s.AddTool(getCursor, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(vt.CursorString()), nil
	})

	resetTool := mcp.NewTool(option.prefix+"reset",
		mcp.WithDescription("[Alterminal] Reset terminal to initial state. Will return screenshot."))
	s.AddTool(resetTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		vt.Reset()
		time.Sleep(time.Second * 3)
		return mcp.NewToolResultText(vt.String()), nil
	})

	controlCodeTool := mcp.NewTool(option.prefix+"control_code",
		mcp.WithDescription("[Alterminal] Send control code to terminal."),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Enum("NUL", "ETX", "ENQ", "BEL", "BS", "HT", "LF", "VT", "FF", "CR", "SO", "SI", "DC1", "DC3", "CAN", "SUB", "ESC", "DEL")),
	)
	s.AddTool(controlCodeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		code, _ := request.RequireString("code")
		controlCodeWrite(vt, code)
		time.Sleep(time.Second * 1)
		return mcp.NewToolResultText("success"), nil
	})

	writeBytesTool := mcp.NewTool(option.prefix+"write_byte",
		mcp.WithDescription("[Alterminal] Write byte to terminal. Will return screenshot."),
		mcp.WithArray("byte", mcp.WithIntegerItems(), mcp.Required()),
	)
	s.AddTool(writeBytesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		b, _ := request.RequireInt("byte")
		vt.WritePty([]byte{byte(b)})
		time.Sleep(1 * time.Second)
		return mcp.NewToolResultText("success"), nil
	})
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func controlCodeWrite(t *vt100.VT100, code string) {
	switch code {
	case "NUL":
		t.WritePty([]byte{0x00})
	case "ETX":
		t.WritePty([]byte{0x03})
	case "ENQ":
		t.WritePty([]byte{0x05})
	case "BEL":
		t.WritePty([]byte{0x07})
	case "BS":
		t.WritePty([]byte{0x08})
	case "HT":
		t.WritePty([]byte{0x09})
	case "LF":
		t.WritePty([]byte{0x0A})
	case "VT":
		t.WritePty([]byte{0x0B})
	case "FF":
		t.WritePty([]byte{0x0C})
	case "CR":
		t.WritePty([]byte{0x0D})
	case "SO":
		t.WritePty([]byte{0x0E})
	case "SI":
		t.WritePty([]byte{0x0F})
	case "DC1":
		t.WritePty([]byte{0x11})
	case "DC3":
		t.WritePty([]byte{0x13})
	case "CAN":
		t.WritePty([]byte{0x18})
	case "SUB":
		t.WritePty([]byte{0x1A})
	case "ESC":
		t.WritePty([]byte{0x1B})
	case "DEL":
		t.WritePty([]byte{0x7F})
	}
}
func description() string {
	return `
# Alterminal
Alterminal 是一個VT100兼容的PTY終端, 你可以使用write和write_byte控制終端機.
`
}
