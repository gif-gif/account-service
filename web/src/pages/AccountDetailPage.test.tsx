import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountDetailPage } from "./AccountDetailPage";
import { clearAuthTokens, setAuthTokens } from "../lib/authTokens";

describe("AccountDetailPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
  });

  async function selectDropdownOption(control: HTMLElement, option: string) {
    await userEvent.click(control);
    await userEvent.click(await screen.findByRole("option", { name: option }));
  }

  it("creates account from form fields", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ account: { id: "account-id", username: "user@example.com" } }), { status: 201 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountDetailPage />);

    await userEvent.type(screen.getByLabelText("Username"), "user@example.com");
    await userEvent.type(screen.getByLabelText("Password"), "plain-password");
    await userEvent.type(screen.getByLabelText("Login URL"), "https://example.com/login");
    await userEvent.type(screen.getByLabelText("Access token"), "access-token");
    await userEvent.type(screen.getByLabelText("Refresh token"), "refresh-token");
    await userEvent.type(screen.getByLabelText("Region"), "us");
    expect(screen.getByLabelText("Account type")).toHaveAttribute("data-slot", "select-trigger");
    await selectDropdownOption(screen.getByLabelText("Account type"), "codex");
    await userEvent.type(screen.getByLabelText("Quota remaining"), "900");
    await userEvent.click(screen.getByRole("button", { name: "Save account" }));

    expect(await screen.findByText("Saved account-id")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts",
      expect.objectContaining({
        method: "POST",
        credentials: "omit",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" }),
        body: expect.stringContaining('"account_type":"codex"'),
      }),
    );
  });
});
