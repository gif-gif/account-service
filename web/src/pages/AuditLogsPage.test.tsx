import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuditLogsPage } from "./AuditLogsPage";

describe("AuditLogsPage", () => {
  beforeEach(() => vi.stubEnv("VITE_API_BASE_URL", "https://api.example.com"));

  it("renders audit logs with request id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(
          JSON.stringify({
            audit_logs: [
              {
                id: "audit-id",
                actor_type: "admin",
                action: "account.update",
                request_id: "request-id",
                metadata: { password: "[REDACTED]" },
              },
            ],
          }),
          { status: 200 },
        ),
      ),
    );

    render(<AuditLogsPage />);

    expect(await screen.findByText("account.update")).toBeInTheDocument();
    expect(screen.getByText("request-id")).toBeInTheDocument();
    expect(screen.getByText("[REDACTED]")).toBeInTheDocument();
  });
});
