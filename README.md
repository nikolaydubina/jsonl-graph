# JSONL graph tools

> Convenient to use with `jq`

[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/jsonl-graph.svg)](https://pkg.go.dev/github.com/nikolaydubina/jsonl-graph)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/jsonl-graph)](https://goreportcard.com/report/github.com/nikolaydubina/jsonl-graph)
[![codecov](https://codecov.io/gh/nikolaydubina/jsonl-graph/branch/main/graph/badge.svg?token=gU3DUNXgX3)](https://codecov.io/gh/nikolaydubina/jsonl-graph)

```bash
$ go install github.com/nikolaydubina/jsonl-graph@latest
# get https://graphviz.org/download/
```

Graph is represented as JSONL of nodes and edges.

Node contains `id` and any fields
```
{
    "id": "github.com/gin-gonic/gin",
    "can_get_git": true,
    "can_run_tests": true,
    "can_get_github": true,
    "github_url": "https://github.com/gin-gonic/gin",
    "git_url": "https://github.com/gin-gonic/gin",
    "git_last_commit": "2021-04-21",
    "git_last_commit_days_since": 4,
    "git_num_contributors": 321,
    ...
}
```

Edge contains `from` and `to` with node `id`s
```json
{
    "from": "github.com/gin-gonic/gin",
    "to": "golang.org/x/tools"
}
```

## Why?

[JSONL](https://jsonlines.org/) is a perfect fit for storing graphs

- can append new nodes and endges by concatenating files
- nodes and edges can have any data
- schemaless
- any subset of lines is a valid graph

## Examples

To illustrate, I am using data from [import-graph](github.com/nikolaydubina/import-graph). If you pass color scheme, then values will be colored.
```bash
$ cat '
{"id":"github.com/gin-gonic/gin","can_get_git":true, ... }
{"id":"github.com/gin-contrib/sse","can_get_git":true,"can_run_tests":true ... }
...
{"from":"github.com/gin-gonic/gin","to":"golang.org/x/tools"}
{"from":"github.com/gin-gonic/gin","to":"github.com/go-playground/validator/v10"}
' | jsonl-graph -color-scheme=file://$PWD/docs/colors.json | dot -Tsvg > colored.svg
```
![gin-color](./docs/gin_color.svg)

By default, no coloring is applied.
```bash
$ cat '
{"id":"github.com/gin-gonic/gin","can_get_git":true, ... }
{"id":"github.com/gin-contrib/sse","can_get_git":true,"can_run_tests":true ... }
...
{"from":"github.com/gin-gonic/gin","to":"golang.org/x/tools"}
{"from":"github.com/gin-gonic/gin","to":"github.com/go-playground/validator/v10"}
' | jsonl-graph | dot -Tsvg > basic.svg
```
![gin-nocolor](./docs/gin_nocolor.svg)

If nodes have less fields, then adjusting graph

```bash
$ cat '
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"bufio"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"bytes"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"encoding/json"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"errors"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"fmt"}
...
' | jsonl-graph | dot -Tsvg > small.svg
```

![small](./docs/small.svg)

## TODO

- [ ] adjust node spacing when graph is large
- [ ] group subgraphs
- [ ] interactive web interface: collapse; highlight; arragement aglrorith; edge shape algorithm;
