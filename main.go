package main

import (
	"syscall/js"
)

const defaultSrc = `package foo

import (
	"fmt"
	"time"
)

func bar() {
	fmt.Println(time.Now())
}`

func main() {
	// Register a function in Go that can be called from JavaScript
	js.Global().Set("toAst", js.FuncOf(toAst))
	js.Global().Set("interpret", js.FuncOf(interpret))
	// prevent the Go program from exiting

	// https://github.com/norunners/vue/issues/40

	select {}
	// <-make(chan bool)
}
