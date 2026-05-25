import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LoginPage } from "./LoginPage";
import { clearAuthTokens } from "../lib/authTokens";
import { createAuthStore } from "../store/auth";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
  });

  it("submits admin credentials", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ user: { id: "admin-id", username: "admin" }, accessToken: "access-token", refreshToken: "refresh-token" }), {
        status: 200,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const store = createAuthStore();

    render(<LoginPage store={store} />);

    await userEvent.type(screen.getByLabelText("用户名"), "admin");
    await userEvent.type(screen.getByLabelText("密码"), "password123");
    await userEvent.click(screen.getByRole("button", { name: "登录" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/admin/login",
      expect.objectContaining({
        method: "POST",
        credentials: "omit",
      }),
    );
    expect(store.getState().user?.username).toBe("admin");
  });

  it("renders as a standalone login surface", () => {
    render(<LoginPage store={createAuthStore()} />);

    const form = screen.getByRole("form", { name: "管理员登录" });
    expect(screen.getByRole("heading", { name: "账号服务后台" })).toBeInTheDocument();
    expect(screen.getByLabelText("用户名")).toBeInTheDocument();
    expect(screen.getByLabelText("密码")).toBeInTheDocument();
    expect(form.closest(".login-card")).toBeInTheDocument();
    expect(screen.queryByRole("navigation", { name: "后台导航" })).not.toBeInTheDocument();
  });
});
