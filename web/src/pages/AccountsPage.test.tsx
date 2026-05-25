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
    localStorage.clear();
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
              password: "plain-password",
              region: "us",
              account_type: "pro",
              status: "active",
              login_url: "https://example.com/login",
              access_token: "provider-access",
              refresh_token: "provider-refresh",
              quota_total: 1000,
              quota_used: 100,
              quota_remaining: 900,
              max_concurrent_leases: 2,
              tags: ["openai"],
              notes: "primary",
              created_at: "2026-05-25T08:00:00Z",
              updated_at: "2026-05-25T09:00:00Z",
            },
          ],
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const store = createAccountsStore();

    render(<AccountsPage store={store} />);

    expect(screen.getByRole("heading", { name: "账号管理" })).toBeInTheDocument();
    expect(screen.getByText("活跃账号")).toBeInTheDocument();
    expect(screen.getAllByText("剩余额度").length).toBeGreaterThan(0);
    await screen.findByText("user@example.com");
    const filters = screen.getByRole("group", { name: "筛选条件" });
    expect(within(filters).getByText("筛选条件")).toBeInTheDocument();
    await userEvent.type(within(filters).getByLabelText("区域"), "us");
    await userEvent.type(within(filters).getByLabelText("类型"), "pro");
    await userEvent.selectOptions(within(filters).getByLabelText("状态"), "active");
    await userEvent.click(screen.getByRole("button", { name: "筛选" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(fetchMock).toHaveBeenLastCalledWith(
      "https://api.example.com/api/v1/accounts/query",
      expect.objectContaining({
        method: "POST",
        credentials: "omit",
        headers: expect.objectContaining({ Authorization: "Bearer access-token" }),
        body: expect.stringContaining('"statuses":["active"]'),
      }),
    );
    expect(screen.getByRole("columnheader", { name: "ID" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "登录地址" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Access Token" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Refresh Token" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "总额度" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "已用额度" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "最大租约数" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "创建时间" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "更新时间" })).toBeInTheDocument();
    expect(screen.queryByText("provider-access")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "查看 Access Token user@example.com" }));
    expect(screen.getByRole("dialog", { name: "Access Token" })).toHaveTextContent("provider-access");
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
                password: "plain-password",
                login_url: "https://example.com/login",
                access_token: "provider-access",
                refresh_token: "provider-refresh",
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
                password: "plain-password",
                login_url: "https://example.com/login",
                access_token: "provider-access",
                refresh_token: "provider-refresh",
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
                password: "plain-password",
                login_url: "https://example.com/login",
                access_token: "provider-access",
                refresh_token: "provider-refresh",
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
    expect(screen.queryByRole("heading", { name: "账号详情" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "查看 user@example.com" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "编辑 user@example.com" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "删除 user@example.com" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "添加账号" }));
    const createDialog = screen.getByRole("dialog", { name: "添加账号" });
    expect(createDialog).toBeInTheDocument();
    await userEvent.type(within(createDialog).getByLabelText("用户名"), "new@example.com");
    await userEvent.type(within(createDialog).getByLabelText("密码"), "plain-password");
    await userEvent.type(within(createDialog).getByLabelText("登录地址"), "https://example.com/login");
    await userEvent.type(within(createDialog).getByLabelText("Access Token"), "provider-access");
    await userEvent.type(within(createDialog).getByLabelText("Refresh Token"), "provider-refresh");
    await userEvent.type(within(createDialog).getByLabelText("区域"), "eu");
    await userEvent.type(within(createDialog).getByLabelText("账号类型"), "team");
    await userEvent.type(within(createDialog).getByLabelText("剩余额度"), "500");
    await userEvent.click(within(createDialog).getByRole("button", { name: "创建" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts", expect.objectContaining({ method: "POST" })));

    await userEvent.click(screen.getByRole("button", { name: "查看 user@example.com" }));
    expect(screen.getByRole("dialog", { name: "账号详情" })).toHaveTextContent("primary");
    await userEvent.click(screen.getByRole("button", { name: "关闭" }));

    await userEvent.click(screen.getByRole("button", { name: "编辑 user@example.com" }));
    const editDialog = screen.getByRole("dialog", { name: "编辑账号" });
    expect(editDialog).toBeInTheDocument();
    await userEvent.clear(within(editDialog).getByLabelText("剩余额度"));
    await userEvent.type(within(editDialog).getByLabelText("剩余额度"), "700");
    await userEvent.click(within(editDialog).getByRole("button", { name: "保存" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts/account-id", expect.objectContaining({ method: "PATCH" })),
    );

    await userEvent.click(screen.getByRole("button", { name: "删除 user@example.com" }));
    expect(screen.getByRole("dialog", { name: "删除账号" })).toHaveTextContent("user@example.com");
    await userEvent.click(screen.getByRole("button", { name: "确认删除" }));

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
