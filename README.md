# Single-file Go+Yaegi WASM Runner

This project builds a **self-contained HTML file** that can run Go code in the browser via [Yaegi](https://github.com/traefik/yaegi).  
The final HTML file embeds:

- `wasm_exec.js` (Go’s WebAssembly runtime shim)  
- `yaegi.wasm` (compiled from your `main.go`) encoded as base64  
- Your HTML/JS/CSS template

That means you can open the HTML file directly from disk with no server and no extra assets.

## How it works

1. **Template**  
  You start with an `template.html` template containing two special placeholders:

  - `/*__WASM_EXEC_JS__*/` → replaced with the contents of `wasm_exec.js` from your Go toolchain  
  - `__WASM_BASE64__` → replaced with the base64-encoded contents of the compiled `yaegi.wasm`  

  Example snippet in the template:
  ```html
  <script>
  /* ===== BEGIN wasm_exec.js ===== */
  /*__WASM_EXEC_JS__*/
  /* ===== END wasm_exec.js ===== */
  </script>

  <script>
  const WASM_B64 = "__WASM_BASE64__";
  </script>
  ```

2. **Makefile**

  * Builds `yaegi.wasm` from `main.go` (`GOOS=js GOARCH=wasm`)
  * Base64-encodes the wasm into `dist/yaegi.wasm.b64`
  * Copies `wasm_exec.js` from your Go toolchain (location differs by platform / Go version)
  * Uses `awk` to splice both into the template and produce `dist/index.html`

3. **Result**
  `dist/index.html` is a single portable HTML file containing everything it needs.
  Open it in any modern browser and you can type Go code, run it through Yaegi, and see the output.

## Usage

```sh
make build
```

Artifacts:

* `dist/yaegi.wasm` — raw WebAssembly binary
* `dist/yaegi.wasm.b64` — base64 version (used during substitution)
* `dist/wasm_exec.js` — Go runtime shim
* `dist/inline.html` — **final single-file app**

Clean build:

```sh
make clean
```

## Notes

* Not all Go stdlib packages work under WebAssembly/browser. Pure compute, `fmt`, `strings`, etc. are fine.
* Placeholders must appear exactly as `/*__WASM_EXEC_JS__*/` and `__WASM_BASE64__` in the template.
