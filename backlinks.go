// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"regexp"
	"sort"
	"strings"
)

var (
	fencedCodePattern = regexp.MustCompile("(?s)```.*?```")
	tildeCodePattern  = regexp.MustCompile("(?s)~~~.*?~~~")
	inlineCodePattern = regexp.MustCompile("`[^`\n]*`")
)

// Backlinks returns the names of the pages that link to target, sorted by path.
// Both [[target]] and ![[target]] count, with an optional alias or heading anchor.
// A link that appears inside frontmatter, a fenced block or an inline code span
// is not a backlink.
func Backlinks(pages map[string]string, target string) []string {
	if target == "" {
		return nil
	}
	pattern := regexp.MustCompile(`\[\[` + regexp.QuoteMeta(target) + `(?:[|#][^\]]*)?\]\]`)
	result := make([]string, 0, len(pages))
	for name, content := range pages {
		if strings.TrimSuffix(name, ".md") == target {
			continue
		}
		if pattern.MatchString(linkableText(content)) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

// linkableText drops the frontmatter, fenced blocks and inline code, so a link
// shown as code is not counted.
func linkableText(content string) string {
	_, body := splitFrontmatter(normalizeNewlines(content))
	body = fencedCodePattern.ReplaceAllString(body, "")
	body = tildeCodePattern.ReplaceAllString(body, "")
	return inlineCodePattern.ReplaceAllString(body, "")
}
