const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const { JSDOM } = require("jsdom");
const axe = require("axe-core");
const html = fs.readFileSync("internal/web/assets/reader.html", "utf8");
const script = fs.readFileSync("internal/web/assets/reader.js", "utf8");
const post = {
  id: "00000000-0000-0000-0000-000000000001",
  title: "A useful story",
  source: "Example",
  source_id: "00000000-0000-0000-0000-000000000002",
  url: "https://example.com/post",
  summary: "A clear summary.",
  published_at: "2026-10-14T08:00:00Z",
};
const library = {
  posts: [post],
  sources: [{ id: post.source_id, name: "Example" }],
  reader: "alice",
  total: 1,
  page: 1,
  page_size: 12,
  sources_truncated: false,
};
const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
async function until(predicate) {
  for (let i = 0; i < 100; i++) {
    if (predicate()) return;
    await wait(10);
  }
  assert.fail("Reader did not reach the expected state");
}
function setup(t, responder = () => library, path = "/") {
  const dom = new JSDOM(html, {
    url: "http://127.0.0.1:8090" + path,
    runScripts: "outside-only",
    pretendToBeVisual: true,
  });
  const w = dom.window;
  w.scrollTo = () => {};
  w.HTMLCanvasElement.prototype.getContext = () => null;
  w.fetch = async (url, options) => {
    const result = await responder(url, options);
    return {
      ok: !result.status || result.status < 400,
      status: result.status || 200,
      json: async () => result.body || result,
    };
  };
  w.eval(script);
  t.after(() => w.close());
  return w;
}
function submit(w, value) {
  w.document.getElementById("search").value = value;
  w.document
    .getElementById("search-form")
    .dispatchEvent(new w.Event("submit", { bubbles: true, cancelable: true }));
}
test("renders actual data, semantic links and script-owned actions", async (t) => {
  const w = setup(t);
  await until(() => w.document.querySelector(".story"));
  assert.match(
    w.document.getElementById("result-count").textContent,
    /1 story/,
  );
  const a = w.document.querySelector('.story-actions a[target="_blank"]');
  assert.match(a.rel, /noopener/);
  assert.match(a.getAttribute("aria-label"), /new tab/);
  assert.equal(w.document.getElementById("refresh").disabled, false);
  w.eval(axe.source);
  const result = await w.axe.run(w.document, {
    rules: { "color-contrast": { enabled: false } },
  });
  assert.equal(
    result.violations.length,
    0,
    result.violations.map((v) => v.id).join(", "),
  );
});
test("feed text is text rather than executable markup", async (t) => {
  const w = setup(t, () => ({
    ...library,
    posts: [
      {
        ...post,
        title: "<img src=x onerror=alert(1)>",
        summary: "<script>bad()</script>",
      },
    ],
  }));
  await until(() => w.document.querySelector(".story"));
  assert.equal(
    w.document.querySelectorAll("#content img,#content script").length,
    0,
  );
  assert.match(w.document.querySelector(".story h2").textContent, /<img/);
});
test("search persists URL state, clears immediately, and ignores stale replies", async (t) => {
  const w = setup(t, async (url) => {
    const q = new URL(url, "http://127.0.0.1:8090").searchParams.get("q");
    if (q === "old") await wait(70);
    return { ...library, posts: [{ ...post, title: q || post.title }] };
  });
  await until(() => w.document.querySelector(".story"));
  submit(w, "old");
  submit(w, "new");
  await until(
    () => w.document.querySelector(".story h2").textContent === "new",
  );
  await wait(100);
  assert.equal(w.document.querySelector(".story h2").textContent, "new");
  assert.match(w.location.search, /q=new/);
  w.document.getElementById("clear-search").click();
  await until(() => !w.location.search.includes("q="));
  assert.equal(w.document.activeElement.id, "search");
});
test("empty and no-results states provide different recovery", async (t) => {
  const w = setup(t, () => ({ ...library, posts: [], sources: [], total: 0 }));
  await until(() => w.document.querySelector(".empty-state h2"));
  assert.match(
    w.document.getElementById("content").textContent,
    /readlane collect/,
  );
  submit(w, "unmatched");
  await until(() =>
    w.document
      .getElementById("content")
      .textContent.includes("No stories found"),
  );
  assert.match(
    w.document.getElementById("content").textContent,
    /Clear filters/,
  );
});
test("refresh failure preserves stories and a retry can recover", async (t) => {
  let fail = false;
  const w = setup(t, () =>
    fail
      ? { status: 503, body: { error: "Database unavailable. Try again." } }
      : library,
  );
  await until(() => w.document.querySelector(".story"));
  fail = true;
  w.document.getElementById("refresh").click();
  await until(() => !w.document.getElementById("error-panel").hidden);
  assert.ok(w.document.querySelector(".story"));
  assert.match(
    w.document.getElementById("error-text").textContent,
    /Database unavailable/,
  );
  fail = false;
  w.document.getElementById("retry").click();
  await until(() => w.document.getElementById("error-panel").hidden);
  assert.ok(w.document.querySelector(".story"));
});
test("post navigation preserves search context and exposes a back path", async (t) => {
  const w = setup(
    t,
    (url) => (url.startsWith("/api/posts/") ? post : library),
    "/?q=useful",
  );
  await until(() => w.document.querySelector(".story h2 a"));
  w.document.querySelector(".story h2 a").click();
  await until(() => w.document.querySelector(".detail"));
  assert.match(w.location.pathname, /posts/);
  assert.match(w.location.search, /q=useful/);
  assert.equal(w.document.title, "A useful story — Readlane");
  assert.match(w.document.querySelector(".back-link").href, /q=useful/);
});
test("IME composition does not dispatch a half-composed search", async (t) => {
  let calls = 0;
  const w = setup(t, () => {
    calls++;
    return library;
  });
  await until(() => w.document.querySelector(".story"));
  const input = w.document.getElementById("search");
  input.dispatchEvent(new w.Event("compositionstart"));
  input.value = "日本";
  input.dispatchEvent(new w.Event("input"));
  await wait(350);
  assert.equal(calls, 1);
  input.dispatchEvent(new w.Event("compositionend"));
  await until(() => calls === 2);
  assert.match(w.location.search, /q=/);
});
