# JSONL Graph Tools

> Convenient to use with `jq`

[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/jsonl-graph.svg)](https://pkg.go.dev/github.com/nikolaydubina/jsonl-graph)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/jsonl-graph)](https://goreportcard.com/report/github.com/nikolaydubina/jsonl-graph)
[![codecov](https://codecov.io/gh/nikolaydubina/jsonl-graph/branch/main/graph/badge.svg?token=gU3DUNXgX3)](https://codecov.io/gh/nikolaydubina/jsonl-graph)
[![Tests](https://github.com/nikolaydubina/jsonl-graph/workflows/Tests/badge.svg)](https://github.com/nikolaydubina/jsonl-graph/actions)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#science-and-data-analysis)

```bash
$ go install github.com/nikolaydubina/jsonl-graph@latest
```

.. or visit TODO: web URL

Self-contained SVG graph renderer.
This can be an alternative to Graphviz.
However, it supports Graphviz as an output target too.

## [JSONL](https://jsonlines.org/) Graph?

Graph is represented as JSONL of nodes and edges.

Node has `id` and any fields:
```json
{
    "id": "github.com/gin-gonic/gin",
    "can_get_github": true,
    "github_url": "https://github.com/gin-gonic/gin",
    "git_last_commit": "2021-04-21",
    "git_num_contributors": 321,
    ...
}
```

Edge has `from` and `to` of node `id`:
```json
{
    "from": "github.com/gin-gonic/gin",
    "to": "golang.org/x/tools",
    ...
}
```

This data encoding is a perfect fit for storing graphs:
- can append new nodes and edges by concatenating files
- nodes and edges can have any data
- schema-less
- any subset of (compressed) lines is a valid graph

## Features

- [x] Target: SVG
- [x] Target: Graphviz
- [ ] Coloring of fields with filters
- [x] Layout: Gravity Force
- [x] Layout: Spring Force
- [x] Layout: Isomap (gonum)
- [x] Layout: Eades (gonum)
- [ ] Layout: Magnetic Force (Kozo Sugiyama)
- [ ] Layout: Layers (Kozo Sugiyama)
- [ ] Metro-style edges
- [x] Interactive Web UI with WebAssembly
- [ ] Touch for zoom and pan in Web UI
- [ ] CLI
- [x] 100% Go
- [x] 100% offline
- [x] Self contained (..almost)

## Example Web UI

TODO: link here
TODO: screenshot here

## Example CLI

To illustrate, I am using data from [import-graph](github.com/nikolaydubina/import-graph). 

[Graphviz](https://graphviz.org/download/) renderer also supported. If you pass color scheme, then values will be colored.
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

## TODO Wishlist

- [ ] Add centerer that puts center of mass + 95% of density at center of screen
- [ ] UI collapsible sections
- [ ] Fetch JSON from URL + jsonpath for field
- [ ] Fetch JSONL from URL
- [ ] Update graph without re-rendering whole

## Missing Features? Want to add something?

Contributions welcomed!

## References

- [Graph Drawing](https://en.wikipedia.org/wiki/Graph_drawing)
- [google/pprof svg](https://github.com/google/pprof/blob/master/internal/driver/svg.go)
- [google/pprof svgpan](https://github.com/google/pprof/blob/master/third_party/svgpan/svgpan.go)
- [Graph Compact Orthogonal Layout Algorithm](https://arxiv.org/pdf/1807.09368.pdf)
- [Coffman-Graham algorithm](https://en.wikipedia.org/wiki/Coffman–Graham_algorithm)
- [gonum isomap layout](https://github.com/gonum/gonum/blob/master/graph/layout/isomap.go)
- [gonum eades layout](https://github.com/gonum/gonum/blob/master/graph/layout/eades.go)
- [Booststrap](https://getbootstrap.com)
- [Go WebAssembly](https://github.com/golang/go/wiki/WebAssembly#debugging)
- [Layered Graph](https://en.wikipedia.org/wiki/Layered_graph_drawing)
- Fruchterman & Reingold forces
- Lombarding-Spring-Embedder (Chernobleskiy, 2012)
- Realistic Node Sizes (Gasner, North 1998)
- [Minsky's Theorem](https://en.wikipedia.org/wiki/Mirsky%27s_theorem)
- Magnetic Spring graph drawing, Kozo Sugiyama