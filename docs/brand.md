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

The cover foregrounds a dark terminal reading list, a violet selection, a
command prompt, and keyboard hints. It is illustrative artwork, not a product
screenshot; story and publication pairings are fictional examples. Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Asset type: replacement README cover for Readlane, a terminal-first RSS and Atom reader. Input image is the edit target and brand reference. Redesign it to strongly communicate a polished terminal product; remove the paper sheets and ribbon illustration completely. Keep the exact brand name "Readlane" and tagline "Your feeds. Your pace.", violet identity, and wide approximately 3:1 landscape composition. Make the entire canvas a deep midnight terminal background #171824, with restrained violet #AEA5FF accents and off-white #F0EFFA lettering. Left third: large perfectly spelled Readlane wordmark, smaller exact tagline, then a small monospace command "$ readlane" with a block cursor. Right two thirds: a crisp, flat, front-facing terminal reading list, editorially composed, generous spacing, fine box-drawing divider lines, no browser chrome, no laptop/device, no perspective. Terminal header "Readlane", navigation "Reading list   Sources", caption "All sources". Three short representative story titles: "A quieter way to keep up", "Small tools, lasting habits", "Notes from the open web". First row selected with a muted violet rectangular highlight and a simple > indicator. Footer in monospace: "j/k move   enter read   / search   q quit". Accurate readable typography, sparse terminal interface, aligned monospaced columns. This is illustrative cover artwork, not a claimed screenshot; keep it plausible and faithful to supported reading-list behavior. Premium developer-tool aesthetic, restrained, calm, strong typographic hierarchy; no neon, no green hacker code, no gradients, no glow, no paper illustration, no browser interface, no extra invented features or logos, no watermark. Opaque background.

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
