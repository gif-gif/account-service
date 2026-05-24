import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIError, apiFetch } from "./api";

describe("apiFetch", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
  });

  it("prefixes VITE_API_BASE_URL and includes credentials", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await apiFetch<{ ok: boolean }>("/api/v1/admin/me");

    expect(result.ok).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/admin/me",
      expect.objectContaining({ credentials: "include" }),
    );
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
});
