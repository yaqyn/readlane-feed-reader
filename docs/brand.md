# Readlane brand guide

**Readlane** is a terminal-first RSS and Atom reader. Its tagline is
**Your feeds. Your pace.** The name combines the reading task with a dedicated
lane through the publications a reader chooses. It is a naming decision,
not a claim of trademark clearance or domain ownership.

## Identity

The terminal is the main product: a clear wordmark, violet selection, a quiet
reading list, and discoverable keyboard controls. Ordinary commands remain
useful in scripts. The web reader is a secondary surface using the same tone
and palette. Marketing should show those working behaviors and avoid claims
about unread tracking, full-article extraction, hosted accounts, or readership
numbers that the application does not support.

The cover imagines a reading world inside a terminal: article-shaped towers,
bookmark silhouettes, and floating blog posts joining one violet lane. The
pixel-art RSS sign connects the world to feeds. A large Readlane wordmark,
a dollar-prefixed readlane command with cursor, and the purpose line
“RSS & Atom feeds. Read in your terminal.” explain the product without
reproducing a screen.
Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Edit target: supplied Readlane pixel-art README cover. Improve this successful concept, preserving the wide 3:1 composition, dark midnight canvas, violet and lilac limited palette, crisp chunky pixel art with dithered depth, floating islands, winding reading lane, and very large readable off-white "Readlane" wordmark on the left. Tie the right-hand digital world specifically to collecting and reading blogs: transform several abstract towers into elegant upright PIXEL ARTICLE PANELS with a bold headline bar and a few short horizontal pixel lines, no tiny legible text. Give a few panels bookmark-shaped silhouettes with a notched bottom. A few small pixel pages travel from separate floating islands into the central violet lane, suggesting chosen blog feeds coming together. Integrate a single subtle recognizable RSS broadcast glyph into a distant panel or sign, violet rather than orange. Keep the world abstract and imaginative, not a literal screenshot, not realistic books or a physical library, no browser chrome. Fewer, larger meaningful forms rather than clutter. LEFT typography hierarchy: preserve giant perfectly spelled "Readlane" as primary. Beneath it, put exactly "$ readlane" in medium-large lavender monospace pixel type, followed by a clearly separated solid block cursor. Beneath that put the exact purpose line "RSS & Atom feeds." and on the next line "Read in your terminal." in clean, bright off-white readable sans-serif type, NOT tiny or dim. Those two purpose lines should be at least half the height of the command letters; generous vertical spacing, clear hierarchy, all text on quiet empty dark background. Do not let any landscape overlap the text. Exact text only: "Readlane", "$ readlane", "RSS & Atom feeds.", "Read in your terminal.". No other words. Keep all text easily readable at a 700px total image width. Premium indie-game title art meets considered developer tool. No blur, no glow haze, no scanlines, no gradient backgrounds, no green hacker code, no watermark. Opaque background.

The selected PNG was inspected for spelling and composition, then converted
without resizing or cropping:

```sh
magick source.png -strip -define webp:lossless=true readlane-cover.webp
exiftool -overwrite_original -all= \
  -XMP-dc:Creator="Abdulrahman M. Yaqyn" \
  -XMP-dc:Rights="Copyright Abdulrahman M. Yaqyn" \
  -XMP-x:XMPToolkit= readlane-cover.webp
magick compare -metric AE source.png readlane-cover.webp null:
exiftool -G1 -a -s readlane-cover.webp
```

A 700-pixel-wide preview was inspected for text readability.
The pixel comparison reported **0 differences**. ExifTool confirmed a lossless
WebP at 2172×724 with only the two author-credit fields as descriptive metadata.
Structural image properties and filesystem timestamps are not embedded creator
metadata. The tracked final asset is `readme-assets/readlane-cover.webp`.
