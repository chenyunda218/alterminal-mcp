package main

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/chenyunda218/alterminal-mcp/tools"
	"github.com/chenyunda218/alterminal-mcp/vt100"
)

func main() {
	mux := http.NewServeMux()
	cmd := exec.Command("bash", "--norc")
	rows, cols := uint16(24), uint16(80)
	vt := vt100.New(int(cols), int(rows))
	vt.Start(cmd)
	// 注册路由
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, vt.ToHTML())
	})
	mux.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, vt.String())
	})
	mux.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		s := ""
		vt.SignalRecord.ForEach(func(index int, value byte) {
			s += fmt.Sprintln(value)
		})
		fmt.Fprint(w, s)
	})
	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		s := ""
		vt.Log.ForEach(func(index int, value string) {
			s += fmt.Sprintln(value)
		})
		fmt.Fprint(w, s)
	})

	go func() {
		httpServer := &http.Server{
			Addr:    ":8081",
			Handler: mux,
		}
		httpServer.ListenAndServe()
	}()
	tools.ServeStdio(vt, tools.WithPrefix("alterminal"))

}
