import { render, screen, waitFor, within } from "@testing-library/react";
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
    await userEvent.type(screen.getByLabelText("Type"), "pro");
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

  it("opens create, view, edit, and delete account dialogs from the table", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            accounts: [
              {
                id: "account-id",
                username: "user@example.com",
                login_url: "https://example.com/login",
                region: "us",
                account_type: "pro",
                status: "active",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["openai"],
                notes: "primary",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ account: { id: "new-account-id" } }), { status: 201 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            accounts: [
              {
                id: "account-id",
                username: "user@example.com",
                login_url: "https://example.com/login",
                region: "us",
                account_type: "pro",
                status: "active",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["openai"],
                notes: "primary",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ account: { id: "account-id" } }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            accounts: [
              {
                id: "account-id",
                username: "user@example.com",
                login_url: "https://example.com/login",
                region: "us",
                account_type: "pro",
                status: "active",
                quota_remaining: 700,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["openai"],
                notes: "primary",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ accounts: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountsPage store={createAccountsStore()} />);

    await screen.findByText("user@example.com");
    expect(screen.queryByRole("heading", { name: "Account detail" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "View user@example.com" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit user@example.com" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete user@example.com" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "New account" }));
    const createDialog = screen.getByRole("dialog", { name: "New account" });
    expect(createDialog).toBeInTheDocument();
    await userEvent.type(within(createDialog).getByLabelText("Username"), "new@example.com");
    await userEvent.type(within(createDialog).getByLabelText("Password"), "plain-password");
    await userEvent.type(within(createDialog).getByLabelText("Login URL"), "https://example.com/login");
    await userEvent.type(within(createDialog).getByLabelText("Provider access token"), "provider-access");
    await userEvent.type(within(createDialog).getByLabelText("Provider refresh token"), "provider-refresh");
    await userEvent.type(within(createDialog).getByLabelText("Region"), "eu");
    await userEvent.type(within(createDialog).getByLabelText("Account type"), "team");
    await userEvent.type(within(createDialog).getByLabelText("Quota remaining"), "500");
    await userEvent.click(within(createDialog).getByRole("button", { name: "Create account" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts", expect.objectContaining({ method: "POST" })));

    await userEvent.click(screen.getByRole("button", { name: "View user@example.com" }));
    expect(screen.getByRole("dialog", { name: "Account details" })).toHaveTextContent("primary");
    await userEvent.click(screen.getByRole("button", { name: "Close" }));

    await userEvent.click(screen.getByRole("button", { name: "Edit user@example.com" }));
    const editDialog = screen.getByRole("dialog", { name: "Edit account" });
    expect(editDialog).toBeInTheDocument();
    await userEvent.clear(within(editDialog).getByLabelText("Quota remaining"));
    await userEvent.type(within(editDialog).getByLabelText("Quota remaining"), "700");
    await userEvent.click(within(editDialog).getByRole("button", { name: "Save changes" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts/account-id", expect.objectContaining({ method: "PATCH" })),
    );

    await userEvent.click(screen.getByRole("button", { name: "Delete user@example.com" }));
    expect(screen.getByRole("dialog", { name: "Delete account" })).toHaveTextContent("user@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Delete account" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts/account-id", expect.objectContaining({ method: "DELETE" })),
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
