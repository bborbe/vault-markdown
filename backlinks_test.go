// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	markdown "github.com/bborbe/vault-markdown"
)

var _ = Describe("Backlinks", func() {
	It("returns the linking pages sorted by path, ignoring the target itself", func() {
		links := markdown.Backlinks(map[string]string{
			"zeta.md":  fixture("zeta"),
			"alpha.md": fixture("alpha"),
			"other.md": "no links here",
			"A.md":     "self link [[A]]",
		}, "A")
		Expect(links).To(Equal([]string{"alpha.md", "zeta.md"}))
	})

	It("returns nothing when no page links to the target", func() {
		Expect(markdown.Backlinks(map[string]string{"one.md": "plain"}, "A")).To(BeEmpty())
	})

	It("returns nothing for an empty target", func() {
		Expect(markdown.Backlinks(map[string]string{"one.md": "[[A]]"}, "")).To(BeEmpty())
	})
})
