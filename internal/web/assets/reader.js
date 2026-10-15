"use strict";
const $ = (id) => document.getElementById(id);
let controller,
  timer,
  composing = false,
  data,
  requestNumber = 0;
const current = () => new URL(location.href);
const state = () => {
  const u = current();
  return {
    q: u.searchParams.get("q") || "",
    source: u.searchParams.get("source") || "",
    page: u.searchParams.get("page") || "1",
  };
};
function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}
function link(text, href, className = "") {
  const a = element("a", className, text);
  a.href = href;
  return a;
}
function listURL(changes = {}) {
  const u = current();
  u.pathname = "/";
  for (const [key, value] of Object.entries(changes)) {
    if (value && String(value) !== "1") u.searchParams.set(key, value);
    else u.searchParams.delete(key);
  }
  return u.pathname + u.search;
}
function navigate(href, focus = true, replace = false) {
  clearTimeout(timer);
  history[replace ? "replaceState" : "pushState"](
    { scroll: window.scrollY },
    "",
    href,
  );
  renderRoute(focus);
}
function setError(message) {
  $("error-text").textContent = message;
  $("error-panel").hidden = false;
}
function dateLabel(raw) {
  if (!raw) return "Publication date unavailable";
  const d = new Date(raw);
  return Number.isNaN(d.getTime())
    ? "Publication date unavailable"
    : d.toLocaleDateString(undefined, {
        day: "numeric",
        month: "short",
        year: "numeric",
      });
}
function meta(post) {
  const box = element("div", "story-meta");
  box.append(element("span", "source-name", post.source));
  const date = element("time", "", dateLabel(post.published_at));
  if (post.published_at) date.dateTime = post.published_at;
  box.append(date);
  return box;
}
function originalLink(post, button = false) {
  if (!post.url) return null;
  const a = link("Read original", post.url, button ? "button primary" : "");
  a.target = "_blank";
  a.rel = "noopener noreferrer";
  a.setAttribute(
    "aria-label",
    `Read original: ${post.title} (opens in a new tab)`,
  );
  return a;
}
function sourceNavigation(result) {
  const s = state();
  $("reader-name").textContent = result.reader;
  const links = $("source-links");
  links.replaceChildren();
  const all = link(
    "All sources",
    listURL({ source: "", page: "" }),
    "source-link",
  );
  if (!s.source) all.setAttribute("aria-current", "page");
  links.append(all);
  const select = $("source-select");
  select.replaceChildren(new Option("All sources", ""));
  for (const source of result.sources) {
    const a = link(
      source.name,
      listURL({ source: source.id, page: "" }),
      "source-link",
    );
    if (s.source === source.id) a.setAttribute("aria-current", "page");
    links.append(a);
    select.add(new Option(source.name, source.id));
  }
  select.value = s.source;
  $("source-note").textContent = result.sources_truncated
    ? "First 200 sources shown. Use readlane sources for the complete registry."
    : result.sources.length === 0
      ? "Follow a source in your terminal to start collecting."
      : "";
  $("library-link").setAttribute("aria-current", "page");
}
function empty(title, text, commands = false) {
  const box = element("div", "empty-state");
  box.append(element("h2", "", title), element("p", "", text));
  if (commands) {
    box.append(
      element(
        "pre",
        "",
        `readlane addfeed "Lane's Blog" https://www.wagslane.dev/index.xml\nreadlane collect 1m --once`,
      ),
    );
  } else {
    const clear = element("button", "button", "Clear filters");
    clear.type = "button";
    clear.addEventListener("click", () => navigate("/"));
    box.append(clear);
  }
  return box;
}
function drawList(result, focus) {
  data = result;
  const s = state();
  sourceNavigation(result);
  $("page-title").textContent = "A little room to read.";
  $("intro").textContent =
    "Stories from the sources you chose. Ready when you are.";
  document.title = "Reading list — Readlane";
  $("search-form").hidden = false;
  $("order-note").hidden = false;
  $("result-count").textContent =
    `${result.total.toLocaleString()} ${result.total === 1 ? "story" : "stories"}${s.q ? " matching your search" : ""}${s.source ? " from this source" : ""}`;
  const content = $("content");
  content.replaceChildren();
  if (!result.posts.length) {
    const filtered = Boolean(s.q || s.source);
    content.append(
      empty(
        filtered ? "No stories found." : "Your reading list starts here.",
        filtered
          ? "Try another phrase or return to all your sources."
          : "Add a source and collect your first stories in the terminal. They will appear here when you refresh.",
        !filtered,
      ),
    );
  }
  for (const post of result.posts) {
    const article = element("article", "story");
    article.append(meta(post));
    const heading = element("h2");
    const detailURL = `/posts/${post.id}${current().search}`;
    heading.append(link(post.title, detailURL));
    article.append(heading);
    if (post.summary)
      article.append(element("p", "story-summary", post.summary));
    const actions = element("div", "story-actions");
    actions.append(link("Read summary", detailURL));
    const external = originalLink(post);
    if (external) actions.append(external);
    article.append(actions);
    content.append(article);
  }
  $("pagination").hidden = result.total === 0;
  $("previous").disabled = result.page <= 1;
  $("next").disabled = result.page * result.page_size >= result.total;
  $("page-label").textContent =
    `Page ${result.page} of ${Math.max(1, Math.ceil(result.total / result.page_size))}`;
  if (Number(s.page) !== result.page) {
    const u = current();
    if (result.page > 1) u.searchParams.set("page", result.page);
    else u.searchParams.delete("page");
    history.replaceState({}, "", u.pathname + u.search);
  }
  if (focus) $("page-title").focus({ preventScroll: true });
}
function drawDetail(post, focus) {
  document.title = `${post.title} — Readlane`;
  $("page-title").textContent = "Stay with the story.";
  $("intro").textContent =
    "A summary from your chosen source. The full article is one click away.";
  $("search-form").hidden = true;
  $("pagination").hidden = true;
  $("order-note").hidden = true;
  $("result-count").textContent = "Feed summary";
  const article = element("article", "detail");
  article.append(link("Back to reading list", listURL(), "back-link"));
  article.append(element("h2", "", post.title), meta(post));
  article.append(
    element(
      "p",
      "detail-body",
      post.summary ||
        "This feed did not provide a summary. Read the original article for the full story.",
    ),
  );
  article.append(
    element(
      "p",
      "detail-note",
      "This is content supplied by the feed, not a copy of the full article.",
    ),
  );
  const external = originalLink(post, true);
  if (external) article.append(external);
  $("content").replaceChildren(article);
  if (focus) {
    window.scrollTo(0, 0);
    $("page-title").focus({ preventScroll: true });
  }
}
async function renderRoute(focus = false) {
  const request = ++requestNumber;
  controller?.abort();
  controller = new AbortController();
  const active = controller;
  const s = state();
  $("search").value = s.q;
  $("clear-search").hidden = !s.q;
  $("error-panel").hidden = true;
  $("content").setAttribute("aria-busy", "true");
  $("refresh").disabled = true;
  $("result-count").textContent = "Opening your reading list…";
  const path = location.pathname;
  let api = "/api/reader?" + new URLSearchParams(s);
  const match = path.match(/^\/posts\/([0-9a-f-]+)$/i);
  if (path !== "/" && !match) {
    $("content").replaceChildren(
      empty(
        "This page has moved.",
        "Return to your reading list to choose another story.",
      ),
    );
    $("page-title").textContent = "Page not found";
    document.title = "Page not found — Readlane";
    $("pagination").hidden = true;
    $("content").setAttribute("aria-busy", "false");
    $("refresh").disabled = false;
    return;
  }
  if (match) api = "/api/posts/" + match[1];
  try {
    const response = await fetch(api, {
      signal: active.signal,
      headers: { Accept: "application/json" },
    });
    const result = await response.json();
    if (!response.ok)
      throw Object.assign(
        new Error(
          result.error || "Your reading list could not be loaded. Try again.",
        ),
        { status: response.status },
      );
    if (request !== requestNumber) return;
    if (match) {
      drawDetail(result, focus);
      if (!data) {
        fetch("/api/reader?" + new URLSearchParams(s), {
          signal: active.signal,
        })
          .then((r) => {
            if (!r.ok) throw new Error("Sources unavailable");
            return r.json();
          })
          .then((context) => {
            if (request === requestNumber) {
              data = context;
              sourceNavigation(context);
            }
          })
          .catch((error) => {
            if (error.name !== "AbortError" && request === requestNumber) {
              $("reader-name").textContent = "Local reader";
              $("source-note").textContent =
                "Sources unavailable. Refresh to try again.";
            }
          });
      }
    } else drawList(result, focus);
  } catch (error) {
    if (error.name === "AbortError" || request !== requestNumber) return;
    const message =
      navigator.onLine === false
        ? "You are offline. Your current stories are still here; reconnect and try again."
        : error instanceof TypeError
          ? "The local reader is unavailable. Check that readlane serve is running, then try again."
          : error.message ||
            "Your reading list could not be loaded. Try again.";
    setError(message);
    $("result-count").textContent = "Could not refresh your reading list.";
    document.title =
      error.status === 404
        ? "Post not found — Readlane"
        : "Reader unavailable — Readlane";
    if (error.status === 404) {
      $("content").replaceChildren(
        empty(
          "This story is no longer here.",
          "It may belong to a source you no longer follow.",
        ),
      );
      $("pagination").hidden = true;
    }
  } finally {
    if (request === requestNumber) {
      $("content").setAttribute("aria-busy", "false");
      $("refresh").disabled = false;
    }
  }
}
function commitSearch() {
  if (composing) return;
  const query = $("search").value.trim();
  navigate(listURL({ q: query, page: "" }), false);
}
$("search").addEventListener("input", () => {
  $("clear-search").hidden = !$("search").value;
  clearTimeout(timer);
  if (!composing) timer = setTimeout(commitSearch, 300);
});
$("search").addEventListener("compositionstart", () => {
  composing = true;
  clearTimeout(timer);
});
$("search").addEventListener("compositionend", () => {
  composing = false;
  clearTimeout(timer);
  timer = setTimeout(commitSearch, 300);
});
$("search").addEventListener("keydown", (event) => {
  if (event.key === "Enter" && event.isComposing) event.preventDefault();
});
$("search-form").addEventListener("submit", (event) => {
  event.preventDefault();
  clearTimeout(timer);
  commitSearch();
});
$("clear-search").addEventListener("click", () => {
  clearTimeout(timer);
  $("search").value = "";
  $("search").focus();
  navigate(listURL({ q: "", page: "" }), false);
});
$("source-select").addEventListener("change", (event) =>
  navigate(listURL({ source: event.target.value, page: "" }), false),
);
$("refresh").addEventListener("click", () => renderRoute());
$("retry").addEventListener("click", () => renderRoute());
$("previous").addEventListener("click", () =>
  navigate(listURL({ page: Math.max(1, Number(state().page) - 1) })),
);
$("next").addEventListener("click", () =>
  navigate(listURL({ page: Number(state().page) + 1 })),
);
$("copy-command").addEventListener("click", async () => {
  try {
    await navigator.clipboard.writeText("readlane collect 1m");
    $("copy-status").textContent = "Command copied.";
  } catch {
    $("copy-status").textContent = "Select and copy the command above.";
  }
});
document.addEventListener("click", (event) => {
  const a = event.target.closest("a");
  if (
    !a ||
    a.target ||
    event.ctrlKey ||
    event.metaKey ||
    event.shiftKey ||
    event.altKey ||
    event.button !== 0
  )
    return;
  const u = new URL(a.href);
  if (
    u.origin === location.origin &&
    !u.hash &&
    !u.pathname.startsWith("/assets/")
  ) {
    event.preventDefault();
    history.replaceState({ scroll: window.scrollY }, "", location.href);
    navigate(u.pathname + u.search);
  }
});
window.addEventListener("popstate", async () => {
  const y = history.state?.scroll || 0;
  await renderRoute(false);
  window.scrollTo(0, y);
});
document.querySelectorAll("button[data-reader-action]").forEach((button) => {
  button.disabled = false;
});
renderRoute();
