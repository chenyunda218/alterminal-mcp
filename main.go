package main

import (
	"log"
	"os/exec"

	"github.com/chenyunda218/alterminal-mcp/vt100"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cmd := exec.Command("bash", "--norc")
	rows, cols := uint16(24), uint16(80)
	vt := vt100.New(int(cols), int(rows))
	vt.Start(cmd)

}
