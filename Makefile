build:
	go build

clean:
	-rm jsonl-graph
	-rm testdata/*.svg testdata/*.dot

docs: clean build
	cat testdata/gin.jsonl | ./jsonl-graph > testdata/gin_nocolor.dot
	cat testdata/gin_nocolor.dot | dot -Tsvg > testdata/gin_nocolor.svg
	cat testdata/gin.jsonl | ./jsonl-graph -color-scheme=file://$$PWD/testdata/colors.json > testdata/gin_color.dot
	cat testdata/gin_color.dot | dot -Tsvg > testdata/gin_color.svg
	cat testdata/small.jsonl | ./jsonl-graph > testdata/small.dot
	cat testdata/small.dot | dot -Tsvg > testdata/small.svg

test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

web: clean
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/
	cd web; GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o main.wasm main.go

serve:
	cd web; python3 -m http.server 8000

.PHONY: build clean docs test web serve
