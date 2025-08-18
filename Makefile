.PHONY: build clean error
DIST := dist

HTML_IN := ./template.html
GOLANG_IN := ./main.go

HTML_OUT := $(DIST)/index.html
WASM_OUT := $(DIST)/yaegi.wasm
WASM_B64 := $(DIST)/yaegi.wasm.b64
WASM_EXEC_OUT := $(DIST)/wasm_exec.js

$(DIST):
	mkdir -p $(DIST)

$(WASM_OUT): $(GOLANG_IN) | $(DIST)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT)

$(WASM_B64): $(WASM_OUT) | $(DIST)
	base64 < "$(WASM_OUT)" | tr -d '\n' > "$(WASM_B64)"

$(WASM_EXEC_OUT): | $(DIST)
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js $(WASM_EXEC_OUT)

$(HTML_OUT): $(HTML_IN) $(WASM_B64) $(WASM_EXEC_OUT) | $(DIST)
	@awk -v b64f="$(WASM_B64)" -v execf="$(WASM_EXEC_OUT)" '\
function putfile_index_nolf(f,   L){ while ((getline L < f) > 0) printf "%s", L; close(f) } \
function putfile_lines(f,        L){ while ((getline L < f) > 0) print L; close(f) } \
BEGIN{ tokB="__WASM_BASE64__"; tokE="/*__WASM_EXEC_JS__*/"; lenB=length(tokB); lenE=length(tokE) } \
{ line=$$0; \
  p=index(line, tokB); \
  if (p) { \
    pre=substr(line,1,p-1); post=substr(line,p+lenB); \
    printf "%s", pre; \
    putfile_index_nolf(b64f); \
    print post; \
    next; \
  } \
  p=index(line, tokE); \
  if (p) { \
    pre=substr(line,1,p-1); post=substr(line,p+lenE); \
    printf "%s", pre; \
    putfile_lines(execf); \
    print post; \
    next; \
  } \
  print; \
}' "$(HTML_IN)" > "$(HTML_OUT)"; \
	echo "Wrote $(HTML_OUT)"

build: $(HTML_OUT)

clean:
	rm -rf $(DIST)

.DEFAULT_GOAL := error

error:
	@echo "Targets:"
	@echo "  make build   # build dist/index.html (self-contained)"
	@echo "  make clean"
	@exit 1
