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

The cover uses a large monospace wordmark, violet prompt and block cursor,
and a high-contrast tagline. Small interface text is deliberately omitted so
the artwork stays readable at README widths. It is brand artwork, not a product
screenshot. Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Asset type: revised Readlane README cover. Edit the supplied cover for excellent readability when displayed at only 800 pixels wide. Replace the dense multi-column reading list entirely. Keep the exact name "Readlane" and exact tagline "Your feeds. Your pace." Keep terminal-first identity, midnight background and violet palette. New composition: approximately 3:1 wide landscape, completely flat solid near-black #11121C canvas with a single oversized central terminal-style wordmark reading "> Readlane" in very bold crisp monospaced type, bright near-white #FAFAFF, the > glyph violet #B9B0FF, and a large solid violet block cursor immediately following the name. Below, the exact tagline "Your feeds. Your pace." in large medium-weight near-white type, high contrast, approximately one third the height of the wordmark. Very generous breathing room and balanced central alignment. Wordmark fills about 75% of the canvas width. Terminal feel comes ONLY from the monospace wordmark, > prompt and solid cursor, not from miniature UI. Remove ALL existing reading-list panels, rows, navigation, small commands, publisher names, timestamps, rules and footer hints. NO other text beyond "> Readlane" and "Your feeds. Your pace.". No texture, no scanlines, no fading, no highlights crossing letters, no blur, no glow, no gradients, no perspective, no 3D, no browser chrome, no decorative patterns, no watermark. Perfect spelling, clean letterforms, premium developer-tool typographic cover. Opaque background.

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

An 800-pixel-wide preview was inspected for text readability.
The pixel comparison reported **0 differences**. ExifTool confirmed a lossless
WebP at 2172×724 with only the two author-credit fields as descriptive metadata.
Structural image properties and filesystem timestamps are not embedded creator
metadata. The tracked final asset is `readme-assets/readlane-cover.webp`.
