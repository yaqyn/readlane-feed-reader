# Readlane reader interaction contract

## Business sources
The working CLI, SQL subscription joins, README security limits, and
`docs/operations.md` define a trusted local reader. The web companion cannot
register users, change subscriptions, reset data, or fetch remote feed URLs.
A selected local user is fixed for the server lifetime and is not a login.

## Canonical UI map
| Capability | Canonical owner | Source of truth | Variants | Verification |
| --- | --- | --- | --- | --- |
| Form | reader.html search form and reader.js | This contract | search | browser and API tests |
| Select/Listbox | Native source select | DESIGN.md | platform-owned popup | keyboard and narrow viewport |
| Scrollbar | reader.css global baseline | DESIGN.md tokens | document scroll | browser computed styles |
| Status | Shared status and error region | reader.js | loading, ready, empty, error | browser state tests |
| Navigation | URL and renderRoute in reader.js | This contract | library, summary | Back, deep links |
| Pagination | Shared page controls | reader.js and bounded API | 12 posts per page | API and browser |

## Workflows
- Library: load selected user's followed posts, ordered by publication date.
- Search: debounce 300ms, suppress during IME composition, cancel superseded
  fetches, and preserve query/source/page in URL state. Explicit Search and
  Clear search act immediately. Clear returns focus to the search field.
- Source: selecting a native option resets page and preserves search.
- Summary: open a stored feed summary, retain list query in the URL, and offer
  Back to reading list plus Read original. External links open a new tab and
  disclose that in their accessible name. No article HTML is executed.
- Pagination: Previous and Next are explicit; stale offsets are clamped.
- Refresh: re-fetch current data without dropping filters. Retain existing
  content on a refresh failure and display a retry action.

## Route document title policy
Library uses `Reading list — Readlane`; summaries use their actual title.
Loading, missing resource, and server errors get honest titles. Unknown routes
show an app-owned not-found view with a link to the reading list.

## Boundaries and recovery
Only loopback listeners are accepted. Unexpected Host, cross-site browser
requests, and foreign Origin values are rejected. API methods are read-only.
No cross-origin access headers are emitted. All queries are scoped to the
server's selected user, including direct post IDs and source filtering.

The empty library points to the actual CLI commands to add a feed and collect.
A zero-result search offers Clear filters. Request failures preserve controls
and provide Retry. Loading reserves content geometry and announces progress.
Keyboard focus moves to the main heading for explicit route changes, not on
search keystrokes or background refresh. Browser Back restores list context.

## Scope
English UI, locale-formatted dates, one light theme, reduced motion and forced
colors supported. No mutations, destructive dialogs, authentication, billing,
local storage of post data, unread states, or fabricated reading-time badges.
