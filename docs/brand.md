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

The cover's expressive element is a violet ribbon carrying paper summaries.
Everything around it stays restrained. Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Asset type: final landscape README cover for a professional local RSS and Atom reader named Readlane. Create a polished high-end brand image at wide landscape approximately 3:1, with the exact app name "Readlane" prominently and perfectly spelled R-e-a-d-l-a-n-e in large elegant modern humanist sans-serif lettering, left aligned with generous breathing room, and exact smaller tagline "Your feeds. Your pace." underneath. Theme: a quiet personal reading lane. Cool porcelain-white backdrop #F5F6FB with deep navy ink #20253B typography. On the right, an original sculptural illustration of several thin layered reading sheets flowing along a softly curved, broad iris-violet ribbon #6253D9, with two subtle pale lilac bookmark tabs; tactile matte paper, subtle material shadows, precise composition, editorial product art rather than a software UI screenshot. Use cool lilac #EAE7FA and slate #667085 supporting colors. Keep text on plain negative space, avoid placing imagery behind the letters. Professional, calm, recognizable, purposeful. Restrained violet color, no gradients as decoration, no neon, no stock-photo people, no mascots, no terminal screenshots, no invented product features, no extra words, no small fake text, no watermark. The name must remain easily readable at GitHub README widths. Opaque background.

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

The pixel comparison reported **0 differences**. ExifTool confirmed a lossless
WebP at 2172×724 with only the two author-credit fields as descriptive metadata.
Structural image properties and filesystem timestamps are not embedded creator
metadata. The tracked final asset is `readme-assets/readlane-cover.webp`.
