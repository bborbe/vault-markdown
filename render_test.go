// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	markdown "github.com/bborbe/vault-markdown"
)

func fixture(name string) string {
	content, err := os.ReadFile(filepath.Join("testdata", "fixtures", name+".md"))
	Expect(err).To(BeNil())
	return string(content)
}

func fixturePages() markdown.Pages {
	return markdown.PagesFunc(func(name string) (string, bool) {
		content, err := os.ReadFile(filepath.Join("testdata", "fixtures", name+".md"))
		if err != nil {
			return "", false
		}
		return string(content), true
	})
}

func fixtureResolver() markdown.Resolver {
	return markdown.ResolverFunc(func(target string) (string, bool) {
		if target == "Page" {
			return "/p/Page", true
		}
		return "", false
	})
}

var _ = Describe("Renderer", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	render := func(source string) string {
		out, err := markdown.New(
			markdown.WithResolver(fixtureResolver()),
			markdown.WithPages(fixturePages()),
		).Render(ctx, source)
		Expect(err).To(BeNil())
		return out
	}

	It("renders GFM tables with their header cell", func() {
		out := render(fixture("page"))
		Expect(out).To(ContainSubstring("<table>"))
		Expect(out).To(ContainSubstring("<th>Name</th>"))
	})

	It("renders task lists with the completed item checked", func() {
		out := render(fixture("page"))
		Expect(out).To(ContainSubstring(`type="checkbox"`))
		Expect(out).To(ContainSubstring("checked"))
		Expect(out).To(ContainSubstring("open item"))
	})

	It("renders frontmatter as a second, distinct table", func() {
		out := render(fixture("page"))
		Expect(strings.Count(out, "<table")).To(Equal(2))
		Expect(out).To(ContainSubstring(`<table class="frontmatter">`))
		Expect(out).To(ContainSubstring("<td>status</td><td>active</td>"))
	})

	It("renders a callout, not a plain blockquote", func() {
		out := render(fixture("page"))
		Expect(out).To(ContainSubstring(`<div class="callout" data-callout="note">`))
		Expect(out).To(ContainSubstring("Callout Title"))
		Expect(out).To(ContainSubstring("Callout body text."))
		Expect(out).NotTo(ContainSubstring("<blockquote>"))
	})

	It("resolves every wikilink form and marks the missing one", func() {
		out := render(fixture("page"))
		Expect(out).To(ContainSubstring(`<a class="wikilink" href="/p/Page">Page</a>`))
		Expect(out).To(ContainSubstring(`<a class="wikilink" href="/p/Page">alias</a>`))
		Expect(out).To(ContainSubstring(`<a class="wikilink" href="/p/Page#Heading">Page</a>`))
		Expect(out).To(ContainSubstring(`<span class="wikilink missing">Missing</span>`))
	})

	It("inlines an embed one level and stops there", func() {
		out := render(fixture("a"))
		Expect(out).To(ContainSubstring("b-body"))
		Expect(out).NotTo(ContainSubstring("c-body"))
		Expect(out).NotTo(ContainSubstring("![[c]]"))
	})

	It("does not inline the embedded page's frontmatter", func() {
		out := render(fixture("embed_frontmatter"))
		Expect(out).To(ContainSubstring("embedded body"))
		Expect(out).NotTo(ContainSubstring(`<table class="frontmatter">`))
		Expect(out).NotTo(ContainSubstring("embedded-status"))
	})

	It("renders a live vault page with its frontmatter, table and wikilinks", func() {
		out := render(fixture("real"))
		Expect(out).To(ContainSubstring(`<table class="frontmatter">`))
		Expect(out).To(ContainSubstring("<table>"))
		Expect(out).To(ContainSubstring(`class="wikilink`))
	})

	It("marks every wikilink missing when no resolver is configured", func() {
		out, err := markdown.New().Render(ctx, "[[Page]]")
		Expect(err).To(BeNil())
		Expect(out).To(ContainSubstring(`<span class="wikilink missing">Page</span>`))
	})

	It("accepts CRLF line endings around the frontmatter", func() {
		out := render("---\r\nstatus: active\r\n---\r\n\r\nbody text\r\n")
		Expect(out).To(ContainSubstring(`<table class="frontmatter">`))
		Expect(out).To(ContainSubstring("<td>status</td><td>active</td>"))
		Expect(out).To(ContainSubstring("body text"))
	})

	It("returns an error for malformed frontmatter instead of panicking", func() {
		_, err := markdown.New().Render(ctx, "---\nstatus: [unclosed\n---\n\nbody\n")
		Expect(err).NotTo(BeNil())
	})

	It("renders a callout without a title", func() {
		out := render("> [!warning]\n> body only\n")
		Expect(out).To(ContainSubstring(`data-callout="warning"`))
		Expect(out).To(ContainSubstring("body only"))
		Expect(out).NotTo(ContainSubstring("callout-title"))
	})

	It("drops Obsidian's fold marker from the callout title", func() {
		out := render("> [!note]- Folded Title\n> body\n")
		Expect(out).To(ContainSubstring(`<p class="callout-title">Folded Title</p>`))
	})

	It("keeps the title when it carries inline markup", func() {
		out := render("> [!note] **Bold** title\n> body\n")
		Expect(out).To(ContainSubstring("callout-title"))
		Expect(out).To(ContainSubstring("Bold"))
		Expect(out).To(ContainSubstring("body"))
	})
})
