import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiKeysPage } from "./ApiKeysPage";

describe("ApiKeysPage", () => {
  beforeEach(() => vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com"));

  it("creates and displays API key once", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(JSON.stringify({ caller: { id: "caller-id", name: "worker", status: "active" }, api_key: "acct_secret" }), {
        status: 201,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    render(<ApiKeysPage />);

    expect(screen.getByText("调用凭证")).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("名称"), "worker");
    await userEvent.type(screen.getByLabelText("描述"), "background worker");
    await userEvent.click(screen.getByRole("button", { name: "创建 API Key" }));

    expect(await screen.findByText("acct_secret")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "关闭 API Key" }));
    expect(screen.queryByText("acct_secret")).not.toBeInTheDocument();
  });
});
