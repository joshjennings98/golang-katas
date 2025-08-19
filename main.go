//go:build wasm

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
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

type caseResult struct {
	Call     string `json:"call"`
	Expected string `json:"expected"`
	Got      string `json:"got"`
	Ok       bool   `json:"ok"`
}

type runResult struct {
	Stdout string       `json:"stdout"`
	Stderr string       `json:"stderr"`
	Total  int          `json:"total"`
	Passed int          `json:"passed"`
	Cases  []caseResult `json:"cases"`
}

var (
	doc       = js.Global().Get("document")
	win       = js.Global().Get("window")
	perf      = js.Global().Get("performance")
	startedAt = 0.0
	successAt float64

	srcEl         js.Value
	outEl         js.Value
	statusEl      js.Value
	runBtn        js.Value
	clearBtn      js.Value
	kataSel       js.Value
	kataDesc      js.Value
	defaultHeight float64

	katas KatasConfig
)

func q(sel string) js.Value         { return doc.Call("querySelector", sel) }
func setText(el js.Value, s string) { el.Set("textContent", s) }
func setVal(el js.Value, s string)  { el.Set("value", s) }

func getComputedHeightPx(el js.Value) float64 {
	cs := win.Call("getComputedStyle", el)
	h := cs.Get("height").String()
	h = strings.TrimSuffix(h, "px")
	v, _ := strconv.ParseFloat(h, 64)
	return v
}

func disableControls() {
	runBtn.Set("disabled", true)
	clearBtn.Set("disabled", true)
}

func b64ToUTF8(b64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func formatElapsed(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.0f ms", ms)
	}
	return fmt.Sprintf("%.2f s", ms/1000)
}

func writeOut(s string) {
	setText(outEl, s)
	resizeOutToContent()
}

func setOutSmart(s string) {
	isSuccess := s == "✓ Success!" || strings.HasPrefix(s, "✓ Success!")
	if isSuccess {
		if successAt == 0 && startedAt > 0 {
			v := perf.Call("now").Float() - startedAt
			successAt = v
		}
		if successAt != 0 {
			writeOut("✓ Success! (" + formatElapsed(successAt) + ")")
			disableControls()
			return
		}
	}
	writeOut(s)
}

func resizeOutToContent() {
	content := outEl.Get("textContent").String()
	isEmpty := strings.TrimSpace(content) == ""
	if isEmpty {
		outEl.Get("style").Set("height", fmt.Sprintf("%.0fpx", defaultHeight))
		outEl.Get("style").Set("overflowY", "hidden")
		return
	}
	outEl.Get("style").Set("height", "auto")
	needed := outEl.Get("scrollHeight").Float()
	h := needed
	if h > defaultHeight {
		h = defaultHeight
	}
	outEl.Get("style").Set("height", fmt.Sprintf("%.0fpx", h))
	if needed > defaultHeight {
		outEl.Get("style").Set("overflowY", "auto")
	} else {
		outEl.Get("style").Set("overflowY", "hidden")
	}
}

func loadKatasFromGlobal() error {
	kb64 := js.Global().Get("KATAS_B64").String()
	jsonStr, err := b64ToUTF8(kb64)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(jsonStr), &katas); err != nil {
		return err
	}
	return nil
}

func populateKataSelect() {
	kataSel.Set("innerHTML", "")
	for _, k := range katas.Katas {
		opt := doc.Call("createElement", "option")
		opt.Set("value", k.Slug)
		opt.Set("textContent", k.Title)
		kataSel.Call("appendChild", opt)
	}

	if l := kataSel.Get("options").Get("length").Int(); l > 0 {
		idx := rand.Intn(l)
		kataSel.Set("selectedIndex", idx)
		setKata(kataSel.Get("value").String())
	}
}

func findKata(slug string) *Kata {
	for i := range katas.Katas {
		if katas.Katas[i].Slug == slug {
			return &katas.Katas[i]
		}
	}
	return nil
}

