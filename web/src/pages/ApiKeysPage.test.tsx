import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiKeysPage } from "./ApiKeysPage";

describe("ApiKeysPage", () => {
  beforeEach(() => vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com"));

  it("lists API keys and supports create update delete", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            callers: [
              {
                id: "caller-id",
                name: "worker",
                description: "background worker",
                status: "active",
                api_key: "should-not-render-from-list",
                created_at: "2026-05-31T00:00:00Z",
                updated_at: "2026-05-31T00:00:00Z",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ api_key: "acct_worker" }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            caller: {
              id: "caller-2",
              name: "batch",
              description: "batch worker",
              status: "disabled",
              created_at: "2026-05-31T00:00:00Z",
              updated_at: "2026-05-31T00:00:00Z",
            },
            api_key: "acct_secret",
          }),
          { status: 201 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            caller: {
              id: "caller-2",
              name: "batch-renamed",
              description: "batch worker",
              status: "active",
              created_at: "2026-05-31T00:00:00Z",
              updated_at: "2026-05-31T00:01:00Z",
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<ApiKeysPage />);

    expect(await screen.findByText("worker")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "API Key" })).toBeInTheDocument();
    expect(screen.getByText("••••••••")).toBeInTheDocument();
    expect(screen.queryByText("should-not-render-from-list")).not.toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "状态" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "创建时间" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "更新时间" })).toBeInTheDocument();
    expect(screen.getByText("可用")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "查看 API Key worker" }));
    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/api-keys/caller-id/secret", expect.objectContaining({ method: "GET" })),
    );
    const secretDialog = screen.getByRole("dialog", { name: "API Key" });
    expect(secretDialog).toHaveTextContent("acct_worker");
    await userEvent.click(within(secretDialog).getByRole("button", { name: "关闭" }));

    await userEvent.click(screen.getByRole("button", { name: "创建 API Key" }));
    const createDialog = screen.getByRole("dialog", { name: "创建 API Key" });
    await userEvent.type(within(createDialog).getByLabelText("名称"), "batch");
    await userEvent.type(within(createDialog).getByLabelText("描述"), "batch worker");
    await userEvent.selectOptions(within(createDialog).getByLabelText("状态"), "disabled");
    await userEvent.click(within(createDialog).getByRole("button", { name: "创建 API Key" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        "https://api.example.com/api/v1/api-keys",
        expect.objectContaining({ method: "POST", body: expect.stringContaining("\"status\":\"disabled\"") }),
      ),
    );
    expect(await screen.findByText("acct_secret")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "关闭 API Key" }));

    await userEvent.click(screen.getByRole("button", { name: "编辑 batch" }));
    const editDialog = screen.getByRole("dialog", { name: "编辑 API Key" });
    await userEvent.clear(within(editDialog).getByLabelText("名称"));
    await userEvent.type(within(editDialog).getByLabelText("名称"), "batch-renamed");
    await userEvent.selectOptions(within(editDialog).getByLabelText("状态"), "active");
    await userEvent.click(within(editDialog).getByRole("button", { name: "保存 API Key" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        "https://api.example.com/api/v1/api-keys/caller-2",
        expect.objectContaining({ method: "PATCH", body: expect.stringContaining("\"status\":\"active\"") }),
      ),
    );
    expect(await screen.findByText("batch-renamed")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "删除 batch-renamed" }));
    const deleteDialog = screen.getByRole("dialog", { name: "删除 API Key" });
    expect(deleteDialog).toHaveTextContent("batch-renamed");
    await userEvent.click(within(deleteDialog).getByRole("button", { name: "确认删除" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/api-keys/caller-2", expect.objectContaining({ method: "DELETE" })),
    );
  });
});
