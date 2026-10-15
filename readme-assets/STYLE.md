# Readlane artwork and palette

`DESIGN.md` owns the palette. `scripts/design-tokens.py` generates terminal
constants, web CSS tokens, and the palette of the canonical README SVG icons.

- Primary violet: `#6253D9`; hover: `#5042BD`.
- Porcelain background: `#F5F6FB`; reading surface: `#FFFFFF`.
- Ink: `#20253B`; supporting text: `#626B80`; selection: `#EAE7FA`.
- Terminal canvas: `#0C0F18`, with lavender `#AEA5FF` accents, pale ink
  `#F0EFFA`, and violet `#302B4D` selection. The full-screen reader owns its
  background so the cover identity survives different terminal themes.

The 2172×724 README cover depicts pixel-art article towers and blog posts
joining a violet reading lane. The unified **$ Readlane** wordmark
and “RSS & Atom feeds. Read in your terminal.” line state its purpose.
It was generated using the built-in image tool. See the [brand guide](../docs/brand.md)
for its prompt and production record. The WebP is lossless; no resize or crop
was applied. Its only descriptive metadata is:

- `XMP-dc:Creator`: `Abdulrahman M. Yaqyn`
- `XMP-dc:Rights`: `Copyright Abdulrahman M. Yaqyn`

Fonts are self-hosted in `internal/web/assets/fonts`; their original OFL license
notices remain alongside the files. SVG icon geometry remains the workspace's
canonical 24-unit family, with 1.6-unit rounded strokes and palette recoloring.
