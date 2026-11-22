// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindCallout is the node kind for a > [!type] callout.
var KindCallout = ast.NewNodeKind("Callout")

type calloutNode struct {
	ast.BaseBlock
	CalloutType string
	Title       string
}

func (n *calloutNode) Kind() ast.NodeKind { return KindCallout }

func (n *calloutNode) Dump(source []byte, level int) {
	ast.DumpHelper(
		n,
		source,
		level,
		map[string]string{"CalloutType": n.CalloutType, "Title": n.Title},
		nil,
	)
}

type calloutTransformer struct{}

// Transform converts every blockquote whose first line is a [!type] marker into
// a callout node carrying the type and the title.
func (t *calloutTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	var found []*ast.Blockquote
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if bq, ok := n.(*ast.Blockquote); ok && calloutMarker(bq, source) != "" {
			found = append(found, bq)
		}
		return ast.WalkContinue, nil
	})
	for _, bq := range found {
		replaceWithCallout(bq, source)
	}
}

// rawFirstLine returns the source text of a block's first line, markup included.
func rawFirstLine(block ast.Node, source []byte) (string, int) {
	lines := block.Lines()
	if lines.Len() == 0 {
		return "", -1
	}
	segment := lines.At(0)
	return string(source[segment.Start:segment.Stop]), segment.Stop
}

// calloutMarker returns the callout type when the blockquote starts with
// [!type], and an empty string otherwise.
func calloutMarker(bq *ast.Blockquote, source []byte) string {
	paragraph := bq.FirstChild()
	if paragraph == nil || paragraph.Kind() != ast.KindParagraph {
		return ""
	}
	line, _ := rawFirstLine(paragraph, source)
	if !strings.HasPrefix(line, "[!") {
		return ""
	}
	end := strings.Index(line, "]")
	if end < 0 {
		return ""
	}
	return line[2:end]
}

func replaceWithCallout(bq *ast.Blockquote, source []byte) {
	marker := calloutMarker(bq, source)
	line, lineEnd := rawFirstLine(bq.FirstChild(), source)
	callout := &calloutNode{CalloutType: marker, Title: calloutTitle(line, marker)}
	bq.Parent().ReplaceChild(bq.Parent(), bq, callout)
	moveChildren(bq, callout)
	stripMarker(callout, lineEnd)
}

// calloutTitle returns the text after [!type], without Obsidian's fold marker.
func calloutTitle(line, marker string) string {
	index := strings.Index(line, "[!"+marker+"]")
	if index < 0 {
		return ""
	}
	title := strings.TrimSpace(line[index+len(marker)+3:])
	return strings.TrimLeft(title, "-+ ")
}

func moveChildren(from, to ast.Node) {
	for child := from.FirstChild(); child != nil; {
		next := child.NextSibling()
		to.AppendChild(to, child)
		child = next
	}
}

// nodeStart returns the source offset where a node's text begins.
func nodeStart(n ast.Node) int {
	if textNode, ok := n.(*ast.Text); ok {
		return textNode.Segment.Start
	}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if start := nodeStart(child); start >= 0 {
			return start
		}
	}
	return -1
}

// stripMarker removes the [!type] title line so it is not repeated in the body.
func stripMarker(callout *calloutNode, firstLineEnd int) {
	paragraph := callout.FirstChild()
	if paragraph == nil || firstLineEnd < 0 {
		return
	}
	for child := paragraph.FirstChild(); child != nil; {
		next := child.NextSibling()
		if start := nodeStart(child); start >= 0 && start < firstLineEnd {
			paragraph.RemoveChild(paragraph, child)
		}
		child = next
	}
	if paragraph.ChildCount() == 0 {
		callout.RemoveChild(callout, paragraph)
	}
}

// renderCallout writes the callout container, its title and its body.
func (c *converter) renderCallout(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	n, ok := node.(*calloutNode)
	if !ok {
		return ast.WalkContinue, nil
	}
	if !entering {
		_, _ = w.WriteString("</div>\n")
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(
		`<div class="callout" data-callout="` + html.EscapeString(n.CalloutType) + `">` + "\n",
	)
	if n.Title != "" {
		_, _ = fmt.Fprintf(w, "<p class=\"callout-title\">%s</p>\n", html.EscapeString(n.Title))
	}
	return ast.WalkContinue, nil
}
