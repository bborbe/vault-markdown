// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindWikilink is the node kind for a [[wikilink]].
var KindWikilink = ast.NewNodeKind("Wikilink")

// KindEmbed is the node kind for an inline ![[embed]].
var KindEmbed = ast.NewNodeKind("Embed")

type wikilinkNode struct {
	ast.BaseInline
	Target string
	Alias  string
	Anchor string
	Embed  bool
}

func (n *wikilinkNode) Kind() ast.NodeKind {
	if n.Embed {
		return KindEmbed
	}
	return KindWikilink
}

func (n *wikilinkNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Target": n.Target,
		"Alias":  n.Alias,
		"Anchor": n.Anchor,
		"Embed":  boolString(n.Embed),
	}, nil)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

type wikilinkParser struct{}

func (p *wikilinkParser) Trigger() []byte { return []byte{'!', '['} }

func (p *wikilinkParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, _ := block.PeekLine()
	embed := len(line) > 0 && line[0] == '!'
	open := 0
	if embed {
		open = 1
	}
	if len(line) < open+4 || line[open] != '[' || line[open+1] != '[' {
		return nil
	}
	closing := strings.Index(string(line[open+2:]), "]]")
	if closing < 0 {
		return nil
	}
	body := string(line[open+2 : open+2+closing])
	target, alias, anchor := splitWikilink(body)
	if target == "" {
		return nil
	}
	block.Advance(open + 2 + closing + 2)
	return &wikilinkNode{Target: target, Alias: alias, Anchor: anchor, Embed: embed}
}

// splitWikilink splits a wikilink body into target, alias and heading anchor.
func splitWikilink(body string) (target, alias, anchor string) {
	target = body
	if i := strings.Index(target, "|"); i >= 0 {
		alias = strings.TrimSpace(target[i+1:])
		target = target[:i]
	}
	if i := strings.Index(target, "#"); i >= 0 {
		anchor = strings.TrimSpace(target[i+1:])
		target = target[:i]
	}
	return strings.TrimSpace(target), alias, anchor
}

// renderWikilink writes a resolved anchor, or a missing marker when the
// resolver has no href for the target.
func (c *converter) renderWikilink(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n, ok := node.(*wikilinkNode)
	if !ok {
		return ast.WalkContinue, nil
	}
	label := n.Target
	if n.Alias != "" {
		label = n.Alias
	}
	href, resolved := c.resolve(n)
	if !resolved {
		_, _ = w.WriteString(
			`<span class="wikilink missing">` + html.EscapeString(label) + `</span>`,
		)
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(
		`<a class="wikilink" href="` + html.EscapeString(
			href,
		) + `">` + html.EscapeString(
			label,
		) + `</a>`,
	)
	return ast.WalkContinue, nil
}

func (c *converter) resolve(n *wikilinkNode) (string, bool) {
	if c.resolver == nil {
		return "", false
	}
	href, ok := c.resolver.Resolve(n.Target)
	if !ok || href == "" {
		return "", false
	}
	if n.Anchor != "" {
		return href + "#" + n.Anchor, true
	}
	return href, true
}

// renderEmbed drops an inline ![[embed]]; block-level embeds are expanded
// before parsing, so only the nested or mid-sentence form reaches this point.
func (c *converter) renderEmbed(
	_ util.BufWriter, _ []byte, _ ast.Node, _ bool,
) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
