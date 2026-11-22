// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"regexp"
	"sort"
	"strings"
)

// Backlinks returns the names of the pages that link to target, sorted by path.
// Both [[target]] and ![[target]] count, with an optional alias or heading anchor.
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
		if pattern.MatchString(content) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
