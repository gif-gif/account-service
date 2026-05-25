import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountsPage } from "./AccountsPage";
import { clearAuthTokens, setAuthTokens } from "../lib/authTokens";
import { createAccountsStore } from "../store/accounts";

describe("AccountsPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
  });

  it("loads accounts and applies filters", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          accounts: [
            {
              id: "account-id",
              username: "user@example.com",
              region: "us",
              account_type: "pro",
              status: "active",
              quota_remaining: 900,
              tags: ["openai"],
            },
          ],
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const store = createAccountsStore();

    render(<AccountsPage store={store} />);

    expect(screen.getByText("Active accounts")).toBeInTheDocument();
    expect(screen.getByText("Total quota")).toBeInTheDocument();
    await screen.findByText("user@example.com");
    await userEvent.type(screen.getByLabelText("Region"), "us");
    await userEvent.type(screen.getByLabelText("Account type"), "pro");
    await userEvent.click(screen.getByRole("button", { name: "Apply filters" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(fetchMock).toHaveBeenLastCalledWith(
      "https://api.example.com/api/v1/accounts/query",
      expect.objectContaining({
        method: "POST",
        credentials: "omit",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" }),
      }),
    );
  });

  it("shows backend errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(JSON.stringify({ error: { code: "internal_error", message: "Database unavailable", request_id: "req" } }), {
          status: 500,
        }),
      ),
    );

    render(<AccountsPage store={createAccountsStore()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Database unavailable");
  });
});
