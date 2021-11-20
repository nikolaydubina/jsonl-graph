# JSONL Graph Tools

> Convenient to use with `jq`

[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/jsonl-graph.svg)](https://pkg.go.dev/github.com/nikolaydubina/jsonl-graph)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/jsonl-graph)](https://goreportcard.com/report/github.com/nikolaydubina/jsonl-graph)
[![codecov](https://codecov.io/gh/nikolaydubina/jsonl-graph/branch/main/graph/badge.svg?token=gU3DUNXgX3)](https://codecov.io/gh/nikolaydubina/jsonl-graph)
[![Tests](https://github.com/nikolaydubina/jsonl-graph/workflows/Tests/badge.svg)](https://github.com/nikolaydubina/jsonl-graph/actions)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#science-and-data-analysis)

```
# get https://graphviz.org/download/ 
$ go install github.com/nikolaydubina/jsonl-graph@latest
```

What is JSONL graph? Node has `id`. Edge has `from` and `to`.
```
{
    "id": "github.com/gin-gonic/gin",
    "can_get_github": true,
    "github_url": "https://github.com/gin-gonic/gin",
    "git_last_commit": "2021-04-21",
    "git_num_contributors": 321,
    ...
}
...
{
    "from": "github.com/gin-gonic/gin",
    "to": "golang.org/x/tools",
    ...
}
```

## Examples

Kubernetes Pod Owners

```bash
$ kubectl get pods -o json | jq '.items[] | {to: (.kind + ":" + .metadata.name), from: (.metadata.ownerReferences[].kind + ":" + .metadata.ownerReferences[].name)}' | ./jsonl-graph | dot -Tsvg > k8s_pod_owners.svg
```

![k8s_pod_owners](./testdata/k8s_pod_owners.svg)

Large nodes and color scheme
```bash
$ cat '
{"id":"github.com/gin-gonic/gin","can_get_git":true, ... }
{"id":"github.com/gin-contrib/sse","can_get_git":true,"can_run_tests":true ... }
...
{"from":"github.com/gin-gonic/gin","to":"golang.org/x/tools"}
{"from":"github.com/gin-gonic/gin","to":"github.com/go-playground/validator/v10"}
' | jsonl-graph -color-scheme=file://$PWD/testdata/colors.json | dot -Tsvg > colored.svg
```
![gin-color](./testdata/gin_color.svg)

Small nodes or only edges
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

![small](./testdata/small.svg)

## Rendering

Currently only Graphviz is supported.
Follow progress of native Go graph rendering in [github.com/nikolaydubina/go-graph-layout](https://github.com/nikolaydubina/go-graph-layout). Once it is ready, it will be integrated into this project.