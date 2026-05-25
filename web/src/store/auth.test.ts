import { beforeEach, describe, expect, it, vi } from "vitest";

import { getAuthTokens } from "../lib/authTokens";
import { createAuthStore } from "./auth";

describe("auth store", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    localStorage.clear();
  });

  it("logs in, restores current user, and logs out", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            user: { id: "admin-id", username: "admin" },
            accessToken: "access-token",
            refreshToken: "refresh-token",
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ user: { id: "admin-id", username: "admin" } }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    const store = createAuthStore();

    await store.getState().login("admin", "password123");
    expect(store.getState().user?.username).toBe("admin");
    expect(getAuthTokens()).toEqual({ accessToken: "access-token", refreshToken: "refresh-token" });

    await store.getState().restore();
    expect(store.getState().user?.id).toBe("admin-id");

    await store.getState().logout();
    expect(store.getState().user).toBeNull();
    expect(getAuthTokens()).toBeNull();
  });

  it("stores failed login message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(JSON.stringify({ error: { code: "unauthorized", message: "Invalid username or password" } }), {
          status: 401,
        }),
      ),
    );
    const store = createAuthStore();

    await expect(store.getState().login("admin", "wrong")).rejects.toThrow("Invalid username or password");
    expect(store.getState().error).toBe("Invalid username or password");
  });
});
