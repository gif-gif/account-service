import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { AccountsPage } from "./AccountsPage";
import { clearAuthTokens, setAuthTokens } from "../lib/authTokens";
import { createAccountsStore } from "../store/accounts";

describe("AccountsPage", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com");
    clearAuthTokens();
    localStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  async function selectDropdownOption(control: HTMLElement, option: string) {
    await userEvent.click(control);
    await userEvent.click(await screen.findByRole("option", { name: option }));
  }

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
              account_type: "codex",
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
    expect(within(filters).getByLabelText("类型")).toHaveAttribute("data-slot", "select-trigger");
    await selectDropdownOption(within(filters).getByLabelText("类型"), "codex");
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
    expect(screen.queryByText("account-id")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "查看 ID user@example.com" }));
    expect(screen.getByRole("dialog", { name: "ID" })).toHaveTextContent("account-id");
    await userEvent.click(screen.getByRole("button", { name: "关闭" }));
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
                account_type: "codex",
                status: "active",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["openai"],
                notes: "primary",
                created_at: "2026-05-25T08:00:00Z",
                updated_at: "2026-05-25T09:00:00Z",
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
                account_type: "codex",
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
                account_type: "codex",
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
    expect(within(createDialog).getByLabelText("账号类型")).toHaveAttribute("data-slot", "select-trigger");
    await selectDropdownOption(within(createDialog).getByLabelText("账号类型"), "kiro-aws");
    await userEvent.selectOptions(within(createDialog).getByLabelText("状态"), "login_failed");
    await userEvent.type(within(createDialog).getByLabelText("剩余额度"), "500");
    await userEvent.click(within(createDialog).getByRole("button", { name: "创建" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts", expect.objectContaining({ method: "POST" })));
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts",
      expect.objectContaining({
        body: expect.stringContaining('"status":"login_failed"'),
        method: "POST",
      }),
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts",
      expect.objectContaining({
        body: expect.stringContaining('"account_type":"kiro-aws"'),
        method: "POST",
      }),
    );

    await userEvent.click(screen.getByRole("button", { name: "查看 user@example.com" }));
    const detailsDialog = screen.getByRole("dialog", { name: "账号详情" });
    [
      "ID",
      "用户名",
      "密码",
      "登录地址",
      "Access Token",
      "Refresh Token",
      "区域",
      "账号类型",
      "状态",
      "总额度",
      "已用额度",
      "剩余额度",
      "最大租约数",
      "标签",
      "备注",
      "创建时间",
      "更新时间",
    ].forEach((label) => expect(within(detailsDialog).getByText(label)).toBeInTheDocument());
    expect(detailsDialog).toHaveTextContent("account-id");
    expect(detailsDialog).toHaveTextContent("user@example.com");
    expect(detailsDialog).toHaveTextContent("https://example.com/login");
    expect(detailsDialog).toHaveTextContent("primary");
    expect(detailsDialog).toHaveTextContent("openai");
    expect(detailsDialog).not.toHaveTextContent("provider-access");
    await userEvent.click(within(detailsDialog).getByRole("button", { name: "查看 Access Token user@example.com" }));
    const secretDialog = screen.getByRole("dialog", { name: "Access Token" });
    expect(secretDialog).toHaveTextContent("provider-access");
    expect(screen.getByRole("dialog", { name: "账号详情" })).toBeInTheDocument();
    await userEvent.click(within(secretDialog).getByRole("button", { name: "关闭" }));
    expect(screen.getByRole("dialog", { name: "账号详情" })).toBeInTheDocument();
    await userEvent.click(within(screen.getByRole("dialog", { name: "账号详情" })).getByRole("button", { name: "关闭" }));

    await userEvent.click(screen.getByRole("button", { name: "编辑 user@example.com" }));
    const editDialog = screen.getByRole("dialog", { name: "编辑账号" });
    expect(editDialog).toBeInTheDocument();
    expect(within(editDialog).getByLabelText("账号类型")).toHaveAttribute("data-slot", "select-trigger");
    expect(within(editDialog).getByLabelText("状态")).toHaveValue("active");
    await userEvent.selectOptions(within(editDialog).getByLabelText("状态"), "disabled");
    await userEvent.clear(within(editDialog).getByLabelText("剩余额度"));
    await userEvent.type(within(editDialog).getByLabelText("剩余额度"), "700");
    await userEvent.click(within(editDialog).getByRole("button", { name: "保存" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts/account-id", expect.objectContaining({ method: "PATCH" })),
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id",
      expect.objectContaining({
        body: expect.stringContaining('"status":"disabled"'),
        method: "PATCH",
      }),
    );

    await userEvent.click(screen.getByRole("button", { name: "删除 user@example.com" }));
    expect(screen.getByRole("dialog", { name: "删除账号" })).toHaveTextContent("user@example.com");
    await userEvent.click(screen.getByRole("button", { name: "确认删除" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/accounts/account-id", expect.objectContaining({ method: "DELETE" })),
    );
  });

  it("shows login only for kiro account types and requires confirmation", async () => {
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
                account_type: "kiro-aws",
                status: "disabled",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["kiro"],
                notes: "primary",
              },
              {
                id: "codex-account-id",
                username: "codex@example.com",
                password: "plain-password",
                login_url: "https://example.com/login",
                access_token: "provider-access",
                refresh_token: "provider-refresh",
                region: "us",
                account_type: "codex",
                status: "disabled",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["codex"],
                notes: "secondary",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ account_id: "account-id", status: "running" }), { status: 202 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ running: true, target_url: "" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountsPage store={createAccountsStore()} />);

    await screen.findByText("user@example.com");
    await screen.findByText("codex@example.com");
    expect(screen.getByRole("button", { name: "账号登录 user@example.com" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "账号登录 codex@example.com" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "账号登录 user@example.com" }));
    expect(fetchMock).not.toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/kiroLogin",
      expect.objectContaining({ method: "POST" }),
    );
    const confirmDialog = screen.getByRole("dialog", { name: "确认账号登录" });
    expect(confirmDialog).toHaveTextContent("user@example.com");
    expect(confirmDialog).toHaveTextContent("用户名");
    expect(confirmDialog).toHaveTextContent("备注");
    expect(confirmDialog).toHaveTextContent("primary");
    await userEvent.click(within(confirmDialog).getByRole("button", { name: "确认登录" }));

    const loginDialog = screen.getByRole("dialog", { name: "账号登录" });
    expect(loginDialog).toHaveTextContent("user@example.com");
    expect(within(loginDialog).queryByText("Close dialog")).not.toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/kiroLogin",
      expect.objectContaining({ method: "POST" }),
    );

    await userEvent.click(within(loginDialog).getByRole("button", { name: "取消" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        "https://api.example.com/api/v1/accounts/account-id/cancelKiroLogin",
        expect.objectContaining({ method: "POST" }),
      ),
    );
    expect(screen.queryByRole("dialog", { name: "账号登录" })).not.toBeInTheDocument();
  });

  it("auto cancels kiro login after two minutes", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = input.toString();
      if (url.endsWith("/kiroLogin")) {
        return new Response(JSON.stringify({ account_id: "account-id", status: "running" }), { status: 202 });
      }
      if (url.endsWith("/kiroLogin/targetUrl")) {
        return new Response(JSON.stringify({ running: true, target_url: "" }), { status: 200 });
      }
      if (url.endsWith("/cancelKiroLogin")) {
        return new Response(JSON.stringify({ ok: true }), { status: 200 });
      }
      return new Response(
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
              account_type: "kiro-aws",
              status: "disabled",
              quota_remaining: 900,
              quota_total: 1000,
              quota_used: 100,
              max_concurrent_leases: 1,
              tags: ["kiro"],
              notes: "primary",
            },
          ],
        }),
        { status: 200 },
      );
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountsPage store={createAccountsStore()} />);

    await screen.findByText("user@example.com");
    vi.useFakeTimers();
    fireEvent.click(screen.getByRole("button", { name: "账号登录 user@example.com" }));
    fireEvent.click(screen.getByRole("button", { name: "确认登录" }));
    expect(screen.getByRole("dialog", { name: "账号登录" })).toBeInTheDocument();

    await vi.advanceTimersByTimeAsync(120_000);

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/cancelKiroLogin",
      expect.objectContaining({ method: "POST" }),
    );
    vi.useRealTimers();
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "账号登录" })).not.toBeInTheDocument());
  });

  it("polls kiro login target url, shows copy action, and closes the dialog when it stops", async () => {
    setAuthTokens({ accessToken: "access-token", refreshToken: "refresh-token" });
    const writeText = vi.fn(async () => undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });
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
                account_type: "kiro-offical",
                status: "disabled",
                quota_remaining: 900,
                quota_total: 1000,
                quota_used: 100,
                max_concurrent_leases: 1,
                tags: ["kiro"],
                notes: "primary",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ account_id: "account-id", status: "running" }), { status: 202 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            running: true,
            target_url: "https://d-90660ed825.awsapps.com/start/#/device?user_code=MPPG-MKGV",
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ running: false, target_url: "" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ accounts: [] }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<AccountsPage store={createAccountsStore()} />);

    await screen.findByText("user@example.com");
    vi.useFakeTimers();
    fireEvent.click(screen.getByRole("button", { name: "账号登录 user@example.com" }));
    fireEvent.click(screen.getByRole("button", { name: "确认登录" }));
    expect(screen.getByRole("dialog", { name: "账号登录" })).toBeInTheDocument();
    await vi.advanceTimersByTimeAsync(0);
    await Promise.resolve();
    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/kiroLogin",
      expect.objectContaining({ method: "POST" }),
    );
    await vi.advanceTimersByTimeAsync(0);
    await Promise.resolve();

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/kiroLogin/targetUrl",
      expect.objectContaining({ method: "GET" }),
    );
    expect(screen.getByRole("dialog", { name: "账号登录" })).toHaveTextContent("https://d-90660ed825.awsapps.com/start/#/device?user_code=MPPG-MKGV");
    fireEvent.click(screen.getByRole("button", { name: "复制 Kiro 登录链接" }));
    expect(writeText).toHaveBeenCalledWith("https://d-90660ed825.awsapps.com/start/#/device?user_code=MPPG-MKGV");

    await vi.advanceTimersByTimeAsync(5_000);
    await vi.advanceTimersByTimeAsync(0);
    await Promise.resolve();
    expect(screen.queryByRole("dialog", { name: "账号登录" })).not.toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalledWith(
      "https://api.example.com/api/v1/accounts/account-id/cancelKiroLogin",
      expect.objectContaining({ method: "POST" }),
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
