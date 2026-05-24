# Web Admin Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the React admin frontend into a shadcn-style operations console with a standalone login screen.

**Architecture:** Keep the backend API contract unchanged and isolate this work to `web`. Add local shadcn-style primitives, a CSS token system, and an authenticated `AdminShell` that owns navigation between existing feature pages.

**Tech Stack:** React, Vite, Zustand, Testing Library, Vitest, local shadcn-style components, plain CSS tokens.

---

### Task 1: Auth Gate And Shell

**Files:**
- Create: `web/src/App.test.tsx`
- Create: `web/src/components/AdminShell.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/main.tsx`

- [ ] **Step 1: Write failing tests**

```tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { App } from "./App";
import { createAuthStore } from "./store/auth";

describe("App", () => {
  it("shows only the standalone login page when signed out", () => {
    const store = createAuthStore();

    render(<App authStore={store} />);

    expect(screen.getByRole("form", { name: "Admin login" })).toBeInTheDocument();
    expect(screen.queryByRole("navigation", { name: "Admin sections" })).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Accounts" })).not.toBeInTheDocument();
  });

  it("shows the operations shell and switches sections when signed in", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ accounts: [], leases: [], audit_logs: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    const store = createAuthStore();
    store.setState({ user: { id: "admin-id", username: "admin" } });

    render(<App authStore={store} />);

    expect(screen.getByRole("navigation", { name: "Admin sections" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Operations overview" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "API Keys" }));

    expect(screen.getByRole("heading", { name: "API Keys" })).toBeInTheDocument();
    expect(screen.queryByRole("form", { name: "Admin login" })).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npm test -- --run src/App.test.tsx`
Expected: FAIL because `App` does not accept `authStore` and all pages render at once.

- [ ] **Step 3: Implement the minimal shell**

Create `AdminShell` with local navigation state for Overview, Accounts, Leases, API Keys, and Audit Logs. Update `App` so signed-out users see only `LoginPage`; signed-in users see only `AdminShell`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npm test -- --run src/App.test.tsx`
Expected: PASS.

### Task 2: shadcn-Style Component Primitives And Theme

**Files:**
- Create: `web/src/styles.css`
- Create: `web/src/lib/cn.ts`
- Create: `web/src/components/ui/button.tsx`
- Create: `web/src/components/ui/card.tsx`
- Create: `web/src/components/ui/input.tsx`
- Create: `web/src/components/ui/label.tsx`
- Create: `web/src/components/ui/badge.tsx`
- Create: `web/src/components/ui/alert.tsx`
- Create: `web/src/components/ui/table.tsx`
- Modify: `web/src/main.tsx`

- [ ] **Step 1: Write failing component tests**

Add focused tests through existing page tests after components are wired. Do not add snapshot tests; verify accessible names and roles.

- [ ] **Step 2: Implement primitives**

Use forwardRef-compatible React components with shadcn-style class names and variants. Keep dependencies local; do not require network installs for styling.

- [ ] **Step 3: Add global CSS tokens**

Define background, foreground, border, muted, primary, destructive, success, warning, radius, shell layout, table, form, card, badge, and responsive rules. Import the CSS once from `main.tsx`.

- [ ] **Step 4: Run existing component tests**

Run: `cd web && npm test -- --run src/components`
Expected: PASS.

### Task 3: Standalone Login Page

**Files:**
- Modify: `web/src/pages/LoginPage.tsx`
- Modify: `web/src/pages/LoginPage.test.tsx`

- [ ] **Step 1: Extend tests**

Assert the login page has the `Account Admin` heading, username/password labels, alert behavior, and no admin navigation.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npm test -- --run src/pages/LoginPage.test.tsx`
Expected: FAIL before style structure is added.

- [ ] **Step 3: Restyle login page**

Use `Card`, `Input`, `Label`, `Button`, and `Alert`. Keep the login page visually independent from the admin shell.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npm test -- --run src/pages/LoginPage.test.tsx`
Expected: PASS.

### Task 4: Operations Pages

**Files:**
- Modify: `web/src/pages/AccountsPage.tsx`
- Modify: `web/src/pages/AccountDetailPage.tsx`
- Modify: `web/src/pages/LeasesPage.tsx`
- Modify: `web/src/pages/ApiKeysPage.tsx`
- Modify: `web/src/pages/AuditLogsPage.tsx`
- Modify: matching page tests

- [ ] **Step 1: Extend page tests**

Verify Accounts has metrics and filter controls, Leases uses a table with status badges, API Keys keeps the one-time secret behavior, and Audit Logs redacts sensitive metadata.

- [ ] **Step 2: Run tests to verify failures**

Run: `cd web && npm test -- --run src/pages`
Expected: FAIL for new visual structure assertions.

- [ ] **Step 3: Restyle pages**

Use shadcn-style components while preserving current API calls and accessible labels. Add metric cards and tables; keep mobile fallback with CSS.

- [ ] **Step 4: Run tests to verify pass**

Run: `cd web && npm test -- --run src/pages`
Expected: PASS.

### Task 5: Full Verification

**Files:**
- No new files.

- [ ] **Step 1: Run all frontend tests**

Run: `cd web && npm test -- --run`
Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run: `cd web && npm run build`
Expected: PASS.

- [ ] **Step 3: Inspect git diff**

Run: `git status --short` and `git diff --stat`
Expected: only frontend redesign files, spec, and plan are changed.
