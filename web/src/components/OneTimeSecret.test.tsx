import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { OneTimeSecret } from "./OneTimeSecret";

describe("OneTimeSecret", () => {
  it("shows secret once and hides after dismissal", async () => {
    render(<OneTimeSecret label="API key" value="acct_secret" />);

    expect(screen.getByText("acct_secret")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "关闭 API key" }));
    expect(screen.queryByText("acct_secret")).not.toBeInTheDocument();
  });
});
