import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { RevealSecret } from "./RevealSecret";

describe("RevealSecret", () => {
  it("hides secret until reveal is clicked", async () => {
    render(<RevealSecret label="Access token" value="secret-token" />);

    expect(screen.queryByText("secret-token")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "查看 Access token" }));
    expect(screen.getByText("secret-token")).toBeInTheDocument();
  });
});
