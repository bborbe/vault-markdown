# Changelog

All notable changes to this project will be documented in this file.

## v0.1.3

- chore: update github.com/bborbe/errors to v1.7.1

## v0.1.2

- fix: quote a frontmatter value made only of bare `[[wikilinks]]` — `tags: [[Go]] [[Trading]]`, or a `- [[A]] [[B]]` list item — before parsing, so the page renders instead of failing the whole frontmatter with a YAML parse error. Obsidian accepts that unquoted shape, YAML reads the second `[[…]]` on the line as a syntax error, and the page view answered 500 for every page carrying it.

## v0.1.1

- fix: render `[[wikilinks]]` in frontmatter values as anchors (quoted scalar, quoted list, and inline unquoted forms) instead of escaped literal text, reusing the configured resolver; unresolved targets render the same `wikilink missing` marker the body uses.

## v0.1.0

- feat: initial release — a goldmark-based renderer for Obsidian-flavoured markdown. Frontmatter renders as a table, `> [!note]` callouts render as callout elements, `![[embeds]]` inline one level deep, `[[Page]]` / `[[Page|alias]]` / `[[Page#Heading]]` resolve through a pluggable resolver with missing targets marked, and `Backlinks` returns the linking pages sorted by path.
