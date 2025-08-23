.PHONY: single split clean error

DIST := dist

HTML_SINGLE_IN := ./template.single.html
HTML_SPLIT_IN := ./template.split.html
GOLANG_IN := ./main.go
KATAS_JSON := ./katas.json
STYLE_IN := ./style.css

HTML_OUT := $(DIST)/index.html
WASM_OUT := $(DIST)/yaegi.wasm
WASM_B64 := $(DIST)/yaegi.wasm.b64
WASM_EXEC_OUT := $(DIST)/wasm_exec.js
KATAS_B64 := $(DIST)/katas.json.b64

$(DIST):
	mkdir -p $(DIST)

$(WASM_OUT): $(GOLANG_IN) | $(DIST)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT)

$(WASM_B64): $(WASM_OUT) | $(DIST)
	base64 < "$(WASM_OUT)" | tr -d '\n' > "$(WASM_B64)"

$(WASM_EXEC_OUT): | $(DIST)
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js $(WASM_EXEC_OUT)

$(KATAS_B64): $(KATAS_JSON) | $(DIST)
	base64 < "$(KATAS_JSON)" | tr -d '\n' > "$(KATAS_B64)"

single: $(HTML_SINGLE_IN) $(STYLE_IN) $(WASM_B64) $(WASM_EXEC_OUT) $(KATAS_B64) | $(DIST)
	cp "$(STYLE_IN)" "$(DIST)/style.css"
	@awk -v b64f="$(WASM_B64)" -v execf="$(WASM_EXEC_OUT)" -v katasf="$(KATAS_B64)" '\
function putfile_index_nolf(f,   L){ while ((getline L < f) > 0) printf "%s", L; close(f) } \
function putfile_lines(f,        L){ while ((getline L < f) > 0) print L; close(f) } \
BEGIN{ tokB="__WASM_BASE64__"; tokE="/*__WASM_EXEC_OUT_JS__*/"; tokK="__KATAS_BASE64__" } \
{ line=$$0; \
  p=index(line, tokB); if (p){ pre=substr(line,1,p-1); post=substr(line,p+length(tokB)); printf "%s", pre; putfile_index_nolf(b64f); print post; next } \
  p=index(line, tokE); if (p){ pre=substr(line,1,p-1); post=substr(line,p+length(tokE)); printf "%s", pre; putfile_lines(execf); print post; next } \
  p=index(line, tokK); if (p){ pre=substr(line,1,p-1); post=substr(line,p+length(tokK)); printf "%s", pre; putfile_index_nolf(katasf); print post; next } \
  print }' "$(HTML_SINGLE_IN)" > "$(HTML_OUT)"; \
	echo "Wrote $(HTML_OUT)"

split: $(HTML_SPLIT_IN) $(STYLE_IN) $(WASM_B64) $(WASM_EXEC_OUT) $(KATAS_B64) | $(DIST)
	cp "$(STYLE_IN)" "$(DIST)/style.css"
	printf 'globalThis.WASM_B64="%s";\n' "$$(cat "$(WASM_B64)")" > "$(DIST)/wasm_embed.js"
	printf 'globalThis.KATAS_B64="%s";\n' "$$(cat "$(KATAS_B64)")" > "$(DIST)/katas_embed.js"
	cp "$(HTML_SPLIT_IN)" "$(HTML_OUT)"
	echo "Wrote $(HTML_OUT) and assets to $(DIST)/"

clean:
	rm -rf $(DIST)

.DEFAULT_GOAL := error

error:
	@echo "Targets:"
	@echo "  make single"
	@echo "  make split"
	@echo "  make clean"
	@exit 1
