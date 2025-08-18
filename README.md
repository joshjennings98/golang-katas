# Standalone Golang Katas

This project builds a **self-contained HTML file** that runs Go code in the browser via [Yaegi](https://github.com/traefik/yaegi).
The final HTML embeds:

* `wasm_exec.js` (Go’s WebAssembly runtime shim)
* `yaegi.wasm` (compiled from your `main.go`) as base64
* `katas.json` (challenge config) as base64
* Your HTML/JS/CSS template

Open `dist/index.html` directly from disk — no server or extra assets needed.

## What’s new

* **Kata dropdown** with description and a visible code skeleton.
* **In-browser tests**: user code + generated harness are evaluated **once**; each test calls your functions directly.
* **Results panel** shows per-case pass/fail and a summary.

## How it works

1. **Template**
   `template.html` contains three placeholders:

   * `/*__WASM_EXEC_JS__*/` → replaced with `wasm_exec.js`
   * `__WASM_BASE64__` → replaced with base64 of `yaegi.wasm`
   * `__KATAS_BASE64__` → replaced with base64 of `katas.json`

   Example snippet:

   ```html
   <script>
   /* ===== BEGIN wasm_exec.js ===== */
   /*__WASM_EXEC_JS__*/
   /* ===== END wasm_exec.js ===== */
   </script>

   <script>
     const WASM_B64  = "__WASM_BASE64__";
     const KATAS_B64 = "__KATAS_BASE64__";
   </script>
   ```

2. **Interpreter (WASM)**
   `main.go` exports `yaegiRunTests(src, slug, katasJSON)` to JS:

   * Ensures `src` is in `package main` (wraps if needed).
   * Parses `katasJSON`, finds the selected `slug`.
   * Generates a small **harness** alongside the user code that:

     * Computes `Got := fmt.Sprint(<call>)` for each test
     * Compares `Got` vs `Expected` (string comparison)
     * Returns a JSON string of `{stdout, stderr, total, passed, cases[]}`

   JS evaluates **once** (user code + harness), then calls the exported helper to get the JSON result.

3. **Makefile**

   * Builds `yaegi.wasm` (`GOOS=js GOARCH=wasm`)
   * Base64-encodes it to `dist/yaegi.wasm.b64`
   * Copies `wasm_exec.js`
   * Base64-encodes `katas.json` to `dist/katas.json.b64`
   * Splices all into `template.html` → `dist/index.html` via `awk`

## Usage

```sh
make build
```

Artifacts:

* `dist/yaegi.wasm` — raw WebAssembly binary
* `dist/yaegi.wasm.b64` — base64 (used during substitution)
* `dist/wasm_exec.js` — Go runtime shim
* `dist/katas.json.b64` — base64 of your challenges
* `dist/index.html` — **final single-file app**

Clean:

```sh
make clean
```

Open:

* Double-click `dist/index.html`, or serve it with any static file server.

## Katas configuration (`katas.json`)

Schema:

```json
{
  "katas": [
    {
      "title": "Add Two Numbers",
      "slug": "add-two-numbers",
      "description": "Return the sum of two integers.",
      "visible_skeleton": "func Add(a int, b int) int {\n    \n    return 0\n}\n",
      "test_cases": [
        { "call": "Add(1, 2)", "expected": "3" }
      ]
    }
  ]
}
```

* `visible_skeleton` is injected into the editor when the kata is selected.
* Each test `call` is evaluated inside the same interpreter context as the user code.
* `expected` is compared to `fmt.Sprint(<call>)`.

> **Typed comparisons?** Extend your JSON to include a `type` (e.g., `"int64"`) and adjust the harness to cast before `fmt.Sprint` or compare typed values directly.

## Template behavior (UI)

* Dropdown chooses the kata and populates:

  * Description
  * Code skeleton
* **Run** executes tests in the browser; results list each case and a summary.
* Output panel auto-sizes up to its default height; scrolls beyond.

## Notes & limitations

* On `GOARCH=wasm`, `int` is **32-bit**. Tests like `2147483647 + 1` overflow an `int`. Prefer `int64` in the skeleton/tests if you need larger ranges.
* Yaegi in the browser: most pure-Go code works; anything requiring unsupported syscalls/APIs won’t.
* The harness uses `fmt.Sprint` so complex values stringify consistently.

## Troubleshooting

* **`undefined: <Name>`** — The function isn’t defined in the user code, or name mismatch with tests.
* **Interpreter compile error** — You’ll see it in `stderr` in the results panel.
* **`ValueOf: invalid value`** — We avoid this by JSON-roundtripping results before returning them to JS.

## Placeholders (must match exactly)

* `/*__WASM_EXEC_JS__*/`
* `__WASM_BASE64__`
* `__KATAS_BASE64__`

---

I used AI to create a tool to help me become a better programmer!
