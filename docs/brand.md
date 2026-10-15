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

The cover imagines a violet reading lane connecting floating book islands and
pages. A large dark wordmark sits on a quiet porcelain background. The artwork
uses no terminal interface or small text; the application remains terminal-first.
Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Asset type: imaginative illustrated README cover for Readlane, a personal RSS and Atom feed reader. Create a completely new visual direction, wide landscape approximately 3:1. A surreal editorial paper landscape: an elegant iris-violet pathway curves through floating ivory pages and small sculptural open-book islands, gathering separate reading streams into one calm lane. Tactile matte paper forms, precise soft shadows, airy porcelain background #F5F6FB, restrained violet #6253D9 and lilac #EAE7FA with tiny warm apricot accents. Art occupies the right half and lower-right edge, with delightful imaginative depth, a clear coherent silhouette, beautifully crafted premium book-jacket art. Left half is quiet clean negative space containing ONLY the exact name "Readlane", spelled R-e-a-d-l-a-n-e, in extremely clear, large, bold humanist sans-serif dark ink #20253B. Use a normal recognizable lowercase l, not a slab or terminal glyph. The name must be the first thing read and occupy about 45% of total image width, high contrast with no illustration behind it. No tagline or small text. No terminal, no code, no cursor, no command prompt, no UI, no browser, no device, no neon, no watermark, no fake lettering on pages. Keep the name readable when the entire image is reduced to 600 pixels wide. Imaginative, inviting, sophisticated, calm; editorial illustration rather than a software screenshot. Opaque background.

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

A 600-pixel-wide preview was inspected for text readability.
The pixel comparison reported **0 differences**. ExifTool confirmed a lossless
WebP at 2172×724 with only the two author-credit fields as descriptive metadata.
Structural image properties and filesystem timestamps are not embedded creator
metadata. The tracked final asset is `readme-assets/readlane-cover.webp`.
