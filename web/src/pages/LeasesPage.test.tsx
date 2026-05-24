import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LeasesPage } from "./LeasesPage";

describe("LeasesPage", () => {
  beforeEach(() => vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com"));

  it("loads leases with status filter", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ leases: [{ lease_id: "lease-id", account_id: "account-id", status: "active" }] }), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    render(<LeasesPage />);

    await screen.findByText("lease-id");
    await userEvent.type(screen.getByLabelText("Lease status"), "released");
    await userEvent.click(screen.getByRole("button", { name: "Filter leases" }));

    expect(fetchMock).toHaveBeenLastCalledWith(
      "https://api.example.com/api/v1/leases?status=released",
      expect.objectContaining({ credentials: "include" }),
    );
  });
});
