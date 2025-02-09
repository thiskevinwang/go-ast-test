package main

import (
	"bytes"
	"fmt"
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// window.interpret?.(val);
func interpret(this js.Value, args []js.Value) interface{} {
	// Capture output from the interpreter to be returned to the caller
	output := &bytes.Buffer{}

	// Create a new interpreter
	i := interp.New(interp.Options{
		Stdout: output,
		Stderr: output,
	})

	// Use the Go standard library in the interpreter
	i.Use(stdlib.Symbols)

	// The Go code to be interpreted, passed as a string from javascript -> WASM
	code := args[0].String()

	// Interpret and run the Go code
	if _, err := i.Eval(code); err != nil {
		return fmt.Sprintf("Error interpreting code:", err)
	}

	// Return captured output
	return js.ValueOf(output.String())
}
