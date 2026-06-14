import { defineConfig } from "@playwright/test";

const HEADED = process.env.AUDIT_HEADED === "1";

export default defineConfig({
  testDir: "./",
  fullyParallel: false,
  timeout: 60_000,
  use: {
    headless: !HEADED,
    baseURL: process.env.AUDIT_BASE_URL ?? "http://127.0.0.1:7373",
  },
  reporter: [["list"]],
});
