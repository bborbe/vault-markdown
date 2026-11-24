// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/bborbe/errors"
	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

// Obsidian accepts an unquoted wikilink value, and a whole list of them on one
// line: `tags: [[Go]] [[Trading]]`. YAML reads the first `[[Go]]` as a flow
// sequence and then rejects the second one on the same line, so before this
// tolerance the whole page failed to render. The two patterns match exactly
// that shape — a value made only of `[[…]]` groups separated by spaces, on a
// mapping line or on a block-list item — and quote it, which parses.
var (
	bareWikilinkMapping = regexp.MustCompile(
		`(?m)^([ \t]*[A-Za-z0-9_.-]+:[ \t]*)` +
			`(\[\[[^\]\n]*\]\](?:[ \t]+\[\[[^\]\n]*\]\])*)[ \t]*$`,
	)
	bareWikilinkListItem = regexp.MustCompile(
		`(?m)^([ \t]*-[ \t]+)(\[\[[^\]\n]*\]\](?:[ \t]+\[\[[^\]\n]*\]\])*)[ \t]*$`,
	)
)

func normalizeNewlines(source string) string {
	return strings.ReplaceAll(source, "\r\n", "\n")
}

// quoteBareWikilinks quotes each value that is nothing but bare `[[…]]` groups,
// so Obsidian's unquoted wikilink shape parses as YAML instead of failing the
// page. Any other line is left byte-identical, so genuinely malformed YAML
// still reports its parse error.
func quoteBareWikilinks(front string) string {
	front = quoteBareWikilinkMatches(front, bareWikilinkMapping)
	return quoteBareWikilinkMatches(front, bareWikilinkListItem)
}

func quoteBareWikilinkMatches(front string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(front, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		return parts[1] + `"` + strings.ReplaceAll(parts[2], `"`, `\"`) + `"`
	})
}

// splitFrontmatter separates a leading YAML frontmatter block from the body.
// A source without frontmatter returns an empty front and the source unchanged.
func splitFrontmatter(source string) (string, string) {
	if !strings.HasPrefix(source, frontmatterDelimiter+"\n") {
		return "", source
	}
	rest := source[len(frontmatterDelimiter)+1:]
	if strings.HasPrefix(rest, frontmatterDelimiter) {
		return "", strings.TrimPrefix(rest[len(frontmatterDelimiter):], "\n")
	}
	end := strings.Index(rest, "\n"+frontmatterDelimiter)
	if end < 0 {
		return "", source
	}
	front := rest[:end]
	body := strings.TrimPrefix(rest[end+len(frontmatterDelimiter)+1:], "\n")
	return front, body
}

// renderFrontmatter renders a YAML frontmatter block as a two-column table.
// resolver, when set, turns a wikilink value into a link the same way the body
// renderer does; without one every wikilink value is marked missing.
func renderFrontmatter(ctx context.Context, front string, resolver Resolver) (string, error) {
	if strings.TrimSpace(front) == "" {
		return "", nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(quoteBareWikilinks(front)), &doc); err != nil {
		return "", errors.Wrap(ctx, err, "parse frontmatter")
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return "", nil
	}
	var b strings.Builder
	b.WriteString(`<table class="frontmatter">` + "\n")
	b.WriteString("<thead><tr><th>Key</th><th>Value</th></tr></thead>\n<tbody>\n")
	writeFrontmatterRows(&b, doc.Content[0], resolver)
	b.WriteString("</tbody>\n</table>\n")
	return b.String(), nil
}

// writeFrontmatterRows writes one row per mapping entry. The key is escaped;
// the value is already safe HTML from frontmatterValue (text escaped, links
// emitted as markup), so it is written as-is.
func writeFrontmatterRows(b *strings.Builder, mapping *yaml.Node, resolver Resolver) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i].Value
		value := frontmatterValue(mapping.Content[i+1], resolver)
		fmt.Fprintf(
			b,
			"<tr><td>%s</td><td>%s</td></tr>\n",
			html.EscapeString(key),
			value,
		)
	}
}

// frontmatterValue flattens a YAML node to a single cell value as safe HTML.
// A wikilink value renders as an anchor; everything else is escaped text.
func frontmatterValue(n *yaml.Node, resolver Resolver) string {
	switch n.Kind {
	case yaml.SequenceNode:
		// Inline unquoted `key: [[Page]]` is valid YAML flow syntax and parses
		// as a sequence holding one sequence holding one scalar, so the
		// brackets are already gone. Reconstruct the wikilink.
		if len(n.Content) == 1 && isSingleScalarSequence(n.Content[0]) {
			return renderFrontmatterScalar("[["+n.Content[0].Content[0].Value+"]]", resolver)
		}
		parts := make([]string, 0, len(n.Content))
		for _, item := range n.Content {
			parts = append(parts, frontmatterValue(item, resolver))
		}
		return strings.Join(parts, ", ")
	case yaml.MappingNode:
		parts := make([]string, 0, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			parts = append(
				parts,
				html.EscapeString(
					n.Content[i].Value,
				)+": "+frontmatterValue(
					n.Content[i+1],
					resolver,
				),
			)
		}
		return strings.Join(parts, ", ")
	default:
		return renderFrontmatterScalar(n.Value, resolver)
	}
}

// isSingleScalarSequence reports whether n is a sequence holding exactly one
// scalar — the shape an inline `[[Page]]` value parses to.
func isSingleScalarSequence(n *yaml.Node) bool {
	return n.Kind == yaml.SequenceNode && len(n.Content) == 1 &&
		n.Content[0].Kind == yaml.ScalarNode
}

// renderFrontmatterScalar renders a scalar cell: each [[wikilink]] becomes an
// anchor (or a missing marker) and the surrounding text is escaped.
func renderFrontmatterScalar(value string, resolver Resolver) string {
	var b strings.Builder
	rest := value
	for {
		start := strings.Index(rest, "[[")
		if start < 0 {
			b.WriteString(html.EscapeString(rest))
			return b.String()
		}
		closing := strings.Index(rest[start+2:], "]]")
		if closing < 0 {
			b.WriteString(html.EscapeString(rest))
			return b.String()
		}
		b.WriteString(html.EscapeString(rest[:start]))
		body := rest[start+2 : start+2+closing]
		target, alias, anchor := splitWikilink(body)
		if target == "" {
			b.WriteString(html.EscapeString(rest[start : start+2+closing+2]))
		} else {
			b.WriteString(renderFrontmatterLink(target, alias, anchor, resolver))
		}
		rest = rest[start+2+closing+2:]
	}
}

// renderFrontmatterLink writes the same anchor/missing marker the body
// renderer emits, so a frontmatter link and a body link never drift.
func renderFrontmatterLink(target, alias, anchor string, resolver Resolver) string {
	label := target
	if alias != "" {
		label = alias
	}
	href, resolved := resolveFrontmatterHref(target, anchor, resolver)
	if !resolved {
		return `<span class="wikilink missing">` + html.EscapeString(label) + `</span>`
	}
	return `<a class="wikilink" href="` + html.EscapeString(
		href,
	) + `">` + html.EscapeString(label) + `</a>`
}

// resolveFrontmatterHref mirrors converter.resolve for a raw cell value: the
// resolver owns the href, and a heading anchor is appended when present.
func resolveFrontmatterHref(target, anchor string, resolver Resolver) (string, bool) {
	if resolver == nil {
		return "", false
	}
	href, ok := resolver.Resolve(target)
	if !ok || href == "" {
		return "", false
	}
	if anchor != "" {
		return href + "#" + anchor, true
	}
	return href, true
}
