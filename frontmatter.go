// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/bborbe/errors"
	"gopkg.in/yaml.v3"
)

const frontmatterDelimiter = "---"

func normalizeNewlines(source string) string {
	return strings.ReplaceAll(source, "\r\n", "\n")
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
func renderFrontmatter(ctx context.Context, front string) (string, error) {
	if strings.TrimSpace(front) == "" {
		return "", nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(front), &doc); err != nil {
		return "", errors.Wrap(ctx, err, "parse frontmatter")
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return "", nil
	}
	var b strings.Builder
	b.WriteString(`<table class="frontmatter">` + "\n")
	b.WriteString("<thead><tr><th>Key</th><th>Value</th></tr></thead>\n<tbody>\n")
	writeFrontmatterRows(&b, doc.Content[0])
	b.WriteString("</tbody>\n</table>\n")
	return b.String(), nil
}

func writeFrontmatterRows(b *strings.Builder, mapping *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i].Value
		value := frontmatterValue(mapping.Content[i+1])
		fmt.Fprintf(
			b,
			"<tr><td>%s</td><td>%s</td></tr>\n",
			html.EscapeString(key),
			html.EscapeString(value),
		)
	}
}

// frontmatterValue flattens a YAML node to a single cell value.
func frontmatterValue(n *yaml.Node) string {
	switch n.Kind {
	case yaml.SequenceNode:
		parts := make([]string, 0, len(n.Content))
		for _, item := range n.Content {
			parts = append(parts, frontmatterValue(item))
		}
		return strings.Join(parts, ", ")
	case yaml.MappingNode:
		parts := make([]string, 0, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			parts = append(parts, n.Content[i].Value+": "+frontmatterValue(n.Content[i+1]))
		}
		return strings.Join(parts, ", ")
	default:
		return n.Value
	}
}
