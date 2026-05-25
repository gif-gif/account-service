import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LeasesPage } from "./LeasesPage";
import { clearAuthTokens, setAuthTokens } from "../lib/authTokens";

describe("LeasesPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
  });

  it("loads leases with status filter", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ leases: [{ lease_id: "lease-id", account_id: "account-id", status: "active" }] }), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    render(<LeasesPage />);

    await screen.findByText("lease-id");
    expect(screen.getByRole("table")).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("Lease status"), "released");
    await userEvent.click(screen.getByRole("button", { name: "Filter leases" }));

    expect(fetchMock).toHaveBeenLastCalledWith(
      "https://api.example.com/api/v1/leases?status=released",
      expect.objectContaining({
        credentials: "omit",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" }),
      }),
    );
  });
});
