---
{
  "version": "alpha",
  "name": "Readlane",
  "description": "A personal reading lane for RSS and Atom, with a scriptable CLI and a local companion reader.",
  "colors": {
    "primary": "#6253D9",
    "primary-hover": "#5042BD",
    "background": "#F5F6FB",
    "surface": "#FFFFFF",
    "ink": "#20253B",
    "muted": "#626B80",
    "border": "#DADDE8",
    "tint": "#EAE7FA",
    "danger": "#A73446",
    "success": "#25664E",
    "scrollbar": "#ABB0C6",
    "scrollbar-hover": "#727A98",
    "terminal-ink-dark": "#F0EFFA",
    "terminal-muted-dark": "#B3B1C8",
    "terminal-primary-dark": "#AEA5FF",
    "terminal-selected-dark": "#302B4D"
  },
  "typography": {
    "sans": {
      "fontFamily": "Manrope, system-ui, sans-serif"
    },
    "reading": {
      "fontFamily": "Source Serif 4, Georgia, serif"
    },
    "mono": {
      "fontFamily": "ui-monospace, monospace"
    }
  },
  "rounded": {
    "sm": "0.5rem",
    "DEFAULT": "0.75rem",
    "lg": "1.25rem"
  },
  "spacing": {
    "page-max": "80rem",
    "reading-max": "44rem",
    "section-gap": "2.5rem"
  },
  "components": {
    "button": {
      "height": "2.75rem"
    },
    "input": {
      "height": "3rem"
    },
    "article": {
      "padding": "1.75rem"
    }
  }
}
---

# Readlane design system

## Overview

### Creative North Star
A quiet reading lane inside the terminal: aligned monospace text, a violet
selection, and visible keyboard controls. The cover pairs that reading list
with a large wordmark and command prompt. Product screens stay quiet so the
stories carry the attention.

### Product context and register
Readers choose publications, collect them locally, and return when ready.
This is an English-language tool for developers and independent readers; no
geographic market, paid plan, or audience research has been established.
Posts may use any language. Dates display in the browser locale with a full
absolute date available; the interface remains English. Mixed-script content
wraps naturally and uses platform fallbacks.

The terminal is the primary product surface. Its interactive reader uses a compact
source/reading navigation, a spacious post list, a scrollable summary, and a
fixed keyboard-help footer. Normal commands remain scriptable. The optional
web reader is a complementary product surface. The README cover is a brand surface. The
current-user and database trust model comes from the verified application and
`docs/operations.md`: selecting a local user is not authentication. The web
companion reads one selected user's followed posts and makes no mutations.

Token ownership is Model A: this file's JSON-compatible YAML frontmatter is
canonical. `scripts/design-tokens.py` generates `internal/web/assets/tokens.css`
and recolors the canonical README icons. Components consume CSS variables.
`--check` fails on drift. Fonts and layout recipes are shared in `reader.css`.

## Colors
Porcelain is the page background; white holds the reading surface. Ink carries
headings and prose, muted slate carries supporting information, and violet
marks navigation and actions. Tint indicates selection. Danger and success
always have accompanying text. Focus uses a violet outline with an offset.
The web product uses one light theme; the terminal adapts to light/dark host
backgrounds and supports NO_COLOR. forced-colors settings retain native contrast.

## Typography
Manrope is self-hosted for navigation, controls, and the wordmark. Source Serif
4 is self-hosted for article titles and the expanded feed summary. Fonts use
`font-display: optional` to prevent a late swap from moving controls. System
and Georgia fallbacks remain usable. Body type is 16px with generous leading;
article titles start at 27px. Reading text stays within 44rem. Commands use the
platform's monospace face only where they are actual commands.

## Layout
Desktop uses a 15rem source rail and one reading column inside an 80rem shell.
Below 800px the rail becomes a compact header; source choice remains available
through a native select. Below 480px toolbar controls stack. The document owns
vertical scroll, including long article summaries. No fixed-height app shell
or hidden-scrollbar treatment is allowed. Async regions reserve at least 18rem.

## Elevation & Depth
Hierarchy comes from spacing and tonal surfaces. Article rows have a subtle
separator and no shadow. The collection help panel uses tint. The reader has
no decorative gradients, dashboard KPI cards, or constantly animated chrome.

## Shapes
Controls use the default radius; the reading sheet uses the large radius.
The brand mark is a small pair of offset bookmark/page shapes. README icons
retain the workspace's canonical geometry and use the new violet palette.

## Components

### Foundational visual states
Controls define hover, focus-visible, active, disabled, and busy states. The
reader has one status region, one inline error with a retry action, and a
bounded loading indicator. Empty library and no search results are distinct.

### Buttons and actions
Buttons use explicit verbs: Refresh, Clear search, Next, Previous, Copy command.
Primary actions are violet; navigation is neutral until selected. Busy buttons
keep their geometry. External article links say Read original and open a new
browser tab with an accompanying accessible label.

### Navigation and data display
The library is page-based because results must remain revisitable. URL state
owns query, source, and page; opening a post retains that context. Returning
restores the list and its prior position. Article summaries are previews of
stored feed content, not claims to reproduce a full source article.

### Forms and overlays
Search is a labeled, IME-safe field with an explicit clear button and 300ms
remote debounce. Source selection is intentionally native: platform-owned
popup geometry and keyboard behavior are accepted. No authored modal is
needed. Shared buttons, search, empty/error regions, and pagination are owned
by the single embedded reader shell and script, not duplicated per route.

### Iconography
The runtime mark is repository-owned SVG. Icons supplement text; actionable
icons have accessible names. Decorative SVGs have empty alternative text.

### Motion
Only control feedback transitions run, lasting 140ms. Reduced motion removes
transitions and animation. Loading uses a text status and reserved region,
not an ornamental page entrance or repeated reveal sequence.

### Content and data visualization
Use short concrete sentences. The primary tagline is Your feeds. Your pace.
Never invent unread status, reading time, worker activity, user growth,
performance figures, or full-article availability. Display actual counts and
publication dates. CLI output stays plain when piped or NO_COLOR is set.

## Do's and Don'ts
- **Do:** Give stories readable titles, source names, dates, and usable links.
- **Do:** Preserve filters and search when navigating to a summary and back.
- **Don't:** Expose a public web listener without a designed authentication boundary.
- **Don't:** turn documentation into unsupported marketing claims.
