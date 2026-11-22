# vault-markdown

Obsidian-style markdown rendering for Go. Renders vault pages — frontmatter,
callouts, embeds, wikilinks — to HTML so a service can display them without
handing off to Obsidian.

## Features

- GFM tables and task lists (via goldmark)
- YAML frontmatter rendered as a two-column `<table class="frontmatter">`
- Callouts: `> [!note] Title` → `<div class="callout" data-callout="note">`
- Embeds: a line holding only `![[page]]` is inlined one level deep; the
  embedded page's frontmatter is not inlined, and a nested embed is dropped
- Wikilinks: `[[Page]]`, `[[Page|alias]]`, `[[Page#Heading]]` through a
  pluggable resolver; missing targets render `<span class="wikilink missing">`
- `Backlinks(pages, target)` — the pages linking to a target, sorted by path,
  ignoring links inside frontmatter, fenced blocks and inline code

Raw HTML in the source is escaped (goldmark's safe default), so a page that was
not authored by the vault owner cannot inject script into the rendered output.

## Install

```bash
go get github.com/bborbe/vault-markdown
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	markdown "github.com/bborbe/vault-markdown"
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
	out, err := renderer.Render(context.Background(), "# Hello\n\n[[Page]]\n")
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
