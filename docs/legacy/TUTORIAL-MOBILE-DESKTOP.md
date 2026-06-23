> **Legado.** Substituído por [docs/TUTORIAL.md](../TUTORIAL.md). Mantido aqui só para referência histórica.

# Tutorial — Mobile (Tauri) + Desktop (Electron) with HarnessX

Goal: drive a Todoist-style task client from one chat session, shipping
a single Rust+Vite Tauri app for iOS / Android / macOS / Windows /
Linux **and** an Electron desktop variant pointing at the same FastAPI
backend you already built in `TUTORIAL-TODOIST.md`.

HarnessX itself does not bundle Tauri or Electron generators; this
walks through wiring those toolchains into the regular `harness chat`
loop so the same spec-driven, test-first flow ships UI bits the same
way it ships REST endpoints.

---

## Prerequisites

```bash
brew install rustup-init node
rustup-init -y
cargo install create-tauri-app
npm i -g electron-forge

harness onboarding --interactive   # pin claude/kimi, install missing tools
```

Confirm:
- backend running: `cd ~/dev/todoist-api && uvicorn app.main:app --reload`
- API reachable: `curl -s localhost:8000/healthz`

---

## Part 1 — Tauri shell (iOS/Android/desktop in one tree)

### 1.1 Scaffold the workspace

```bash
mkdir ~/dev/todoist-mobile && cd ~/dev/todoist-mobile
cargo create-tauri-app --manager npm --template vanilla-ts --tauri-version 2
cd todoist
npm install
```

### 1.2 Wire HarnessX

```bash
harness init --yes
harness use claude        # implementation
harness chat --auto-gate
```

Inside chat:

```
/drive add a Tasks list view in src/main.ts that GETs http://localhost:8000/tasks
        and renders one <li> per row. Include a vitest spec under
        src/__tests__/tasks.test.ts that mocks fetch and asserts
        the DOM updates.
```

The PEV loop:
1. spec generated under `.harness/artifacts/specs/`
2. cheap chain emits `src/__tests__/tasks.test.ts` (vitest, failing)
3. implementation chain edits `src/main.ts` until vitest green
4. `harness ci` runs the sensor gate (lint + tests)
5. conventional commit on current branch

### 1.3 Run the mobile preview

```bash
npm run tauri dev          # macOS / linux / windows
npm run tauri ios init     # one-off
npm run tauri ios dev      # opens iOS simulator
npm run tauri android dev  # opens Android emulator
```

### 1.4 Iterate with `/drive`

Inside `harness chat`:

```
/drive add a "+" floating button that POSTs {title} to /tasks and refreshes the list.
/drive add swipe-to-delete sending DELETE /tasks/{id}.
/drive show error toast when the API returns >=400.
```

Each `/drive` keeps the cycle: spec → test → impl → ci → commit. Cost
appears in `/cost` per turn and across sessions in
`harness analytics --since 24h`.

---

## Part 2 — Electron desktop variant

The same FastAPI backend feeds an Electron build that adds tray icon,
global hotkey, and native notifications — pieces Tauri can do but the
team may want to ship through Electron Forge for parity with existing
internal tooling.

### 2.1 Scaffold

```bash
mkdir ~/dev/todoist-desktop && cd ~/dev/todoist-desktop
npm init electron-app@latest . -- --template=vite-typescript
npm install
harness init --yes
harness use claude
harness chat --auto-gate
```

### 2.2 Drive the tray + hotkey feature

```
/drive add a system tray icon. Clicking opens a 400x600 BrowserWindow
        showing http://localhost:5173 (the Vite dev server). Add a vitest
        for src/main/tray.ts asserting createTray() returns a Tray with
        the expected tooltip.
```

```
/drive register Cmd+Shift+T as a global shortcut that toggles the window.
        Test under src/main/__tests__/shortcut.test.ts with electron
        mocked.
```

```
/drive show a native Notification('Task added', { body: title }) after
        every successful POST /tasks. Add a vitest covering the
        notification call.
```

### 2.3 Build a signed binary

```bash
npm run package          # dev build per platform
npm run make             # signed installer (needs APPLE_ID/APPLE_PASSWORD env)
```

For CI, point `make ship` at `electron-builder publish` so HarnessX'
release pipeline tracks Electron artefacts alongside Tauri ones.

---

## Part 3 — Shared API contracts

Keep both clients aligned with the backend by generating a typed
OpenAPI client off the running FastAPI server:

```bash
cd ~/dev/todoist-api && python -m app.openapi > openapi.json
cp openapi.json ~/dev/todoist-mobile/openapi.json
cp openapi.json ~/dev/todoist-desktop/openapi.json
```

In each client:

```
/drive regenerate the typed client under src/api/ from ./openapi.json
        using openapi-typescript-codegen. Update existing imports.
```

When the backend grows a field, re-run the export + the `/drive` command
on both clients — HarnessX runs ts + vitest gates so a missing typed
import surfaces as a CI failure rather than at runtime.

---

## Part 4 — Cross-project analytics

After a few drive cycles:

```bash
harness analytics \
  --root ~/dev/todoist-api \
  --root ~/dev/todoist-mobile \
  --root ~/dev/todoist-desktop \
  --since 168h
```

Reports per-stack (python / node / rust), per-adapter+task
(claude/implementation, kimi/cheap_review, gemini/planning), and per-day
spend so the team sees where the budget is landing across the whole
product, not just one repo.

---

## Wrap

You now ship a Tauri mobile app + Electron desktop app + FastAPI
backend from a single agentic loop. Same spec template, same sensor
gate, same router routing planning to a cheap model and implementation
to a strong one. The Todoist case from `TUTORIAL-TODOIST.md` extends
naturally; `TUTORIAL-MULTI-STACK.md` covers backends in
Go/Rails/Rust/Ruby if the team picks a different backend.
