# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: audit/audit.spec.ts >> audit pipeline
- Location: audit/audit.spec.ts:68:1

# Error details

```
Error: p01-public-landing@desktop reason: missing selectors: [data-testid='shell'], [data-testid='nav-home']

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "selector_missing"
```

```
Error: p01-public-landing@tablet reason: missing selectors: [data-testid='shell'], [data-testid='nav-home']

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "selector_missing"
```

```
Error: p01-public-landing@mobile reason: body width 528 > viewport 390

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "layout_collapsed"
```

```
Error: p02-sensors@desktop reason: missing selectors: [data-testid='nav-sensors']

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "selector_missing"
```

```
Error: p02-sensors@mobile reason: body width 527 > viewport 390

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "layout_collapsed"
```

```
Error: p03-agents@desktop reason: missing selectors: [data-testid='nav-agents']

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "selector_missing"
```

```
Error: p04-memory@desktop reason: missing selectors: [data-testid='nav-memory']

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "selector_missing"
```

```
Error: p05-design@desktop reason: 1 console errors

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "console_error"
```

```
Error: p06-roadmap@desktop reason: 1 console errors

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "console_error"
```

```
Error: p07-settings@desktop reason: 1 console errors

expect(received).toBe(expected) // Object.is equality

Expected: "passed"
Received: "console_error"
```

# Test source

```ts
  74  |       : map.viewports;
  75  |     for (const viewport of targetViewports) {
  76  |       await runFeature(browser, feature, viewport);
  77  |     }
  78  |   }
  79  | });
  80  | 
  81  | async function runFeature(browser: import("@playwright/test").Browser, feature: Feature, viewport: Viewport) {
  82  |   const start = Date.now();
  83  |   const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height } });
  84  |   const page = await context.newPage();
  85  |   const consoleErrors: Result["console_errors"] = [];
  86  |   const networkErrors: Result["network_errors"] = [];
  87  | 
  88  |   page.on("console", (msg) => {
  89  |     if (msg.type() === "error") {
  90  |       consoleErrors!.push({ feature_id: feature.id, viewport: viewport.name, severity: SEVERITY_ERROR, message: msg.text() });
  91  |     }
  92  |   });
  93  |   page.on("requestfailed", (req: Request) => {
  94  |     networkErrors!.push({ feature_id: feature.id, viewport: viewport.name, url: req.url(), status: 0, method: req.method() });
  95  |   });
  96  |   page.on("response", (res: Response) => {
  97  |     if (res.status() >= 400) {
  98  |       networkErrors!.push({ feature_id: feature.id, viewport: viewport.name, url: res.url(), status: res.status(), method: res.request().method() });
  99  |     }
  100 |   });
  101 | 
  102 |   const url = feature.route.startsWith("http") ? feature.route : `${BASE_URL}${feature.route}`;
  103 |   let httpStatus = 0;
  104 |   let status = STATUS_PASSED;
  105 |   let reason = "";
  106 |   const missing: string[] = [];
  107 |   let layout: Result["layout"] | undefined;
  108 |   let screenshot: string | undefined;
  109 | 
  110 |   try {
  111 |     const response = await page.goto(url, { waitUntil: "networkidle", timeout: 15_000 });
  112 |     if (response) {
  113 |       httpStatus = response.status();
  114 |       if (httpStatus !== feature.expected_http_status) {
  115 |         status = STATUS_API_ERROR;
  116 |         reason = `expected http ${feature.expected_http_status}, got ${httpStatus}`;
  117 |       }
  118 |     }
  119 |     if (feature.expected_selectors?.length) {
  120 |       for (const selector of feature.expected_selectors) {
  121 |         const locator = page.locator(selector).first();
  122 |         const count = await locator.count();
  123 |         if (count === 0) {
  124 |           missing.push(selector);
  125 |         }
  126 |       }
  127 |       if (missing.length) {
  128 |         status = STATUS_SELECTOR_MISSING;
  129 |         reason = `missing selectors: ${missing.join(", ")}`;
  130 |       }
  131 |     }
  132 |     const dim = await page.evaluate(() => ({
  133 |       bodyWidth: document.body?.scrollWidth ?? 0,
  134 |       viewportWidth: window.innerWidth,
  135 |     }));
  136 |     layout = {
  137 |       feature_id: feature.id,
  138 |       viewport: viewport.name,
  139 |       body_width: dim.bodyWidth,
  140 |       viewport_width: dim.viewportWidth,
  141 |       has_horizontal_scroll: dim.bodyWidth > dim.viewportWidth + 1,
  142 |     };
  143 |     if (layout.has_horizontal_scroll) {
  144 |       status = STATUS_LAYOUT_COLLAPSED;
  145 |       reason = `body width ${dim.bodyWidth} > viewport ${dim.viewportWidth}`;
  146 |     }
  147 |     if (consoleErrors!.length) {
  148 |       status = STATUS_CONSOLE_ERROR;
  149 |       reason = `${consoleErrors!.length} console errors`;
  150 |     }
  151 |     screenshot = path.join("current", "screenshots", `${feature.id}-${viewport.name}.png`);
  152 |     await page.screenshot({ path: path.join(AUDIT_OUT, screenshot), fullPage: true });
  153 |   } catch (err) {
  154 |     status = STATUS_FAILED;
  155 |     reason = err instanceof Error ? err.message : String(err);
  156 |   }
  157 | 
  158 |   await context.close();
  159 |   results.push({
  160 |     feature_id: feature.id,
  161 |     viewport: viewport.name,
  162 |     status,
  163 |     reason,
  164 |     url_final: page.url(),
  165 |     http_status: httpStatus,
  166 |     screenshot,
  167 |     console_errors: consoleErrors!.length ? consoleErrors : undefined,
  168 |     network_errors: networkErrors!.length ? networkErrors : undefined,
  169 |     missing_selectors: missing.length ? missing : undefined,
  170 |     layout,
  171 |     duration_ms: Date.now() - start,
  172 |     recorded_at: new Date().toISOString(),
  173 |   });
> 174 |   expect.soft(status, `${feature.id}@${viewport.name} reason: ${reason}`).toBe(STATUS_PASSED);
      |                                                                           ^ Error: p07-settings@desktop reason: 1 console errors
  175 | }
  176 | 
```