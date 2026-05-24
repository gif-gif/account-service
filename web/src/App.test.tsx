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
