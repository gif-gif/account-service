import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";
import { clearAuthTokens, setAuthTokens } from "./lib/authTokens";
import { createAuthStore } from "./store/auth";

describe("App", () => {
  beforeEach(() => {
    clearAuthTokens();
    localStorage.clear();
  });

  it("shows only the standalone login page when signed out", () => {
    const store = createAuthStore();

    render(<App authStore={store} />);

    expect(screen.getByRole("form", { name: "管理员登录" })).toBeInTheDocument();
    expect(screen.queryByRole("navigation", { name: "后台导航" })).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "账号管理" })).not.toBeInTheDocument();
  });

  it("shows the operations shell and switches sections when signed in", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ accounts: [], leases: [], audit_logs: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    const store = createAuthStore();
    store.setState({ user: { id: "admin-id", username: "admin" } });

    render(<App authStore={store} />);

    expect(screen.getByRole("navigation", { name: "后台导航" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "运营概览" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Key 管理" }));

    expect(screen.getByRole("heading", { name: "Key 管理" })).toBeInTheDocument();
    expect(screen.queryByRole("form", { name: "管理员登录" })).not.toBeInTheDocument();
  });

  it("switches the interface language between Chinese and English", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ accounts: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    const store = createAuthStore();
    store.setState({ user: { id: "admin-id", username: "admin" } });

    render(<App authStore={store} />);

    expect(screen.getByRole("heading", { name: "运营概览" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Language: 中文" }));

    expect(screen.getByRole("heading", { name: "Operations overview" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Language: English" })).toBeInTheDocument();
  });

  it("restores the signed-in user from saved tokens on startup", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ user: { id: "admin-id", username: "admin" } }), { status: 200 })));
    const store = createAuthStore();

    render(<App authStore={store} />);

    expect(await screen.findByRole("navigation", { name: "后台导航" })).toBeInTheDocument();
  });
});
