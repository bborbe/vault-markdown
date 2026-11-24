# Changelog

All notable changes to this project will be documented in this file.

## v0.1.1

- fix: render `[[wikilinks]]` in frontmatter values as anchors (quoted scalar, quoted list, and inline unquoted forms) instead of escaped literal text, reusing the configured resolver; unresolved targets render the same `wikilink missing` marker the body uses.

## v0.1.0

- feat: initial release — a goldmark-based renderer for Obsidian-flavoured markdown. Frontmatter renders as a table, `> [!note]` callouts render as callout elements, `![[embeds]]` inline one level deep, `[[Page]]` / `[[Page|alias]]` / `[[Page#Heading]]` resolve through a pluggable resolver with missing targets marked, and `Backlinks` returns the linking pages sorted by path.