func setKata(slug string) {
	k := findKata(slug)
	if k == nil {
		return
	}
	setVal(kataSel, k.Slug)
	setText(kataDesc, k.Description)
	setVal(srcEl, k.VisibleSkeleton)
}

func runTests(src, slug string) runResult {
	var out bytes.Buffer
	i := interp.New(interp.Options{Stdout: &out, Stderr: &out})
	i.Use(stdlib.Symbols)

	userCode := strings.TrimSpace(src)
	if !strings.HasPrefix(userCode, "package ") {
		userCode = `package main

import (
	_fmt "fmt"
	_json "encoding/json"
)
` + userCode
	}

	k := findKata(slug)
	if k == nil {
		return runResult{Stdout: out.String(), Stderr: "kata not found\n"}
	}

	var h strings.Builder
	h.WriteString(`
type __Case struct {
	Call	 string ` + "`json:\"call\"`" + `
	Expected string ` + "`json:\"expected\"`" + `
	Got	  string ` + "`json:\"got\"`" + `
	Ok	   bool   ` + "`json:\"ok\"`" + `
}
var __cases = []__Case{
`)
	for _, tc := range k.TestCases {
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
		Total  int	  ` + "`json:\"total\"`" + `
		Passed int	  ` + "`json:\"passed\"`" + `
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

	if _, err := i.Eval(allCode); err != nil {
		return runResult{
			Stdout: out.String(),
			Stderr: fmt.Sprintf("compile error: %v\n", err),
			Total:  len(k.TestCases),
			Passed: 0,
			Cases:  nil,
		}
	}

	stdoutQ := strconv.Quote(out.String())
	v, err := i.Eval(`__resultJSON(` + stdoutQ + `)`)
	if err != nil {
		return runResult{
			Stdout: out.String(),
			Stderr: fmt.Sprintf("result error: %v\n", err),
			Total:  len(k.TestCases),
			Passed: 0,
			Cases:  nil,
		}
	}

	var rr runResult
	_ = json.Unmarshal([]byte(v.String()), &rr)
	return rr
}

func onRun(_ js.Value, _ []js.Value) any {
	runBtn.Set("disabled", true)
	setText(statusEl, "Running...")
	setOutSmart("")

	src := srcEl.Get("value").String()
	slug := kataSel.Get("value").String()

	res := runTests(src, slug)

	lines := make([]string, 0, 8+len(res.Cases))
	if res.Stdout != "" {
		lines = append(lines, res.Stdout)
	}
	if res.Stderr != "" {
		lines = append(lines, res.Stderr)
	}
	if res.Total > 0 {
		lines = append(lines, fmt.Sprintf("Kata: %s - %d/%d passed", slug, res.Passed, res.Total))
	}
	for _, c := range res.Cases {
		mark := "✓"
		if !c.Ok {
			mark = "✗"
		}
		line := fmt.Sprintf("%s %s => got %s (expected %s)", mark, c.Call, c.Got, c.Expected)
		lines = append(lines, line)
	}
	out := strings.Join(lines, "\n")

	success := !(strings.Contains(out, "✗") || strings.Contains(out, "compile error:"))
	if success {
		setOutSmart("✓ Success!")
		disableControls()
	} else {
		setOutSmart(out)
		runBtn.Set("disabled", false)
	}

	setText(statusEl, "Ready")
	return nil
}

func onClear(_ js.Value, _ []js.Value) any {
	setText(outEl, "")
	resizeOutToContent()
	return nil
}

func onKataChange(_ js.Value, _ []js.Value) any {
	setKata(kataSel.Get("value").String())
	return nil
}

func onResize(_ js.Value, _ []js.Value) any {
	defaultHeight = getComputedHeightPx(outEl)
	resizeOutToContent()
	return nil
}

const indent = "	"

func handleKeydown(this js.Value, args []js.Value) any {
	e := args[0]
	key := e.Get("key").String()
	if key != "Tab" && key != "Enter" {
		return nil
	}

	el := e.Get("target")
	value := el.Get("value").String()
	s := el.Get("selectionStart").Int()
	eend := el.Get("selectionEnd").Int()

	set := func(text string, start, end int) {
		el.Set("value", text)
		el.Set("selectionStart", start)
		el.Set("selectionEnd", end)
	}

	if key == "Tab" {
		e.Call("preventDefault")

		startLine := strings.LastIndex(value[:s], "\n") + 1
		endLineBreak := strings.Index(value[eend:], "\n")
		endLine := 0
		if endLineBreak == -1 {
			endLine = len(value)
		} else {
			endLine = eend + endLineBreak
		}

		block := value[startLine:endLine]
		multiline := s != eend && strings.Contains(block, "\n")

		shift := e.Get("shiftKey").Bool()
		if multiline {
			lines := strings.Split(block, "\n")
			if shift {
				removedTotal := 0
				for i := range lines {
					if strings.HasPrefix(lines[i], indent) {
						lines[i] = lines[i][len(indent):]
						removedTotal += len(indent)
					} else if len(lines[i]) > 0 && (lines[i][0] == ' ' || lines[i][0] == '\t') {
						lines[i] = lines[i][1:]
						removedTotal += 1
					}
				}
				out := strings.Join(lines, "\n")
				before := value[:startLine]
				after := value[endLine:]
				newText := before + out + after
				newS := s - min(len(indent), s-startLine)
				newE := eend - removedTotal
				set(newText, newS, newE)
			} else {
				for i := range lines {
					lines[i] = indent + lines[i]
				}
				out := strings.Join(lines, "\n")
				before := value[:startLine]
				after := value[endLine:]
				newText := before + out + after
				added := len(indent) * len(lines)
				newS := s + len(indent)
				newE := eend + added
				set(newText, newS, newE)
			}
			return nil
		}

		if shift {
			lineStart := strings.LastIndex(value[:s], "\n") + 1
			if strings.HasPrefix(value[lineStart:], indent) {
				set(value[:lineStart]+value[lineStart+len(indent):], s-len(indent), eend-len(indent))
			} else if lineStart < len(value) && (value[lineStart] == ' ' || value[lineStart] == '\t') {
				set(value[:lineStart]+value[lineStart+1:], max(lineStart, s-1), max(lineStart, eend-1))
			}
		} else {
			before := value[:s]
			after := value[eend:]
			set(before+indent+after, s+len(indent), s+len(indent))
		}
		return nil
	}

	if key == "Enter" {
		e.Call("preventDefault")
		lineStart := strings.LastIndex(value[:s], "\n") + 1
		curLine := value[lineStart:s]
		carry := leadingWS(curLine)
		before := value[:s]
		after := value[eend:]
		insert := "\n" + carry
		set(before+insert+after, s+len(insert), s+len(insert))
		return nil
	}

	return nil
}

func leadingWS(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[:i]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	srcEl = q("#src")
	outEl = q("#out")
	statusEl = q("#status")
	runBtn = q("#runBtn")
	clearBtn = q("#clearBtn")
	kataSel = q("#kataSelect")
	kataDesc = q("#kataDesc")

	startedAt = perf.Call("now").Float()
	defaultHeight = getComputedHeightPx(outEl)

	runBtn.Call("addEventListener", "click", js.FuncOf(onRun))
	clearBtn.Call("addEventListener", "click", js.FuncOf(onClear))
	kataSel.Call("addEventListener", "change", js.FuncOf(onKataChange))
	win.Call("addEventListener", "resize", js.FuncOf(onResize))
	srcEl.Call("addEventListener", "keydown", js.FuncOf(handleKeydown))

	setText(statusEl, "Loading WASM...")
	if err := loadKatasFromGlobal(); err != nil {
		setText(statusEl, "Error loading katas")
		writeOut(fmt.Sprintf("bad katas.json: %v\n", err))
		return
	}

	setText(statusEl, "Ready")
	runBtn.Set("disabled", false)
	populateKataSelect()

	clearBtn.Call("addEventListener", "click", js.FuncOf(onClear))

	select {}
}
