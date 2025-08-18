//go:build wasm

package main

import (
	"bytes"
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func run(this js.Value, args []js.Value) any {
	code := args[0].String()
	var out bytes.Buffer
	i := interp.New(interp.Options{Stdout: &out, Stderr: &out})
	i.Use(stdlib.Symbols)
	_, err := i.Eval(code)
	// Return a JS object {stdout: string, stderr: string}
	// Yaegi writes to Stdout/Stderr we passed; if you want separate, split by tracking writes.
	res := map[string]any{"stdout": out.String(), "stderr": ""}
	if err != nil {
		res["stderr"] = err.Error() + "\n"
	}
	return js.ValueOf(res)
}

func main() {
	js.Global().Set("yaegiRun", js.FuncOf(run))
	select {} // keep runtime alive
}
