import { describe, expect, it, vi } from "vitest";

import { apiBaseURL, appEnvironment } from "./env";

describe("env", () => {
  it("reads API base URL from VITE_API_BASE_URL", () => {
    vi.stubEnv("VITE_API_BASE_URL", "https://account.goio.uk");

    expect(apiBaseURL()).toBe("https://account.goio.uk");
  });

  it("reads app environment from VITE_APP_ENV", () => {
    vi.stubEnv("VITE_APP_ENV", "test");

    expect(appEnvironment()).toBe("test");
  });
});
