# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

- feat: initial release — a goldmark-based renderer for Obsidian-flavoured markdown. Frontmatter renders as a table, `> [!note]` callouts render as callout elements, `![[embeds]]` inline one level deep, `[[Page]]` / `[[Page|alias]]` / `[[Page#Heading]]` resolve through a pluggable resolver with missing targets marked, and `Backlinks` returns the linking pages sorted by path.
