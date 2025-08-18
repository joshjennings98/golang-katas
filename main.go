//go:build wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type TestCase struct {
	Call     string `json:"call"`
	Expected string `json:"expected"`
}

type Kata struct {
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Description     string     `json:"description"`
	VisibleSkeleton string     `json:"visible_skeleton"`
	TestCases       []TestCase `json:"test_cases"`
}

type KatasConfig struct {
	Katas []Kata `json:"katas"`
}

// runTests(src, slug, katasJSON) => {passed,total,cases:[...],stdout,stderr}
func runTests(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return js.ValueOf(map[string]any{"stderr": "invalid args\n"})
	}
	code := args[0].String()
	slug := args[1].String()
	cfgJSON := args[2].String()

	var out bytes.Buffer
	i := interp.New(interp.Options{Stdout: &out, Stderr: &out})
	i.Use(stdlib.Symbols)

	// 1) Normalize user code into package main if needed.
	userCode := strings.TrimSpace(code)
	if !strings.HasPrefix(userCode, "package ") {
		userCode = `
package main

import (
	_fmt "fmt"
	_json "encoding/json"
)

` + userCode
	}

	// 2) Parse katas JSON to find the selected kata.
	var cfg KatasConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return js.ValueOf(map[string]any{"stdout": out.String(), "stderr": fmt.Sprintf("bad katas.json: %v\n", err)})
	}
	var kata *Kata
	for idx := range cfg.Katas {
		if cfg.Katas[idx].Slug == slug {
			kata = &cfg.Katas[idx]
			break
		}
	}
	if kata == nil {
		return js.ValueOf(map[string]any{"stdout": out.String(), "stderr": "kata not found\n"})
	}

	// 3) Generate a harness that:
	//    - Imports fmt/json under alias names to avoid collisions.
	//    - Builds __cases with Got := fmt.Sprint(<call>) computed at runtime.
	//    - Grades cases and exposes __resultJSON(stdout string) -> string.
	var h strings.Builder
	h.WriteString(`
type __Case struct {
	Call     string ` + "`json:\"call\"`" + `
	Expected string ` + "`json:\"expected\"`" + `
	Got      string ` + "`json:\"got\"`" + `
	Ok       bool   ` + "`json:\"ok\"`" + `
}
var __cases = []__Case{
`)
	for _, tc := range kata.TestCases {
		// Each entry computes the result by actually calling user's function(s).
		fmt.Fprintf(&h, `{Call:%q, Expected:%q, Got:_fmt.Sprint(%s)},`+"\n", tc.Call, tc.Expected, tc.Call)
	}
	h.WriteString(`}
func __grade() {
	for i := range __cases {
		if __cases[i].Got == __cases[i].Expected {
			__cases[i].Ok = true
		}
	}
}
func __resultJSON(stdout string) string {
	__grade()
	passed := 0
	for _, c := range __cases { if c.Ok { passed++ } }
	type Res struct {
		Stdout string   ` + "`json:\"stdout\"`" + `
		Stderr string   ` + "`json:\"stderr\"`" + `
		Total  int      ` + "`json:\"total\"`" + `
		Passed int      ` + "`json:\"passed\"`" + `
		Cases  []__Case ` + "`json:\"cases\"`" + `
	}
	b, _ := _json.Marshal(Res{
		Stdout: stdout,
		Stderr: "",
		Total:  len(__cases),
		Passed: passed,
		Cases:  __cases,
	})
	return string(b)
}
`)

	allCode := userCode + "\n" + h.String()

	// 4) Evaluate once: user code + harness together.
	if _, err := i.Eval(allCode); err != nil {
		// Compilation/runtime error in user code or harness.
		res := map[string]any{
			"stdout": out.String(),
			"stderr": fmt.Sprintf("compile error: %v\n", err),
			"total":  len(kata.TestCases),
			"passed": 0,
			"cases":  []any{},
		}
		b, _ := json.Marshal(res)
		return js.Global().Get("JSON").Call("parse", string(b))
	}

	// 5) Fetch JSON result from the evaluated snippet (single cheap call).
	//    Quote stdout so it’s a proper Go string literal.
	stdoutQ := strconv.Quote(out.String())
	v, err := i.Eval(`__resultJSON(` + stdoutQ + `)`)
	if err != nil {
		res := map[string]any{
			"stdout": out.String(),
			"stderr": fmt.Sprintf("result error: %v\n", err),
			"total":  len(kata.TestCases),
			"passed": 0,
			"cases":  []any{},
		}
		b, _ := json.Marshal(res)
		return js.Global().Get("JSON").Call("parse", string(b))
	}

	// v is a Go string containing JSON
	jsonStr := v.String()
	return js.Global().Get("JSON").Call("parse", jsonStr)
}

func main() {
	js.Global().Set("yaegiRunTests", js.FuncOf(runTests))
	select {}
}
