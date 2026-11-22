# vault-markdown

Obsidian-style markdown rendering for Go. Renders vault pages — frontmatter,
callouts, embeds, wikilinks — to HTML so a service can display them without
handing off to Obsidian.

## Features

- GFM tables and task lists (via goldmark)
- YAML frontmatter rendered as a two-column table
- Callouts: `> [!note] Title` → `<div class="callout" data-callout="note">`
- Embeds: `![[page]]` inlined one level deep
- Wikilinks: `[[Page]]`, `[[Page|alias]]`, `[[Page#Heading]]` through a pluggable resolver; missing targets are marked
- `Backlinks(pages, target)` — the pages linking to a target, sorted by path

## Install

```bash
go get github.com/bborbe/vault-markdown
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/bborbe/vault-markdown"
)

func main() {
	renderer := markdown.New(
		markdown.WithResolver(markdown.ResolverFunc(func(target string) (string, bool) {
			return "/p/" + target, true
		})),
		markdown.WithPages(markdown.PagesFunc(func(name string) (string, bool) {
			return "", false
		})),
	)
	out, err := renderer.Render("# Hello\n\n[[Page]]\n")
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
```

`Backlinks` works on a plain map of page name to content:

```go
links := markdown.Backlinks(map[string]string{
	"zeta.md":  "see [[A]]",
	"alpha.md": "see [[A|the page]]",
}, "A")
// links == []string{"alpha.md", "zeta.md"}
```

## License

BSD-style — see [LICENSE](LICENSE).
