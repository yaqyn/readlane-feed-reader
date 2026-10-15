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
a unified dollar-prefixed title, and the purpose line
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

> Use case: precise-object-edit. Edit the supplied Readlane pixel-world cover. Preserve the right-hand pixel-art reading world exactly: article panels, bookmark towers, floating posts, RSS sign, violet pathway, dark palette and crisp pixel style. Improve ONLY the left typography into a confident unified title. Remove both the existing large "Readlane" and the separate "$ readlane" command. Replace them with ONE single large wordmark reading exactly "$ Readlane", all on one horizontal baseline, with a lavender dollar prompt and bright off-white Readlane in the same bold readable pixel-edged lettering as the current title. Dollar symbol approximately 65 percent the title height, visually integrated with generous one-character spacing; entire title occupies left 46 percent of canvas without touching the world. Name is displayed ONCE only. No cursor attached to the title. Beneath the title preserve the two-line clear purpose statement exactly "RSS & Atom feeds." then "Read in your terminal.", slightly larger than current if space allows, off-white, generous line spacing. Position the left text block vertically centered as one cohesive lockup. No redundant command, no repeated name, no extra text. Keep Readlane perfectly spelled and very readable at 700px overall width, normal recognizable lowercase l. Keep all illustration, dimensions, palette and background unchanged. Opaque background.

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
