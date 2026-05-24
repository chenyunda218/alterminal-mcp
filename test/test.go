//go:build tool1
// +build tool1

package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	r, _ := utf8.DecodeRune([]byte{226, 150, 189})
	fmt.Println(string(r))
}
