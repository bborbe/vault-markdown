// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"bytes"
	"context"
	"strings"

	"github.com/bborbe/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	grenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Resolver resolves a wikilink target to an href.
type Resolver interface {
	// Resolve returns the href for target and whether target exists.
	Resolve(target string) (string, bool)
}

// ResolverFunc adapts a function to Resolver.
type ResolverFunc func(target string) (string, bool)

// Resolve implements Resolver.
func (f ResolverFunc) Resolve(target string) (string, bool) { return f(target) }

// Pages returns page content by name so ![[embeds]] can be inlined.
type Pages interface {
	// Page returns the content of name and whether it exists.
	Page(name string) (string, bool)
}

// PagesFunc adapts a function to Pages.
type PagesFunc func(name string) (string, bool)

// Page implements Pages.
func (f PagesFunc) Page(name string) (string, bool) { return f(name) }

// Renderer renders Obsidian-flavoured markdown to HTML.
type Renderer interface {
	// Render renders source to HTML.
	Render(ctx context.Context, source string) (string, error)
}

type renderer struct {
	resolver Resolver
	pages    Pages
}

// Option configures a Renderer.
type Option func(*renderer)

// WithResolver sets the wikilink resolver. Without one every wikilink is missing.
func WithResolver(r Resolver) Option {
	return func(rd *renderer) { rd.resolver = r }
}

// WithPages sets the source used to expand ![[embeds]]. Without one an embed is dropped.
func WithPages(p Pages) Option {
	return func(rd *renderer) { rd.pages = p }
}

// New returns a Renderer configured with opts.
func New(opts ...Option) Renderer {
	r := &renderer{}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *renderer) Render(ctx context.Context, source string) (string, error) {
	c := &converter{ctx: ctx, resolver: r.resolver, pages: r.pages}
	return c.render(source)
}

// converter carries the state of one render pass; depth counts embed nesting.
type converter struct {
	ctx      context.Context
	resolver Resolver
	pages    Pages
	depth    int
}

func (c *converter) render(source string) (string, error) {
	source = normalizeNewlines(source)
	front, body := splitFrontmatter(source)
	table, err := renderFrontmatter(c.ctx, front)
	if err != nil {
		return "", err
	}
	converted, err := c.convert(c.expandEmbeds(body))
	if err != nil {
		return "", err
	}
	return table + converted, nil
}

// expandEmbeds replaces a line holding only ![[page]] with that page's body.
// Expansion stops after one level: an embed inside the inlined body is dropped.
func (c *converter) expandEmbeds(body string) string {
	if c.pages == nil {
		return body
	}
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		target, ok := loneEmbed(line)
		if !ok {
			continue
		}
		if c.depth > 0 {
			lines[i] = ""
			continue
		}
		content, found := c.pages.Page(target)
		if !found {
			continue
		}
		child := &converter{ctx: c.ctx, resolver: c.resolver, pages: c.pages, depth: c.depth + 1}
		_, inner := splitFrontmatter(normalizeNewlines(content))
		lines[i] = child.expandEmbeds(inner)
	}
	return strings.Join(lines, "\n")
}

// loneEmbed reports whether line is nothing but an ![[target]] embed.
func loneEmbed(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "![[") || !strings.HasSuffix(trimmed, "]]") {
		return "", false
	}
	target := strings.TrimSpace(trimmed[3 : len(trimmed)-2])
	if target == "" || strings.ContainsAny(target, "[]") {
		return "", false
	}
	return target, true
}

func (c *converter) convert(source string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithInlineParsers(util.Prioritized(&wikilinkParser{}, 100)),
			parser.WithASTTransformers(util.Prioritized(&calloutTransformer{}, 100)),
		),
		goldmark.WithRendererOptions(grenderer.WithNodeRenderers(util.Prioritized(c, 100))),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", errors.Wrap(c.ctx, err, "convert markdown")
	}
	return buf.String(), nil
}

// RegisterFuncs implements renderer.NodeRenderer.
func (c *converter) RegisterFuncs(reg grenderer.NodeRendererFuncRegisterer) {
	reg.Register(KindWikilink, c.renderWikilink)
	reg.Register(KindEmbed, c.renderEmbed)
	reg.Register(KindCallout, c.renderCallout)
}
