import { test, expect, type Page, type Request, type Response } from "@playwright/test";
import { promises as fs } from "node:fs";
import path from "node:path";

type Feature = {
  id: string;
  name: string;
  route: string;
  role: string;
  category: string;
  priority: string;
  expected_http_status: number;
  expected_selectors: string[];
  expected_content?: string[];
  apis_used?: string[];
  viewports?: string[];
};

type Viewport = { name: string; width: number; height: number };

type Result = {
  feature_id: string;
  viewport: string;
  status: string;
  reason?: string;
  url_final?: string;
  http_status?: number;
  screenshot?: string;
  console_errors?: Array<{ feature_id: string; viewport: string; severity: string; message: string }>;
  network_errors?: Array<{ feature_id: string; viewport: string; url: string; status: number; method: string }>;
  missing_selectors?: string[];
  visual?: null;
  layout?: { feature_id: string; viewport: string; has_horizontal_scroll: boolean; body_width: number; viewport_width: number };
  duration_ms: number;
  recorded_at: string;
};

const STATUS_PASSED = "passed";
const STATUS_FAILED = "failed";
const STATUS_SELECTOR_MISSING = "selector_missing";
const STATUS_API_ERROR = "api_error";
const STATUS_CONSOLE_ERROR = "console_error";
const STATUS_LAYOUT_COLLAPSED = "layout_collapsed";
const SEVERITY_ERROR = "error";

const BASE_URL = process.env.AUDIT_BASE_URL ?? "http://127.0.0.1:7373";
const AUDIT_OUT = process.env.AUDIT_OUT ?? path.join(process.cwd(), "tmp", "app-audit", "manual");
const ONLY_FEATURE = process.env.AUDIT_FEATURE ?? "";

const featureMapPath = path.join(AUDIT_OUT, "json", "feature-map.json");
const resultsPath = path.join(AUDIT_OUT, "json", "results.json");
const screenshotsDir = path.join(AUDIT_OUT, "current", "screenshots");

const results: Result[] = [];

test.describe.configure({ mode: "serial" });

test.beforeAll(async () => {
  await fs.mkdir(screenshotsDir, { recursive: true });
});

test.afterAll(async () => {
  const payload = { generated_at: new Date().toISOString(), base_url: BASE_URL, results };
  await fs.mkdir(path.dirname(resultsPath), { recursive: true });
  await fs.writeFile(resultsPath, JSON.stringify(payload, null, 2));
});

test("audit pipeline", async ({ browser }) => {
  const map = JSON.parse(await fs.readFile(featureMapPath, "utf-8")) as { features: Feature[]; viewports: Viewport[] };
  const features = ONLY_FEATURE ? map.features.filter((f) => f.id === ONLY_FEATURE) : map.features;
  for (const feature of features) {
    const targetViewports = feature.viewports?.length
      ? map.viewports.filter((v) => feature.viewports?.includes(v.name))
      : map.viewports;
    for (const viewport of targetViewports) {
      await runFeature(browser, feature, viewport);
    }
  }
});

async function runFeature(browser: import("@playwright/test").Browser, feature: Feature, viewport: Viewport) {
  const start = Date.now();
  const context = await browser.newContext({ viewport: { width: viewport.width, height: viewport.height } });
  const page = await context.newPage();
  const consoleErrors: Result["console_errors"] = [];
  const networkErrors: Result["network_errors"] = [];

  page.on("console", (msg) => {
    if (msg.type() === "error") {
      consoleErrors!.push({ feature_id: feature.id, viewport: viewport.name, severity: SEVERITY_ERROR, message: msg.text() });
    }
  });
  page.on("requestfailed", (req: Request) => {
    networkErrors!.push({ feature_id: feature.id, viewport: viewport.name, url: req.url(), status: 0, method: req.method() });
  });
  page.on("response", (res: Response) => {
    if (res.status() >= 400) {
      networkErrors!.push({ feature_id: feature.id, viewport: viewport.name, url: res.url(), status: res.status(), method: res.request().method() });
    }
  });

  const url = feature.route.startsWith("http") ? feature.route : `${BASE_URL}${feature.route}`;
  let httpStatus = 0;
  let status = STATUS_PASSED;
  let reason = "";
  const missing: string[] = [];
  let layout: Result["layout"] | undefined;
  let screenshot: string | undefined;

  try {
    const response = await page.goto(url, { waitUntil: "networkidle", timeout: 15_000 });
    if (response) {
      httpStatus = response.status();
      if (httpStatus !== feature.expected_http_status) {
        status = STATUS_API_ERROR;
        reason = `expected http ${feature.expected_http_status}, got ${httpStatus}`;
      }
    }
    if (feature.expected_selectors?.length) {
      for (const selector of feature.expected_selectors) {
        const locator = page.locator(selector).first();
        const count = await locator.count();
        if (count === 0) {
          missing.push(selector);
        }
      }
      if (missing.length) {
        status = STATUS_SELECTOR_MISSING;
        reason = `missing selectors: ${missing.join(", ")}`;
      }
    }
    const dim = await page.evaluate(() => ({
      bodyWidth: document.body?.scrollWidth ?? 0,
      viewportWidth: window.innerWidth,
    }));
    layout = {
      feature_id: feature.id,
      viewport: viewport.name,
      body_width: dim.bodyWidth,
      viewport_width: dim.viewportWidth,
      has_horizontal_scroll: dim.bodyWidth > dim.viewportWidth + 1,
    };
    if (layout.has_horizontal_scroll) {
      status = STATUS_LAYOUT_COLLAPSED;
      reason = `body width ${dim.bodyWidth} > viewport ${dim.viewportWidth}`;
    }
    if (consoleErrors!.length) {
      status = STATUS_CONSOLE_ERROR;
      reason = `${consoleErrors!.length} console errors`;
    }
    screenshot = path.join("current", "screenshots", `${feature.id}-${viewport.name}.png`);
    await page.screenshot({ path: path.join(AUDIT_OUT, screenshot), fullPage: true });
  } catch (err) {
    status = STATUS_FAILED;
    reason = err instanceof Error ? err.message : String(err);
  }

  await context.close();
  results.push({
    feature_id: feature.id,
    viewport: viewport.name,
    status,
    reason,
    url_final: page.url(),
    http_status: httpStatus,
    screenshot,
    console_errors: consoleErrors!.length ? consoleErrors : undefined,
    network_errors: networkErrors!.length ? networkErrors : undefined,
    missing_selectors: missing.length ? missing : undefined,
    layout,
    duration_ms: Date.now() - start,
    recorded_at: new Date().toISOString(),
  });
  expect.soft(status, `${feature.id}@${viewport.name} reason: ${reason}`).toBe(STATUS_PASSED);
}
