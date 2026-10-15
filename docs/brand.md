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

The cover imagines an abstract pixel world inside a terminal: floating digital
platforms, dithered depth, and a violet path against a dark canvas. A large
Readlane wordmark and a separate pixel-art dollar prompt with block cursor
anchor the composition. It evokes the terminal without reproducing a screen.
Product text uses concrete labels:
Reading list, Sources, Search, Feed summary, and Read original.

`DESIGN.md` is the durable visual contract and token source. The default web
palette is porcelain, white, violet, and deep ink. The terminal adapts to light
or dark host colors. Manrope and Source Serif 4 are the self-hosted companion
fonts; terminals use their user's own typeface. OFL license files are bundled.

## Cover production

The built-in image generation tool produced the cover. The final prompt was:

> Use case: ads-marketing. Asset type: imaginative Readlane README cover, wide landscape approximately 3:1. Create a distinctive art-directed PIXEL ART world that evokes life inside a terminal without literally depicting a terminal window or computer. Deep almost-black navy background #11121C, limited palette of off-white, slate, electric lilac and violet #AEA5FF, subtle tiny mint accents. On the RIGHT half: a surreal miniature digital landscape assembled from chunky pixel clusters, floating stepped platforms, crystalline rectangular towers, and a winding luminous violet pixel pathway suggesting a personal lane through information. Inventive negative space, deep atmospheric layers made through discrete dither patterns and hard-edged pixel shading, beautifully composed like a premium indie-game title screen. Objects should be abstract digital architecture, not realistic buildings, books, mountains or browser UI. Sparse tiny square particles suggest packets of stories traveling along the lane. Terminal feeling comes from disciplined alignment, character-cell proportions, restrained phosphor-like colors, dark negative space and command-line confidence; NOT from ASCII illustrations, code dumps, terminal frames or screenshots. LEFT 45 percent: very large clear exact wordmark "Readlane", R-e-a-d-l-a-n-e, in bold clean readable off-white geometric sans-serif letters with a subtle stepped pixel treatment only at outer corners, normal recognizable lowercase l. Wordmark around 40 percent of total canvas width, with generous empty dark space around it, nothing behind letters. No tagline, no extra text, no command symbols attached to letters. A handful of small violet pixels near the name ties the wordmark into the world. Crisp graphic contours, intentional chunky pixels, no smoothing blur, no gradients, no neon haze, no scanlines crossing letters, no laptop, no literal terminal UI, no paper/book illustration, no watermark. Creative, welcoming, memorable, sophisticated pixel art rather than generic retro hacking. The name must be effortlessly readable when the whole image is displayed at 600px wide. Opaque background.
>
> Final edit: Use case: precise-object-edit. Edit this Readlane pixel-art cover. Preserve the entire floating violet pixel world, dark palette, large readable wordmark, composition, dimensions and all existing artwork. Add a clearly recognizable violet "$" shell-prompt symbol in chunky crisp pixel-art lettering in the empty dark area directly beneath the LEFT side of the Readlane wordmark, followed by a small separated solid violet block cursor: "$ ▋". Make these symbols large enough to recognize at 600px overall image width, about one third of the wordmark letter height. This is the familiar terminal prompt anchoring the imaginary digital world. Keep ample spacing from the wordmark, do not obscure letters. No other changes and no other added text. Opaque background.

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
