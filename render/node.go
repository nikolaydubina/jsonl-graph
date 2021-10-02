package render

import (
	"fmt"
	"image"
	"sort"
	"strings"
)

const nodeFontSize int = 9

// Node is rendered point.
// Can render contents as table.
type Node struct {
	ID         string      // used to make DOM IDs
	LeftBottom image.Point // lowest X and Y coordinate of node box
	ShowData   bool        // if true then render contents of node besides title
	Title      string
	NodeData   map[string]interface{}
}

func (n Node) TitleID() string {
	return fmt.Sprintf("svg:graph:node:title:%s", n.ID)
}

// Reference: https://developer.mozilla.org/en-US/docs/Web/SVG/Element/foreignObject
func (n Node) Render() string {
	padding := 10
	body := ""
	if n.ShowData {
		body = NodeDataTable{NodeData: n.NodeData, FontSize: nodeFontSize}.Render()
	}
	return fmt.Sprintf(`
		<g>
			<foreignObject x="%d" y="%d" width="%d" height="%d">
				<div xmlns="http://www.w3.org/1999/xhtml" class="unselectable" style="overflow: hidden; background: white; border: 1px solid lightgray; border-radius: 5px;">
					%s
					%s
				</div>
			</foreignObject>
		</g>
		`,
		n.LeftBottom.X,
		n.LeftBottom.Y,
		n.Width()+padding,
		n.Height()+padding,
		NodeTitle{ID: fmt.Sprintf("svg:graph:node:title:%s", n.ID), Title: n.Title, FontSize: nodeFontSize}.Render(),
		body,
	)
}

func (n Node) Width() int {
	w := int(float64(nodeFontSize*len(n.Title)) * 0.8)
	if !n.ShowData {
		return w
	}

	nd := NodeDataTable{NodeData: n.NodeData, FontSize: nodeFontSize}
	if nd.Width() > w {
		w = nd.Width()
	}
	return w
}

func (n Node) Height() int {
	titleHeight := 2 * nodeFontSize
	if !n.ShowData {
		return titleHeight
	}

	nd := NodeDataTable{NodeData: n.NodeData, FontSize: nodeFontSize}
	return titleHeight + nd.Height()
}

type NodeTitle struct {
	ID       string
	Title    string
	FontSize int
}

func (n NodeTitle) Render() string {
	return fmt.Sprintf(`
		<div id="%s", style="font-size: %dpx; text-align: center; padding: 4px; cursor: pointer;">
			%s
		</div>`,
		n.ID,
		n.FontSize,
		n.Title,
	)
}

// NodeDataTable renders key-value data of node.
// It will render table.
type NodeDataTable struct {
	NodeData map[string]interface{}
	FontSize int
}

func (n NodeDataTable) Width() int {
	maxlen := 0
	for k, v := range n.NodeData {
		if k == "id" || strings.HasSuffix(k, "_url") {
			continue
		}
		currLen := len(k) + len(RenderValue(v))
		if currLen > maxlen {
			maxlen = currLen
		}
	}
	return int(float64(nodeFontSize*maxlen) * 0.8)
}

func (n NodeDataTable) Height() int {
	nrows := 0
	for k := range n.NodeData {
		if k == "id" || strings.HasSuffix(k, "_url") {
			continue
		}
		nrows++
	}
	return nodeFontSize * nrows * 2
}

func (n NodeDataTable) Render() string {
	rows := []string{}

	for k, v := range n.NodeData {
		if k == "id" || strings.HasSuffix(k, "_url") {
			continue
		}

		row := fmt.Sprintf(`
			<tr>
				<td border="1" align="left">%s</td>
				<td border="1" align="right">%s</td>
			</tr>`,
			k,
			RenderValue(v),
		)

		rows = append(rows, row)
	}

	// sort by key, since key is first
	sort.Strings(rows)

	return fmt.Sprintf(
		`<div style="font-size: %dpx; padding: 0px 4px 4px 4px; border-top: 1px solid lightgrey;">
			<table border="0" cellspacing="0" cellpadding="1" style="width: 100%%;">
			%s
			</table>
		</div>
		`,
		n.FontSize,
		strings.Join(rows, "\n"),
	)
}
