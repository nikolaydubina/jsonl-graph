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

.PHONY: docs clean build