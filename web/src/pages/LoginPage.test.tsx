import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LoginPage } from "./LoginPage";
import { createAuthStore } from "../store/auth";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
  });

  it("submits admin credentials", async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ user: { id: "admin-id", username: "admin" } }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    const store = createAuthStore();

    render(<LoginPage store={store} />);

    await userEvent.type(screen.getByLabelText("Username"), "admin");
    await userEvent.type(screen.getByLabelText("Password"), "password123");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/admin/login",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
    expect(store.getState().user?.username).toBe("admin");
  });
});
