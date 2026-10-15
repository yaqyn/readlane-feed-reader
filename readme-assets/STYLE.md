# Readlane artwork and palette

`DESIGN.md` owns the palette. `scripts/design-tokens.py` generates terminal
constants, web CSS tokens, and the palette of the canonical README SVG icons.

- Primary violet: `#6253D9`; hover: `#5042BD`.
- Porcelain background: `#F5F6FB`; reading surface: `#FFFFFF`.
- Ink: `#20253B`; supporting text: `#626B80`; selection: `#EAE7FA`.
- Terminal dark variants retain readable violet, ink, and supporting text.

The 2172×724 README cover uses a large monospace wordmark, violet terminal
prompt, and block cursor on a dark background,
with the exact name **Readlane** and tagline **Your feeds. Your pace.**
It was generated using the built-in image tool. See the [brand guide](../docs/brand.md)
for its prompt and production record. The WebP is lossless; no resize or crop
was applied. Its only descriptive metadata is:

- `XMP-dc:Creator`: `Abdulrahman M. Yaqyn`
- `XMP-dc:Rights`: `Copyright Abdulrahman M. Yaqyn`

Fonts are self-hosted in `internal/web/assets/fonts`; their original OFL license
notices remain alongside the files. SVG icon geometry remains the workspace's
canonical 24-unit family, with 1.6-unit rounded strokes and palette recoloring.
