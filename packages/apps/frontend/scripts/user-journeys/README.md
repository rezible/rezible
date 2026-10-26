# User journeys

Run repeatable screenshot walkthroughs against an already-running development stack.
Requires Bun with `Bun.WebView` (verified with 1.3.13), Chrome/Chromium, and `APP_URL`.
Bun discovers Chrome automatically; set `BUN_CHROME_PATH` if necessary. Local HTTPS
must be trusted by Chrome. The runner never starts services or seeds data.

The journey browser disables HTTP/2 to avoid an observed Chrome startup verifier
refresh aborting script requests (`ERR_CERT_VERIFIER_CHANGED`). TLS certificate
validation stays enabled. These captures do not test HTTP/2 transport behavior.
Each plan's browser uses a temporary profile that the runner removes afterwards,
because Bun leaves its own Chrome profile directories behind.

From the repository root:

```sh
just frontend::user-journeys
just frontend::user-journeys home-to-situations
just frontend::user-journeys home-to-situations another-plan
```

The command type-checks the runner and all plans before opening a browser.
The running SvelteKit dev server maintains the generated route types. If running
without it, generate types first with `just frontend::run sync`. Avoid running
sync alongside the dev server because it can trigger reloads. The Just recipe forwards to the frontend package’s
`bun run user-journeys` command.

No arguments runs all `plans/*.ts` files in filename order. Arguments select plans
in the supplied order. Each plan runs sequentially in its own Bun/browser process,
with fresh storage. The first failure stops the run and returns a nonzero exit code.

## Authoring

Export a declarative object with `satisfies JourneyPlan`; see
`plans/home-to-situations.ts`. Plan names come from filenames. Use lowercase
letters, digits, and hyphens for plan and screenshot names. Screenshot names must
be unique within a plan; `failure` is reserved.

Navigation paths use SvelteKit's generated `Pathname` type. Query parameters are
separate. Every run type-checks plans before execution. Dynamic pathname types
do not verify that a database record
exists; follow links from lists instead of embedding record IDs. Import frontend
types only, not modules requiring Vite or the application runtime.

Actions: `navigate`, `waitFor`, `click`, `fill`, `press`, `scrollTo`, `resize`, and
`screenshot`. See `types.ts` for their fields. Selectors are standard CSS, not
Playwright selectors. Targeted actions require one visible match and scroll it
into view. Prefer stable IDs and attributes over styling classes. `fill` replaces
contents using native selection and typing. `press` supports named keys and
optional modifiers.

`waitFor` supports `visible` (one visible element) and `absent` (no matching DOM
elements). Use it to establish meaningful readiness after navigation or an
interaction. Page load alone does not mean application data is ready. Screenshots
wait for fonts and two animation frames, but do not wait for application queries.

The default timeout is 10 seconds; override with `timeoutMs` on a plan or step.
A timeout allows up to two seconds for failure capture, then closes the browser
to cancel pending operations. Arbitrary JavaScript,
branching, full-page screenshots, and automatic retries are not supported.

Plans operate on the live database. Write non-mutating walkthroughs by default;
the Home-to-Situations example only navigates and captures screenshots.
Prerequisites are documentation, not automatic provisioning or checks.

## Output

`output/<plan>/` is gitignored and replaced when that plan starts. It contains
named viewport PNGs and `run.json`: step outcomes, URLs, viewport dimensions,
captions, runtime versions, and console/network diagnostics. Diagnostics are
recorded but do not automatically fail a journey. This is evidence collection,
not an assertion that the application is error-free.

Partial results remain on failure. The runner attempts `failure.html` (the current
rendered DOM) and `failure.png`, including on timeouts before closing the browser.
Capture is best-effort and bounded to two seconds; a crashed or unresponsive browser
may prevent either artifact. HTML is a DOM snapshot, not a self-contained archive
of styles, images, JavaScript state, shadow roots, or iframe contents. Output can
contain development data and request URLs. There is no automatic upload.
