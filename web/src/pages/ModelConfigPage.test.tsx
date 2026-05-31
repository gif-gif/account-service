import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ModelConfigPage } from "./ModelConfigPage";

describe("ModelConfigPage", () => {
  beforeEach(() => vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com"));

  it("loads model config items and supports create update delete", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            items: [
              {
                id: "item-1",
                kind: "fallback_model",
                key: "auto",
                value: "",
                display_order: 10,
                created_at: "2026-05-31T00:00:00Z",
                updated_at: "2026-05-31T00:00:00Z",
              },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            item: {
              id: "item-2",
              kind: "model_alias",
              key: "claude-opus-4-7",
              value: "claude-opus-4.7",
              display_order: 20,
              created_at: "2026-05-31T00:00:00Z",
              updated_at: "2026-05-31T00:00:00Z",
            },
          }),
          { status: 201 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            item: {
              id: "item-2",
              kind: "model_alias",
              key: "claude-opus-4-7",
              value: "claude-opus-4.8",
              display_order: 20,
              created_at: "2026-05-31T00:00:00Z",
              updated_at: "2026-05-31T00:01:00Z",
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<ModelConfigPage />);

    expect(await screen.findByText("auto")).toBeInTheDocument();
    expect(screen.queryByRole("dialog", { name: "新增模型配置" })).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "创建配置" }));
    const createDialog = screen.getByRole("dialog", { name: "新增模型配置" });
    await userEvent.selectOptions(within(createDialog).getByLabelText("类型"), "model_alias");
    await userEvent.type(within(createDialog).getByLabelText("键"), "claude-opus-4-7");
    await userEvent.type(within(createDialog).getByLabelText("值"), "claude-opus-4.7");
    await userEvent.clear(within(createDialog).getByLabelText("排序"));
    await userEvent.type(within(createDialog).getByLabelText("排序"), "20");
    await userEvent.click(within(createDialog).getByRole("button", { name: "创建配置" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        "https://api.example.com/api/v1/model-config/items",
        expect.objectContaining({ method: "POST", body: expect.stringContaining("claude-opus-4-7") }),
      ),
    );
    expect(await screen.findByText("claude-opus-4.7")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "新增模型配置" })).not.toBeInTheDocument());

    await userEvent.click(screen.getByRole("button", { name: "编辑 claude-opus-4-7" }));
    const editDialog = screen.getByRole("dialog", { name: "编辑模型配置" });
    await userEvent.clear(within(editDialog).getByLabelText("值"));
    await userEvent.type(within(editDialog).getByLabelText("值"), "claude-opus-4.8");
    await userEvent.click(within(editDialog).getByRole("button", { name: "保存配置" }));

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        "https://api.example.com/api/v1/model-config/items/item-2",
        expect.objectContaining({ method: "PATCH", body: expect.stringContaining("claude-opus-4.8") }),
      ),
    );
    expect(await screen.findByText("claude-opus-4.8")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "删除 claude-opus-4-7" }));
    const deleteDialog = screen.getByRole("dialog", { name: "删除模型配置" });
    expect(deleteDialog).toHaveTextContent("claude-opus-4-7");
    await userEvent.click(within(deleteDialog).getByRole("button", { name: "确认删除" }));
    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith("https://api.example.com/api/v1/model-config/items/item-2", expect.objectContaining({ method: "DELETE" })),
    );
  });
});
