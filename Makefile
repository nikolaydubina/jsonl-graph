build:
	go build

clean:
	-rm jsonl-graph
	-rm docs/*.svg docs/*.dot

docs: clean build
	cat docs/gin.jsonl | ./jsonl-graph > docs/gin_nocolor.dot
	cat docs/gin_nocolor.dot | dot -Tsvg > docs/gin_nocolor.svg
	cat docs/gin.jsonl | ./jsonl-graph -color-scheme=file://$$PWD/docs/basic-colors.json > docs/gin_color.dot
	cat docs/gin_color.dot | dot -Tsvg > docs/gin_color.svg
	cat docs/small.jsonl | ./jsonl-graph > docs/small.dot
	cat docs/small.dot | dot -Tsvg > docs/small.svg

test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

web: clean
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/
	cd web; GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o main.wasm main.go

serve:
	cd web; python3 -m http.server 8000

.PHONY: docs clean build
