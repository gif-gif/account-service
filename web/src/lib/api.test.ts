import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIError, apiFetch } from "./api";
import { clearAuthTokens, getAuthTokens, setAuthTokens } from "./authTokens";

describe("apiFetch", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
  });

  it("prefixes VITE_API_BASE_URL and sends access token authorization", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await apiFetch<{ ok: boolean }>("/api/v1/admin/me");

    expect(result.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/admin/me",
      expect.objectContaining({
        credentials: "omit",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" }),
      }),
    );
  });

  it("refreshes tokens and retries once when access token is expired", async () => {
    setAuthTokens({ accessToken: "expired-access", refreshToken: "refresh-token" });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: { code: "unauthorized", message: "Access token is required" } }), { status: 401 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            user: { id: "admin-id", username: "admin" },
            accessToken: "fresh-access",
            refreshToken: "fresh-refresh",
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await apiFetch<{ ok: boolean }>("/api/v1/accounts/query", { method: "POST", body: JSON.stringify({ limit: 10 }) });

    expect(result.ok).toBe(true);
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "https://api.example.com/api/v1/admin/refresh",
      expect.objectContaining({ body: JSON.stringify({ refreshToken: "refresh-token" }) }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "https://api.example.com/api/v1/accounts/query",
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: "Bearer fresh-access" }),
      }),
    );
    expect(getAuthTokens()).toEqual({ accessToken: "fresh-access", refreshToken: "fresh-refresh" });
  });

  it("clears tokens when refresh token is expired", async () => {
    setAuthTokens({ accessToken: "expired-access", refreshToken: "expired-refresh" });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: { code: "unauthorized", message: "Access token is required" } }), { status: 401 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: { code: "unauthorized", message: "Refresh token is invalid or expired" } }), { status: 401 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(apiFetch("/api/v1/admin/me")).rejects.toMatchObject({ status: 401 });
    expect(getAuthTokens()).toBeNull();
  });

  it("throws normalized API errors with request id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            error: { code: "unauthorized", message: "Admin session is required", request_id: "request-id" },
          }),
          { status: 401 },
        ),
      ),
    );

    await expect(apiFetch("/api/v1/admin/me")).rejects.toMatchObject(
      new APIError("unauthorized", "Admin session is required", "request-id", 401),
    );
  });

  it("normalizes non-json error responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("Method Not Allowed", { status: 405, headers: { "X-Request-ID": "request-id" } })),
    );

    await expect(apiFetch("/api/v1/accounts/account-id", { method: "DELETE" })).rejects.toMatchObject(
      new APIError("request_failed", "Method Not Allowed", "request-id", 405),
    );
  });
});
